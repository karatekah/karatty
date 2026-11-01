package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/karatekah/termlink/pkg/protocol"
	"golang.org/x/crypto/ssh"
)

type Event struct {
	Page  protocol.Page
	Err   error
	Fatal bool
}
type Backend interface {
	List(context.Context) ([]protocol.Conversation, error)
	History(context.Context, string, int64, int) (protocol.Page, error)
	Send(context.Context, string, protocol.Send) (protocol.Message, error)
	Direct(context.Context, string) (protocol.Conversation, error)
	Stream(context.Context, string, int64) <-chan Event
}
type HTTP struct {
	Base      string
	Token     string
	Transport *http.Client
}
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }
func terminal(err error) bool {
	var e *Error
	return errors.As(err, &e) && (e.Status == 401 || e.Status == 403 || e.Status == 400)
}
func New(base string, insecure bool) (*HTTP, error) {
	u, err := url.Parse(base)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
		return nil, errors.New("use an https://host:port server URL")
	}
	if u.Scheme != "https" {
		ip := net.ParseIP(u.Hostname())
		if !insecure || u.Scheme != "http" || ip == nil || !ip.IsLoopback() {
			return nil, errors.New("TLS required; development HTTP requires an explicit loopback IP")
		}
	}
	return &HTTP{Base: strings.TrimRight(base, "/"), Transport: &http.Client{Timeout: 10 * time.Second}}, nil
}
func MessageID() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func (c *HTTP) request(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.Base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	resp, err := c.Transport.Do(req)
	if err != nil {
		return errors.New("server unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var f protocol.Failure
		_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&f)
		if f.Error == "" {
			f.Error = "request failed"
		}
		return &Error{Status: resp.StatusCode, Message: f.Error}
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1048576)).Decode(out)
}
func (c *HTTP) Login(ctx context.Context, signer ssh.Signer, username, invitation string) (protocol.Session, error) {
	var challenge protocol.Challenge
	err := c.request(ctx, "POST", "/v1/auth/challenge", protocol.ChallengeRequest{PublicKey: string(ssh.MarshalAuthorizedKey(signer.PublicKey())), Username: username, Invitation: invitation}, &challenge)
	if err != nil {
		return protocol.Session{}, err
	}
	if !strings.HasPrefix(challenge.Payload, "TERMLINK-AUTH-v1\n") || len(challenge.Payload) > 256 {
		return protocol.Session{}, errors.New("invalid server challenge")
	}
	sig, err := signer.Sign(rand.Reader, []byte(challenge.Payload))
	if err != nil {
		return protocol.Session{}, err
	}
	var session protocol.Session
	err = c.request(ctx, "POST", "/v1/auth/proof", protocol.Proof{ID: challenge.ID, Signature: ssh.Marshal(sig)}, &session)
	if err == nil {
		c.Token = session.Token
	}
	return session, err
}
func (c *HTTP) Logout(ctx context.Context) error {
	return c.request(ctx, "DELETE", "/v1/session", nil, nil)
}
func (c *HTTP) DeleteAccount(ctx context.Context) error {
	if err := c.request(ctx, "DELETE", "/v1/account", nil, nil); err != nil {
		return err
	}
	c.Token = ""
	return nil
}
func (c *HTTP) Metrics(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.Base+"/metrics", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.Transport.Do(req)
	if err != nil {
		return "", errors.New("server unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		var f protocol.Failure
		_ = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&f)
		if f.Error == "" {
			f.Error = "request failed"
		}
		return "", &Error{Status: resp.StatusCode, Message: f.Error}
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return string(data), err
}
func (c *HTTP) List(ctx context.Context) ([]protocol.Conversation, error) {
	var out []protocol.Conversation
	err := c.request(ctx, "GET", "/v1/conversations", nil, &out)
	return out, err
}
func (c *HTTP) History(ctx context.Context, conv string, after int64, limit int) (protocol.Page, error) {
	var p protocol.Page
	err := c.request(ctx, "GET", "/v1/conversations/"+url.PathEscape(conv)+"/messages?after="+strconv.FormatInt(after, 10)+"&limit="+strconv.Itoa(limit), nil, &p)
	return p, err
}
func (c *HTTP) Send(ctx context.Context, conv string, in protocol.Send) (protocol.Message, error) {
	var m protocol.Message
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = c.request(ctx, "POST", "/v1/conversations/"+url.PathEscape(conv)+"/messages", in, &m)
		if err == nil {
			return m, nil
		}
		var e *Error
		if errors.As(err, &e) && e.Status < 500 {
			return m, err
		}
		if !pause(ctx, time.Duration(attempt+1)*250*time.Millisecond) {
			break
		}
	}
	return m, err
}
func (c *HTTP) Direct(ctx context.Context, username string) (protocol.Conversation, error) {
	var out protocol.Conversation
	err := c.request(ctx, "POST", "/v1/direct", map[string]string{"username": username}, &out)
	return out, err
}
func pause(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
func publish(ctx context.Context, ch chan Event, event Event) bool {
	select {
	case ch <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

// Stream reconnects from the last delivered sequence; messages committed while
// disconnected are read from PostgreSQL. Backpressure is bounded to one page.
func (c *HTTP) Stream(ctx context.Context, conv string, after int64) <-chan Event {
	out := make(chan Event, 1)
	go func() {
		defer close(out)
		delay := 250 * time.Millisecond
		for ctx.Err() == nil {
			endpoint := strings.Replace(c.Base, "https://", "wss://", 1)
			endpoint = strings.Replace(endpoint, "http://", "ws://", 1)
			dialCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			conn, resp, err := websocket.Dial(dialCtx, endpoint+"/v1/conversations/"+url.PathEscape(conv)+"/events?after="+strconv.FormatInt(after, 10), &websocket.DialOptions{HTTPClient: c.Transport, HTTPHeader: http.Header{"Authorization": []string{"Bearer " + c.Token}}})
			cancel()
			if err != nil {
				if resp != nil && resp.Body != nil {
					resp.Body.Close()
				}
				fatal := resp != nil && (resp.StatusCode == 401 || resp.StatusCode == 403 || resp.StatusCode == 400)
				if !publish(ctx, out, Event{Err: errors.New("connection interrupted; synchronizing"), Fatal: fatal}) || fatal {
					return
				}
			} else {
				conn.SetReadLimit(1048576)
				delay = 250 * time.Millisecond
				for ctx.Err() == nil {
					var page protocol.Page
					err = wsjson.Read(ctx, conn, &page)
					if err != nil {
						break
					}
					filtered := protocol.Page{Messages: []protocol.Message{}, Next: after}
					for _, m := range page.Messages {
						if m.Seq > filtered.Next {
							filtered.Messages = append(filtered.Messages, m)
							filtered.Next = m.Seq
						}
					}
					if !publish(ctx, out, Event{Page: filtered}) {
						conn.CloseNow()
						return
					}
					after = filtered.Next
				}
				conn.CloseNow()
				if websocket.CloseStatus(err) == websocket.StatusPolicyViolation {
					publish(ctx, out, Event{Err: errors.New("subscription ended; sign in again or check access"), Fatal: true})
					return
				}
				if !publish(ctx, out, Event{Err: fmt.Errorf("connection interrupted; synchronizing after %d", after)}) {
					return
				}
			}
			if !pause(ctx, delay) {
				return
			}
			delay = min(delay*2, 8*time.Second)
		}
	}()
	return out
}
