// Package localstate stores client state encrypted with a local passphrase.
package localstate

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const maxPlaintext = 16 << 20
const maxFile = maxPlaintext + (1 << 20)

type envelope struct {
	Version    int    `json:"version"`
	Salt       []byte `json:"salt"`
	Nonce      []byte `json:"nonce"`
	Ciphertext []byte `json:"ciphertext"`
}

type Store struct {
	mu     sync.Mutex
	path   string
	aad    []byte
	salt   [16]byte
	key    [32]byte
	lock   *os.File
	unlock func() error
}

// Open locks one account and returns its decrypted contents, if present.
// root must be a private application directory; the caller clears passphrase.
func Open(root, server, uid string, passphrase []byte) (*Store, []byte, error) {
	if !filepath.IsAbs(root) || server == "" || uid == "" || len(passphrase) < 12 {
		return nil, nil, errors.New("invalid local state parameters")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, nil, err
	}
	h := sha256.Sum256([]byte(server + "\n" + uid))
	name := "state-" + hex.EncodeToString(h[:])
	lock, err := os.OpenFile(filepath.Join(root, name+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, nil, err
	}
	unlock, err := lockFile(lock)
	if err != nil {
		lock.Close()
		return nil, nil, errors.New("local state already in use")
	}
	s := &Store{path: filepath.Join(root, name+".json"), aad: []byte("KARATTY-LOCAL-STATE-v1\n" + server + "\n" + uid), lock: lock, unlock: unlock}
	var saved envelope
	var data []byte
	state, err := os.Open(s.path)
	if err == nil {
		data, err = io.ReadAll(io.LimitReader(state, maxFile+1))
		if closeErr := state.Close(); err == nil {
			err = closeErr
		}
	}
	if errors.Is(err, os.ErrNotExist) {
		if _, err = rand.Read(s.salt[:]); err == nil {
			s.derive(passphrase)
			return s, nil, nil
		}
	} else if err == nil {
		if len(data) > maxFile || json.Unmarshal(data, &saved) != nil || saved.Version != 1 || len(saved.Salt) != 16 || len(saved.Nonce) != chacha20poly1305.NonceSizeX || len(saved.Ciphertext) > maxPlaintext+chacha20poly1305.Overhead {
			err = errors.New("invalid local state")
		} else {
			copy(s.salt[:], saved.Salt)
			s.derive(passphrase)
			aead, aeadErr := chacha20poly1305.NewX(s.key[:])
			err = aeadErr
			if err == nil {
				data, err = aead.Open(nil, saved.Nonce, saved.Ciphertext, s.aad)
			}
			if err == nil {
				return s, data, nil
			}
			err = errors.New("cannot unlock local state")
		}
	}
	_ = s.Close()
	return nil, nil, err
}

func (s *Store) derive(passphrase []byte) {
	key := argon2.IDKey(passphrase, s.salt[:], 3, 64*1024, 1, 32)
	copy(s.key[:], key)
	clear(key)
}

// Save replaces the encrypted state. The caller updates ratchet and pending
// envelope in one payload before calling Save, and sends only after it succeeds.
func (s *Store) Save(plaintext []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lock == nil || len(plaintext) > maxPlaintext {
		return errors.New("local state unavailable or too large")
	}
	aead, err := chacha20poly1305.NewX(s.key[:])
	if err != nil {
		return err
	}
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	data, err := json.Marshal(envelope{Version: 1, Salt: s.salt[:], Nonce: nonce, Ciphertext: aead.Seal(nil, nonce, plaintext, s.aad)})
	if err != nil {
		return err
	}
	// shortcut: rewrite a bounded file for the alpha; use a transactional store if it grows.
	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".state-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err = tmp.Chmod(0600); err == nil {
		_, err = tmp.Write(data)
	}
	if err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return durableReplace(tmp.Name(), s.path)
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lock == nil {
		return nil
	}
	clear(s.key[:])
	err := s.unlock()
	if closeErr := s.lock.Close(); err == nil {
		err = closeErr
	}
	s.lock = nil
	return err
}
