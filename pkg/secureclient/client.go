// Package secureclient is the experimental local E2EE direct client.
package secureclient

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/karatekah/karatty/internal/security"
	"github.com/karatekah/karatty/pkg/client"
	"github.com/karatekah/karatty/pkg/cryptobridge"
	"github.com/karatekah/karatty/pkg/localstate"
	"github.com/karatekah/karatty/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

type wire struct {
	Version        int             `json:"version"`
	SessionID      string          `json:"session_id"`
	SenderIdentity string          `json:"sender_identity"`
	Message        json.RawMessage `json:"message"`
}
type plaintext struct {
	Version        int    `json:"version"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	ClientID       string `json:"client_id"`
	Body           []byte `json:"body"`
}
type outgoing struct {
	ConversationID string    `json:"conversation_id"`
	ClientID       string    `json:"client_id"`
	Body           string    `json:"body"`
	Envelope       string    `json:"envelope"`
	CreatedAt      time.Time `json:"created_at"`
}
type UncertainMessage struct {
	ConversationID string    `json:"conversation_id"`
	ClientID       string    `json:"client_id"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}
type disk struct {
	Version       int                                   `json:"version"`
	UserID        string                                `json:"user_id"`
	SSHKey        string                                `json:"ssh_key"`
	IdentityEd    string                                `json:"identity_ed"`
	IdentityCurve string                                `json:"identity_curve"`
	Account       json.RawMessage                       `json:"account"`
	PendingKeys   []protocol.OneTimeKey                 `json:"pending_keys"`
	Pins          map[string]protocol.DeviceBundle      `json:"pins"`
	Sessions      map[string]map[string]json.RawMessage `json:"sessions"`
	Active        map[string]string                     `json:"active"`
	Outbox        map[string]outgoing                   `json:"outbox"`
	Archive       map[string][]protocol.Message         `json:"archive"`
	Cursor        map[string]int64                      `json:"cursor"`
	Reverified    map[string]bool                       `json:"reverified"`
	Uncertain     []outgoing                            `json:"uncertain"`
	NeedsBaseline bool                                  `json:"needs_baseline"`
}
type Client struct {
	mu     sync.Mutex
	http   *client.HTTP
	crypto string
	signer ssh.Signer
	store  *localstate.Store
	state  disk
}

func invoke(ctx context.Context, binary string, request any) (map[string]json.RawMessage, error) {
	data, err := cryptobridge.Call(ctx, binary, request)
	if err != nil {
		return nil, err
	}
	var out map[string]json.RawMessage
	if err = json.Unmarshal(data, &out); err != nil {
		return nil, errors.New("invalid local crypto response")
	}
	return out, nil
}
func stringField(data map[string]json.RawMessage, field string) (string, error) {
	var value string
	if err := json.Unmarshal(data[field], &value); err != nil || value == "" {
		return "", errors.New("invalid local crypto response")
	}
	return value, nil
}
func newDisk(uid, sshKey string) disk {
	return disk{Version: 1, UserID: uid, SSHKey: sshKey, Pins: map[string]protocol.DeviceBundle{}, Sessions: map[string]map[string]json.RawMessage{}, Active: map[string]string{}, Outbox: map[string]outgoing{}, Archive: map[string][]protocol.Message{}, Cursor: map[string]int64{}, Reverified: map[string]bool{}, NeedsBaseline: true}
}
func (c *Client) save(next disk) error {
	// shortcut: the alpha rewrites one bounded encrypted file per change.
	cutoff := time.Now().Add(-720 * time.Hour)
	for conv, messages := range next.Archive {
		kept := messages[:0]
		for _, message := range messages {
			if message.CreatedAt.After(cutoff) {
				kept = append(kept, message)
			}
		}
		next.Archive[conv] = kept
	}
	for key, item := range next.Outbox {
		if item.CreatedAt.Before(cutoff) {
			delete(next.Outbox, key)
		}
	}
	keptUncertain := next.Uncertain[:0]
	for _, item := range next.Uncertain {
		if !item.CreatedAt.Before(cutoff) {
			keptUncertain = append(keptUncertain, item)
		}
	}
	next.Uncertain = keptUncertain
	data, err := json.Marshal(next)
	if err == nil {
		err = c.store.Save(data)
	}
	if err == nil {
		c.state = next
	}
	return err
}
func (c *Client) copy() (disk, error) {
	data, err := json.Marshal(c.state)
	if err != nil {
		return disk{}, err
	}
	var next disk
	err = json.Unmarshal(data, &next)
	return next, err
}

// Open unlocks or creates one device. Initial keys are saved before upload.
func Open(ctx context.Context, http *client.HTTP, signer ssh.Signer, uid string, passphrase []byte, root, cryptoBinary string) (*Client, error) {
	if http == nil || signer == nil || !security.Identifier(uid) || cryptoBinary == "" {
		return nil, security.ErrInvalid
	}
	sshKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(signer.PublicKey())))
	keyHash := sha256.Sum256([]byte(sshKey))
	deviceRoot := filepath.Join(root, "device-"+hex.EncodeToString(keyHash[:]))
	store, data, err := localstate.Open(deviceRoot, http.Base, uid, passphrase)
	if err != nil {
		return nil, err
	}
	c := &Client{http: http, crypto: cryptoBinary, signer: signer, store: store}
	createdNew := len(data) == 0
	if createdNew {
		c.state = newDisk(uid, sshKey)
		created, e := invoke(ctx, c.crypto, map[string]any{"op": "create", "one_time_keys": 20})
		if e != nil {
			store.Close()
			return nil, e
		}
		var identity struct{ Ed, Curve string }
		var values map[string]string
		if e = json.Unmarshal(created["identity"], &values); e == nil {
			identity.Ed, identity.Curve = values["ed25519"], values["curve25519"]
		}
		if e == nil {
			e = json.Unmarshal(created["one_time_keys"], &c.state.PendingKeys)
		}
		if e != nil || identity.Ed == "" || identity.Curve == "" || len(c.state.PendingKeys) != 20 {
			store.Close()
			return nil, errors.New("invalid local crypto identity")
		}
		c.state.IdentityEd, c.state.IdentityCurve, c.state.Account = identity.Ed, identity.Curve, created["account"]
	} else if err = json.Unmarshal(data, &c.state); err != nil || c.state.Version != 1 || c.state.UserID != uid || c.state.SSHKey != sshKey || len(c.state.Account) == 0 || c.state.IdentityEd == "" || c.state.IdentityCurve == "" {
		store.Close()
		return nil, errors.New("local E2EE state belongs to another key or is invalid")
	}
	if c.state.Pins == nil || c.state.Sessions == nil || c.state.Active == nil || c.state.Outbox == nil || c.state.Archive == nil || c.state.Cursor == nil {
		store.Close()
		return nil, errors.New("invalid local E2EE state")
	}
	if c.state.Reverified == nil {
		c.state.Reverified = map[string]bool{}
	}
	oldStores, oldStates, err := olderDevices(root, deviceRoot, http.Base, uid, passphrase)
	if err != nil {
		store.Close()
		return nil, err
	}
	defer func() {
		for _, old := range oldStores {
			_ = old.Close()
		}
	}()
	for _, previous := range oldStates {
		for _, messages := range previous.Archive {
			for _, message := range messages {
				appendArchive(&c.state, message)
			}
		}
		for peer, bundle := range previous.Pins {
			if _, found := c.state.Pins[peer]; !found {
				c.state.Pins[peer] = bundle
			}
		}
		for _, item := range previous.Outbox {
			appendUncertain(&c.state, item)
		}
		for _, item := range previous.Uncertain {
			appendUncertain(&c.state, item)
		}
	}
	if err = c.save(c.state); err != nil {
		store.Close()
		return nil, err
	}
	if len(c.state.PendingKeys) > 0 {
		if err = c.upload(ctx); err != nil {
			store.Close()
			return nil, err
		}
	}
	if c.state.NeedsBaseline {
		if err = c.baseline(ctx); err != nil {
			store.Close()
			return nil, err
		}
	}
	for _, old := range oldStores {
		if err = old.Remove(); err != nil {
			store.Close()
			return nil, err
		}
	}
	return c, nil
}

func appendUncertain(state *disk, item outgoing) {
	for _, old := range state.Uncertain {
		if old.ConversationID == item.ConversationID && old.ClientID == item.ClientID {
			return
		}
	}
	state.Uncertain = append(state.Uncertain, item)
}

func olderDevices(root, current, server, uid string, passphrase []byte) ([]*localstate.Store, []disk, error) {
	dirs, err := filepath.Glob(filepath.Join(root, "device-*"))
	if err != nil {
		return nil, nil, err
	}
	var stores []*localstate.Store
	var states []disk
	for _, dir := range dirs {
		if dir == current {
			continue
		}
		info, statErr := os.Lstat(dir)
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			for _, held := range stores {
				_ = held.Close()
			}
			return nil, nil, errors.New("invalid previous local device path")
		}
		if _, err = os.Stat(localstate.FilePath(dir, server, uid)); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			for _, held := range stores {
				_ = held.Close()
			}
			return nil, nil, err
		}
		old, data, openErr := localstate.Open(dir, server, uid, passphrase)
		if openErr != nil {
			for _, held := range stores {
				_ = held.Close()
			}
			return nil, nil, errors.New("unlock the previous local device before rotating")
		}
		var state disk
		if json.Unmarshal(data, &state) != nil || state.Version != 1 || state.UserID != uid {
			_ = old.Close()
			for _, held := range stores {
				_ = held.Close()
			}
			return nil, nil, errors.New("invalid previous local device")
		}
		stores = append(stores, old)
		states = append(states, state)
	}
	return stores, states, nil
}

// RemoveLocal deletes this account's files in the configured application root.
func RemoveLocal(root, server, uid string) error {
	if !filepath.IsAbs(root) {
		return security.ErrInvalid
	}
	dirs, err := filepath.Glob(filepath.Join(root, "device-*"))
	if err != nil {
		return err
	}
	for _, dir := range append(dirs, root) {
		info, statErr := os.Lstat(dir)
		if errors.Is(statErr, os.ErrNotExist) {
			continue
		}
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("invalid local device path")
		}
		if err = localstate.RemoveAccount(dir, server, uid); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) UncertainCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.state.Uncertain)
}
func (c *Client) Uncertain() []UncertainMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]UncertainMessage, 0, len(c.state.Uncertain))
	for _, item := range c.state.Uncertain {
		out = append(out, UncertainMessage{ConversationID: item.ConversationID, ClientID: item.ClientID, Body: item.Body, CreatedAt: item.CreatedAt})
	}
	return out
}
func (c *Client) Pending() []UncertainMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]UncertainMessage, 0, len(c.state.Outbox))
	for _, item := range c.state.Outbox {
		if item.CreatedAt.After(time.Now().Add(-720 * time.Hour)) {
			out = append(out, UncertainMessage{ConversationID: item.ConversationID, ClientID: item.ClientID, Body: item.Body, CreatedAt: item.CreatedAt})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out
}
func (c *Client) Close() error { return c.store.Close() }

// A newly created device cannot decrypt messages addressed to an older device.
// Record the existing tail before the user verifies peers and starts receiving.
func (c *Client) baseline(ctx context.Context) error {
	rooms, err := c.http.List(ctx)
	if err != nil {
		return err
	}
	next, err := c.copy()
	if err != nil {
		return err
	}
	for _, room := range rooms {
		if room.Kind != "direct" {
			continue
		}
		var after int64
		for {
			page, pageErr := c.http.History(ctx, room.ID, after, 100)
			if pageErr != nil {
				return pageErr
			}
			if len(page.Messages) == 0 || page.Next <= after {
				break
			}
			after = page.Next
			if len(page.Messages) < 100 {
				break
			}
		}
		next.Cursor[room.ID] = after
	}
	next.NeedsBaseline = false
	return c.save(next)
}
func (c *Client) upload(ctx context.Context) error {
	for len(c.state.PendingKeys) > 0 {
		batch := min(20, len(c.state.PendingKeys))
		in := protocol.DeviceUpload{IdentityEd: c.state.IdentityEd, IdentityCurve: c.state.IdentityCurve, OneTimeKeys: c.state.PendingKeys[:batch]}
		payload := "KARATTY-E2EE-IDENTITY-v1\n" + c.state.UserID + "\n" + in.IdentityEd + "\n" + in.IdentityCurve
		signature, err := c.signer.Sign(rand.Reader, []byte(payload))
		if err != nil {
			return err
		}
		in.SSHSignature = ssh.Marshal(signature)
		if err = c.http.UploadDevice(ctx, in); err != nil {
			return err
		}
		next, err := c.copy()
		if err != nil {
			return err
		}
		next.PendingKeys = next.PendingKeys[batch:]
		if err = c.save(next); err != nil {
			return err
		}
	}
	return nil
}

func verifyBundle(bundle protocol.DeviceBundle) (string, error) {
	key, _, options, rest, err := ssh.ParseAuthorizedKey([]byte(bundle.SSHKey))
	if err != nil || key.Type() != ssh.KeyAlgoED25519 || len(options) != 0 || len(strings.TrimSpace(string(rest))) != 0 || bundle.UserID == "" {
		return "", errors.New("invalid peer identity")
	}
	var sig ssh.Signature
	payload := "KARATTY-E2EE-IDENTITY-v1\n" + bundle.UserID + "\n" + bundle.IdentityEd + "\n" + bundle.IdentityCurve
	if ssh.Unmarshal(bundle.SSHSignature, &sig) != nil || key.Verify([]byte(payload), &sig) != nil {
		return "", errors.New("invalid peer identity signature")
	}
	ed, err := base64.RawStdEncoding.DecodeString(bundle.IdentityEd)
	if err != nil || len(ed) != ed25519.PublicKeySize {
		return "", errors.New("invalid peer Ed25519 key")
	}
	curve, err := base64.RawStdEncoding.DecodeString(bundle.IdentityCurve)
	if err != nil || len(curve) != 32 {
		return "", errors.New("invalid peer Curve25519 key")
	}
	if bundle.OneTimeKey != nil {
		otk := bundle.OneTimeKey
		decoded, e := base64.RawStdEncoding.DecodeString(otk.Key)
		signature, e2 := base64.RawStdEncoding.DecodeString(otk.Signature)
		if e != nil || e2 != nil || len(decoded) != 32 || !ed25519.Verify(ed, []byte("KARATTY-OTK-v1\n"+bundle.IdentityCurve+"\n"+otk.Key), signature) {
			return "", errors.New("invalid peer one-time key")
		}
	}
	h := sha256.Sum256([]byte(bundle.UserID + "\n" + bundle.SSHKey + "\n" + bundle.IdentityEd + "\n" + bundle.IdentityCurve))
	return hex.EncodeToString(h[:]), nil
}
func sameIdentity(a, b protocol.DeviceBundle) bool {
	return a.UserID == b.UserID && a.SSHKey == b.SSHKey && a.IdentityEd == b.IdentityEd && a.IdentityCurve == b.IdentityCurve && bytes.Equal(a.SSHSignature, b.SSHSignature)
}

// SelfFingerprint is the value the user shares through an independent channel.
func (c *Client) SelfFingerprint() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	h := sha256.Sum256([]byte(c.state.UserID + "\n" + c.state.SSHKey + "\n" + c.state.IdentityEd + "\n" + c.state.IdentityCurve))
	return hex.EncodeToString(h[:])
}

// Fingerprint must be compared with the other person outside Karatty.
func (c *Client) Fingerprint(ctx context.Context, username string) (string, error) {
	bundle, err := c.http.Peer(ctx, username)
	if err != nil {
		return "", err
	}
	if bundle.Username != username {
		return "", errors.New("peer username mismatch")
	}
	return verifyBundle(bundle)
}
func (c *Client) Verify(ctx context.Context, username, expected string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	bundle, err := c.http.Peer(ctx, username)
	if err != nil {
		return err
	}
	if bundle.Username != username {
		return errors.New("peer username mismatch")
	}
	fingerprint, err := verifyBundle(bundle)
	if err != nil {
		return err
	}
	if !strings.EqualFold(fingerprint, strings.TrimSpace(expected)) {
		return errors.New("fingerprint does not match")
	}
	next, err := c.copy()
	if err != nil {
		return err
	}
	if old, found := next.Pins[username]; found && !sameIdentity(old, bundle) {
		return errors.New("peer identity changed; use reverify after comparing the new fingerprint")
	}
	next.Pins[username] = bundle
	return c.save(next)
}

// Reverify accepts an independently checked new identity. Old unread messages
// become explicit unavailable placeholders; pending sends must be resolved first.
func (c *Client) Reverify(ctx context.Context, conv, username, expected string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !security.Identifier(conv) || !security.Identifier(username) {
		return security.ErrInvalid
	}
	for _, room := range c.state.Outbox {
		if room.ConversationID == conv {
			return errors.New("pending sends must be resolved before re-verifying")
		}
	}
	bundle, err := c.http.Peer(ctx, username)
	if err != nil {
		return err
	}
	if bundle.Username != username {
		return errors.New("peer username mismatch")
	}
	fingerprint, err := verifyBundle(bundle)
	if err != nil || !strings.EqualFold(fingerprint, strings.TrimSpace(expected)) {
		return errors.New("new fingerprint does not match")
	}
	old, found := c.state.Pins[username]
	if !found || sameIdentity(old, bundle) {
		return errors.New("no changed peer identity to re-verify")
	}
	rooms, err := c.http.List(ctx)
	if err != nil {
		return err
	}
	valid := false
	for _, room := range rooms {
		if room.ID == conv && room.Kind == "direct" && room.Name == username {
			valid = true
			break
		}
	}
	if !valid {
		return errors.New("direct conversation unavailable")
	}
	next, err := c.copy()
	if err != nil {
		return err
	}
	next.Pins[username] = bundle
	delete(next.Sessions, conv)
	delete(next.Active, conv)
	next.Reverified[conv] = true
	return c.save(next)
}

// DiscardPending abandons uncertain outgoing envelopes in one conversation.
// The server may already have accepted them; the user must inspect the chat.
func (c *Client) DiscardPending(conv string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !security.Identifier(conv) {
		return 0, security.ErrInvalid
	}
	next, err := c.copy()
	if err != nil {
		return 0, err
	}
	count := 0
	for key, item := range next.Outbox {
		if item.ConversationID == conv {
			delete(next.Outbox, key)
			count++
		}
	}
	if count > 0 {
		if err = c.save(next); err != nil {
			return 0, err
		}
	}
	return count, nil
}

func (c *Client) pinned(ctx context.Context, username string) (protocol.DeviceBundle, error) {
	current, err := c.http.Peer(ctx, username)
	if err != nil {
		return current, err
	}
	if current.Username != username {
		return current, errors.New("peer username mismatch")
	}
	fingerprint, err := verifyBundle(current)
	if err != nil {
		return current, err
	}
	stored, found := c.state.Pins[username]
	if !found {
		return current, fmt.Errorf("verify peer fingerprint outside Karatty: %s", fingerprint)
	}
	if !sameIdentity(stored, current) {
		return current, errors.New("peer identity changed; sending stopped")
	}
	return current, nil
}

func (c *Client) Send(ctx context.Context, conv, username string, in protocol.Send) (protocol.Message, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var result protocol.Message
	if !security.Identifier(conv) || !security.Identifier(in.ClientID) || !security.Body(in.Body) {
		return result, security.ErrInvalid
	}
	key := conv + "/" + in.ClientID
	pending, found := c.state.Outbox[key]
	if found && (pending.Body != in.Body || pending.ConversationID != conv) {
		return result, security.ErrConflict
	}
	if !found {
		for _, old := range c.state.Archive[conv] {
			if old.SenderID == c.state.UserID && old.ClientID == in.ClientID {
				if old.Body != in.Body {
					return result, security.ErrConflict
				}
				return old, nil
			}
		}
	}
	if found {
		if _, err := c.pinned(ctx, username); err != nil {
			return result, err
		}
	}
	if !found {
		peer, err := c.pinned(ctx, username)
		if err != nil {
			return result, err
		}
		next, err := c.copy()
		if err != nil {
			return result, err
		}
		sid := next.Active[conv]
		var session json.RawMessage
		if sid != "" {
			session = next.Sessions[conv][sid]
		}
		var otk string
		if len(session) == 0 {
			claimed, claimErr := c.http.ClaimPeer(ctx, username)
			if claimErr != nil {
				return result, claimErr
			}
			if _, claimErr = verifyBundle(claimed); claimErr != nil || claimed.Username != username || !sameIdentity(peer, claimed) || claimed.OneTimeKey == nil {
				return result, errors.New("peer key changed or invalid")
			}
			otk = claimed.OneTimeKey.Key
		}
		inner, _ := json.Marshal(plaintext{Version: 1, ConversationID: conv, SenderID: next.UserID, ClientID: in.ClientID, Body: []byte(in.Body)})
		encrypted, err := invoke(ctx, c.crypto, map[string]any{"op": "encrypt", "account": next.Account, "session": session, "peer_identity": peer.IdentityCurve, "peer_one_time_key": otk, "plaintext": string(inner)})
		if err != nil {
			return result, err
		}
		sid, err = stringField(encrypted, "session_id")
		if err != nil {
			return result, err
		}
		if next.Sessions[conv] == nil {
			next.Sessions[conv] = map[string]json.RawMessage{}
		}
		next.Sessions[conv][sid] = encrypted["session"]
		next.Active[conv] = sid
		outer, _ := json.Marshal(wire{Version: 1, SessionID: sid, SenderIdentity: next.IdentityCurve, Message: encrypted["message"]})
		pending = outgoing{ConversationID: conv, ClientID: in.ClientID, Body: in.Body, Envelope: string(outer), CreatedAt: time.Now()}
		next.Outbox[key] = pending
		if err = c.save(next); err != nil {
			return result, err
		}
	}
	result, err := c.http.Send(ctx, conv, protocol.Send{ClientID: in.ClientID, Body: pending.Envelope})
	if err != nil {
		return result, err
	}
	if !validMessage(result, conv) || result.ClientID != in.ClientID || result.SenderID != c.state.UserID || result.Body != pending.Envelope {
		return protocol.Message{}, errors.New("invalid encrypted send acknowledgement")
	}
	next, err := c.copy()
	if err != nil {
		return result, err
	}
	result.Body = in.Body
	appendArchive(&next, result)
	delete(next.Outbox, key)
	if err = c.save(next); err != nil {
		return result, err
	}
	return result, nil
}
func appendArchive(state *disk, message protocol.Message) {
	if now := time.Now(); message.CreatedAt.After(now) {
		message.CreatedAt = now
	}
	for _, old := range state.Archive[message.ConversationID] {
		if old.ID == message.ID {
			return
		}
	}
	state.Archive[message.ConversationID] = append(state.Archive[message.ConversationID], message)
	sort.Slice(state.Archive[message.ConversationID], func(i, j int) bool {
		return state.Archive[message.ConversationID][i].Seq < state.Archive[message.ConversationID][j].Seq
	})
}

func validMessage(message protocol.Message, conv string) bool {
	return message.ConversationID == conv && security.Identifier(message.ID) && security.Identifier(message.ClientID) && security.Identifier(message.SenderID) && message.Seq > 0 && !message.CreatedAt.IsZero()
}

// Sync decrypts new server messages and persists ratchet, archive and cursor
// before returning any plaintext to the caller.
func (c *Client) Sync(ctx context.Context, conv, username string) (protocol.Page, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	page := protocol.Page{Messages: []protocol.Message{}, Next: c.state.Cursor[conv]}
	peer, err := c.pinned(ctx, username)
	if err != nil {
		return page, err
	}
	remote, err := c.http.History(ctx, conv, page.Next, 100)
	if err != nil {
		return page, err
	}
	for _, message := range remote.Messages {
		if !validMessage(message, conv) {
			return page, errors.New("invalid encrypted message metadata")
		}
		if message.Seq <= c.state.Cursor[conv] {
			continue
		}
		next, copyErr := c.copy()
		if copyErr != nil {
			return page, copyErr
		}
		if message.SenderID == next.UserID {
			if pending, ok := next.Outbox[conv+"/"+message.ClientID]; ok {
				if message.Body != pending.Envelope {
					return page, errors.New("pending encrypted envelope mismatch")
				}
				message.Body = pending.Body
				delete(next.Outbox, conv+"/"+message.ClientID)
			} else {
				message.Body = "[mensaje propio anterior no disponible en este equipo]"
				for _, old := range next.Archive[conv] {
					if old.ID == message.ID {
						message.Body = old.Body
						break
					}
				}
			}
		} else {
			if message.SenderID != peer.UserID || !security.Envelope(message.Body) {
				return page, errors.New("untrusted encrypted message")
			}
			var outer wire
			if json.Unmarshal([]byte(message.Body), &outer) != nil {
				return page, errors.New("invalid peer envelope")
			}
			if outer.SenderIdentity != peer.IdentityCurve {
				if !next.Reverified[conv] {
					return page, errors.New("peer identity mismatch")
				}
				message.Body = "[mensaje anterior al cambio de identidad; no disponible]"
				next.Cursor[conv] = message.Seq
				appendArchive(&next, message)
				if err = c.save(next); err != nil {
					return page, err
				}
				page.Messages = append(page.Messages, message)
				page.Next = message.Seq
				continue
			}
			session := next.Sessions[conv][outer.SessionID]
			decrypted, decryptErr := invoke(ctx, c.crypto, map[string]any{"op": "decrypt", "account": next.Account, "session": session, "sender_identity": peer.IdentityCurve, "message": outer.Message})
			if decryptErr != nil {
				return page, decryptErr
			}
			sid, fieldErr := stringField(decrypted, "session_id")
			if fieldErr != nil || sid != outer.SessionID {
				return page, errors.New("Olm session mismatch")
			}
			var encoded string
			if json.Unmarshal(decrypted["plaintext"], &encoded) != nil {
				return page, errors.New("invalid encrypted plaintext")
			}
			var inner plaintext
			if json.Unmarshal([]byte(encoded), &inner) != nil || inner.Version != 1 || inner.ConversationID != conv || inner.SenderID != peer.UserID || inner.ClientID != message.ClientID || !security.Body(string(inner.Body)) {
				return page, errors.New("encrypted metadata mismatch")
			}
			message.Body = string(inner.Body)
			next.Account = decrypted["account"]
			if next.Sessions[conv] == nil {
				next.Sessions[conv] = map[string]json.RawMessage{}
			}
			next.Sessions[conv][sid] = decrypted["session"]
			if next.Active[conv] == "" {
				next.Active[conv] = sid
			}
			if len(session) == 0 {
				replenished, replenishErr := invoke(ctx, c.crypto, map[string]any{"op": "replenish", "account": next.Account, "one_time_keys": 1})
				if replenishErr != nil {
					return page, replenishErr
				}
				next.Account = replenished["account"]
				var keys []protocol.OneTimeKey
				if json.Unmarshal(replenished["one_time_keys"], &keys) != nil || len(keys) != 1 {
					return page, errors.New("invalid replenished key")
				}
				next.PendingKeys = append(next.PendingKeys, keys...)
			}
		}
		next.Cursor[conv] = message.Seq
		appendArchive(&next, message)
		if err = c.save(next); err != nil {
			return page, err
		}
		page.Messages = append(page.Messages, message)
		page.Next = message.Seq
	}
	if len(c.state.PendingKeys) > 0 {
		_ = c.upload(ctx)
	}
	return page, nil
}
func (c *Client) Archive(conv string) []protocol.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	cutoff := time.Now().Add(-720 * time.Hour)
	var out []protocol.Message
	for _, message := range c.state.Archive[conv] {
		if message.CreatedAt.After(cutoff) {
			out = append(out, message)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Seq < out[j].Seq })
	return out
}

func (c *Client) Cursor(conv string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state.Cursor[conv]
}
