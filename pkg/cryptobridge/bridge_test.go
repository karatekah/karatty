package cryptobridge

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
)

func call(t *testing.T, bin string, request any) map[string]json.RawMessage {
	t.Helper()
	data, err := Call(context.Background(), bin, request)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]json.RawMessage
	if err = json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestOlmBridgeRoundTrip(t *testing.T) {
	bin := os.Getenv("KARATTY_CRYPTO_BIN")
	if bin == "" {
		t.Skip("set KARATTY_CRYPTO_BIN to test the Rust bridge")
	}
	alice := call(t, bin, map[string]any{"op": "create", "one_time_keys": 1})
	bob := call(t, bin, map[string]any{"op": "create", "one_time_keys": 1})
	var bobIdentity struct {
		Ed    string `json:"ed25519"`
		Curve string `json:"curve25519"`
	}
	if err := json.Unmarshal(bob["identity"], &bobIdentity); err != nil {
		t.Fatal(err)
	}
	var keys []struct{ Key, Signature string }
	if err := json.Unmarshal(bob["one_time_keys"], &keys); err != nil || len(keys) != 1 {
		t.Fatal("bad one-time key", err)
	}
	ed, _ := base64.RawStdEncoding.DecodeString(bobIdentity.Ed)
	sig, _ := base64.RawStdEncoding.DecodeString(keys[0].Signature)
	if !ed25519.Verify(ed, []byte("KARATTY-OTK-v1\n"+bobIdentity.Curve+"\n"+keys[0].Key), sig) {
		t.Fatal("Rust one-time key signature failed Go verification")
	}
	var aliceIdentity struct {
		Curve string `json:"curve25519"`
	}
	if err := json.Unmarshal(alice["identity"], &aliceIdentity); err != nil {
		t.Fatal(err)
	}
	first := call(t, bin, map[string]any{
		"op": "encrypt", "account": alice["account"], "session": nil,
		"peer_identity": bobIdentity.Curve, "peer_one_time_key": keys[0].Key, "plaintext": "hola",
	})
	received := call(t, bin, map[string]any{
		"op": "decrypt", "account": bob["account"], "session": nil,
		"sender_identity": aliceIdentity.Curve, "message": first["message"],
	})
	var plaintext string
	if err := json.Unmarshal(received["plaintext"], &plaintext); err != nil || plaintext != "hola" {
		t.Fatal("cross-language round trip failed", err)
	}
}
