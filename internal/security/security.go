package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
