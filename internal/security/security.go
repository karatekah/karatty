package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	ErrUnauthorized = errors.New("authentication required")
	ErrForbidden    = errors.New("conversation access denied")
	ErrInvalid      = errors.New("invalid input")
	ErrConflict     = errors.New("identifier already used with different content")
	ErrLimited      = errors.New("rate limit exceeded")
	identifier      = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,64}$`)
)

func ID() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func Hash(s string) string     { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func Identifier(s string) bool { return identifier.MatchString(s) }
func Body(s string) bool {
	if len(s) == 0 || len(s) > 4096 || !utf8.ValidString(s) || strings.TrimSpace(s) == "" {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return false
		}
	}
	return true
}

// Envelope validates the public-mode wire shape, not cryptographic correctness.
func Envelope(s string) bool {
	if len(s) == 0 || len(s) > 8192 {
		return false
	}
	var in struct {
		Version        int    `json:"version"`
		SessionID      string `json:"session_id"`
		SenderIdentity string `json:"sender_identity"`
		Message        struct {
			Type int    `json:"type"`
			Body string `json:"body"`
		} `json:"message"`
	}
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&in) != nil || decoder.Decode(&struct{}{}) != io.EOF || in.Version != 1 || (in.Message.Type != 0 && in.Message.Type != 1) {
		return false
	}
	for _, field := range []string{in.SessionID, in.SenderIdentity} {
		decoded, err := base64.RawStdEncoding.DecodeString(field)
		if err != nil || len(decoded) != 32 || base64.RawStdEncoding.EncodeToString(decoded) != field {
			return false
		}
	}
	decoded, err := base64.RawStdEncoding.DecodeString(in.Message.Body)
	return err == nil && len(decoded) >= 16 && len(decoded) <= 6144 && base64.RawStdEncoding.EncodeToString(decoded) == in.Message.Body
}

// Display is defense in depth for untrusted text from any server or legacy row.
func Display(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return '�'
		}
		return r
	}, s)
}

type bucket struct {
	at     time.Time
	tokens float64
}

// Limiter has a bounded map; exhausted capacity rejects new identities.
type Limiter struct {
	mu          sync.Mutex
	entries     map[string]bucket
	Rate, Burst float64
	Capacity    int
}

func NewLimiter(rate, burst float64) *Limiter {
	return &Limiter{entries: map[string]bucket{}, Rate: rate, Burst: burst, Capacity: 4096}
}
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b, ok := l.entries[key]
	if !ok {
		if len(l.entries) >= l.Capacity {
			for k, v := range l.entries {
				if now.Sub(v.at) > time.Minute*5 {
					delete(l.entries, k)
				}
			}
			if len(l.entries) >= l.Capacity {
				return false
			}
		}
		b = bucket{at: now, tokens: l.Burst}
	}
	b.tokens = min(l.Burst, b.tokens+now.Sub(b.at).Seconds()*l.Rate)
	b.at = now
	allowed := b.tokens >= 1
	if allowed {
		b.tokens--
	}
	l.entries[key] = b
	return allowed
}
