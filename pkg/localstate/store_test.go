package localstate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptedStateLockAndTamper(t *testing.T) {
	root := t.TempDir()
	passphrase := []byte("correct horse battery staple")
	state, data, err := Open(root, "https://example.dev", "user-a", passphrase)
	if err != nil || data != nil {
		t.Fatal("new state did not open", err)
	}
	if err = state.Save([]byte(`{"message":"old"}`)); err != nil {
		t.Fatal(err)
	}
	if err = state.Save([]byte(`{"message":"secret text"}`)); err != nil {
		t.Fatal(err)
	}
	if _, _, err = Open(root, "https://example.dev", "user-a", passphrase); err == nil {
		t.Fatal("two processes opened one ratchet state")
	}
	if err = state.Close(); err != nil {
		t.Fatal(err)
	}
	if _, _, err = Open(root, "https://example.dev", "user-a", []byte("wrong passphrase here")); err == nil {
		t.Fatal("wrong passphrase opened state")
	}
	state, data, err = Open(root, "https://example.dev", "user-a", passphrase)
	if err != nil || string(data) != `{"message":"secret text"}` {
		t.Fatal("state did not round-trip", err)
	}
	if err = state.Close(); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(root, "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatal("state file not found", err)
	}
	onDisk, err := os.ReadFile(files[0])
	if err != nil || bytes.Contains(onDisk, []byte("secret text")) {
		t.Fatal("plaintext leaked to disk", err)
	}
	var record envelope
	if err = json.Unmarshal(onDisk, &record); err != nil {
		t.Fatal(err)
	}
	record.Ciphertext[len(record.Ciphertext)-1] ^= 1
	onDisk, _ = json.Marshal(record)
	if err = os.WriteFile(files[0], onDisk, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err = Open(root, "https://example.dev", "user-a", passphrase); err == nil {
		t.Fatal("tampered state opened")
	}
}

func TestMaximumStateCanBeReopened(t *testing.T) {
	root := t.TempDir()
	pass := []byte("maximum state passphrase")
	state, _, err := Open(root, "https://example.dev", "large", pass)
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Repeat([]byte("a"), maxPlaintext)
	if err = state.Save(want); err != nil {
		t.Fatal(err)
	}
	if err = state.Close(); err != nil {
		t.Fatal(err)
	}
	state, got, err := Open(root, "https://example.dev", "large", pass)
	if err != nil {
		t.Fatal("maximum saved state could not be reopened", err)
	}
	defer state.Close()
	if !bytes.Equal(got, want) {
		t.Fatal("maximum state changed")
	}
}
