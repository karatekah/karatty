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
type SignupStart struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	PublicKey string `json:"public_key"`
}
type EmailProof struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Signature []byte `json:"signature"`
}
type SignupResult struct {
	Session      Session `json:"session"`
	RecoveryCode string  `json:"recovery_code"`
}
type RecoveryStart struct {
	Email     string `json:"email"`
	PublicKey string `json:"public_key"`
}
type RecoveryProof struct {
	EmailProof
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
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
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
