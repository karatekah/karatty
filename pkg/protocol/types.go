// Package protocol defines the versioned wire contract shared by both clients.
package protocol

import "time"

type ChallengeRequest struct {
	PublicKey  string `json:"public_key"`
	Username   string `json:"username,omitempty"`
	Invitation string `json:"invitation,omitempty"`
}
type Challenge struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
}
type RecoveryStart struct {
	Username  string `json:"username"`
	PublicKey string `json:"public_key"`
}
type RecoveryProof struct {
	Proof
	RecoveryCode string `json:"recovery_code"`
}
type RecoveryResult struct {
	RecoveryCode string `json:"recovery_code"`
}
type Proof struct {
	ID        string `json:"id"`
	Signature []byte `json:"signature"`
}
type Session struct {
	Token        string    `json:"token"`
	UserID       string    `json:"user_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	RecoveryCode string    `json:"recovery_code,omitempty"`
	PublicMode   bool      `json:"public_mode"`
}
type Conversation struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Closed  bool   `json:"closed"`
	Pending bool   `json:"pending,omitempty"`
}
type DirectRequest struct {
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
type OneTimeKey struct {
	Key       string `json:"key"`
	Signature string `json:"signature"`
}
type DeviceUpload struct {
	IdentityEd    string       `json:"identity_ed"`
	IdentityCurve string       `json:"identity_curve"`
	SSHSignature  []byte       `json:"ssh_signature"`
	OneTimeKeys   []OneTimeKey `json:"one_time_keys"`
}
type DeviceBundle struct {
	UserID        string      `json:"user_id"`
	Username      string      `json:"username"`
	SSHKey        string      `json:"ssh_key"`
	IdentityEd    string      `json:"identity_ed"`
	IdentityCurve string      `json:"identity_curve"`
	SSHSignature  []byte      `json:"ssh_signature"`
	OneTimeKey    *OneTimeKey `json:"one_time_key,omitempty"`
}
type Send struct {
	ClientID string `json:"client_id"`
	Body     string `json:"body"`
}
type Message struct {
	ID             string    `json:"id"`
	ClientID       string    `json:"client_id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Username       string    `json:"username"`
	Seq            int64     `json:"seq"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}
type Page struct {
	Messages []Message `json:"messages"`
	Next     int64     `json:"next"`
}
type Failure struct {
	Error string `json:"error"`
}
