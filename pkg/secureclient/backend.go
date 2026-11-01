package secureclient

import (
	"context"
	"errors"

	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
)

// Backend adapts the encrypted client to the local TUI. It never forwards a
// ciphertext page to the view.
type Backend struct {
	HTTP   *client.HTTP
	Secure *Client
}

var _ client.Backend = (*Backend)(nil)

func (b *Backend) List(ctx context.Context) ([]protocol.Conversation, error) {
	return b.HTTP.List(ctx)
}

func (b *Backend) peer(ctx context.Context, conv string) (string, error) {
	rooms, err := b.HTTP.List(ctx)
	if err != nil {
		return "", err
	}
	for _, room := range rooms {
		if room.ID == conv && room.Kind == "direct" {
			return room.Name, nil
		}
	}
	return "", errors.New("direct conversation unavailable")
}

func (b *Backend) syncAll(ctx context.Context, conv, peer string) error {
	for {
		page, err := b.Secure.Sync(ctx, conv, peer)
		if err != nil {
			return err
		}
		if len(page.Messages) < 100 {
			return nil
		}
	}
}

func (b *Backend) History(ctx context.Context, conv string, after int64, limit int) (protocol.Page, error) {
	page := protocol.Page{Messages: []protocol.Message{}, Next: after}
	if after < 0 || limit < 1 || limit > 100 {
		return page, errors.New("invalid history page")
	}
	peer, err := b.peer(ctx, conv)
	if err != nil {
		return page, err
	}
	if err = b.syncAll(ctx, conv, peer); err != nil {
		return page, err
	}
	for _, message := range b.Secure.Archive(conv) {
		if message.Seq > after {
			page.Messages = append(page.Messages, message)
			page.Next = message.Seq
			if len(page.Messages) == limit {
				break
			}
		}
	}
	return page, nil
}

func (b *Backend) Send(ctx context.Context, conv string, in protocol.Send) (protocol.Message, error) {
	peer, err := b.peer(ctx, conv)
	if err != nil {
		return protocol.Message{}, err
	}
	return b.Secure.Send(ctx, conv, peer, in)
}

func (b *Backend) Direct(ctx context.Context, peer string) (protocol.Conversation, error) {
	return b.HTTP.Direct(ctx, peer)
}
func (b *Backend) DirectRequests(ctx context.Context) ([]protocol.DirectRequest, error) {
	return b.HTTP.DirectRequests(ctx)
}
func (b *Backend) DeclineDirect(ctx context.Context, peer string) error {
	return b.HTTP.DeclineDirect(ctx, peer)
}
func (b *Backend) BlockDirect(ctx context.Context, peer string) error {
	return b.HTTP.BlockDirect(ctx, peer)
}
func (b *Backend) UnblockDirect(ctx context.Context, peer string) error {
	return b.HTTP.UnblockDirect(ctx, peer)
}
func (b *Backend) Fingerprint(ctx context.Context, peer string) (string, error) {
	return b.Secure.Fingerprint(ctx, peer)
}
func (b *Backend) Verify(ctx context.Context, peer, fingerprint string) error {
	return b.Secure.Verify(ctx, peer, fingerprint)
}
func (b *Backend) SelfFingerprint() string { return b.Secure.SelfFingerprint() }

func sendEvent(ctx context.Context, out chan<- client.Event, event client.Event) bool {
	select {
	case out <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

func (b *Backend) Stream(ctx context.Context, conv string, after int64) <-chan client.Event {
	out := make(chan client.Event, 1)
	go func() {
		defer close(out)
		peer, err := b.peer(ctx, conv)
		if err != nil {
			sendEvent(ctx, out, client.Event{Err: err, Fatal: true})
			return
		}
		if err = b.syncAll(ctx, conv, peer); err != nil {
			sendEvent(ctx, out, client.Event{Err: err, Fatal: true})
			return
		}
		page := protocol.Page{Messages: []protocol.Message{}, Next: after}
		for _, message := range b.Secure.Archive(conv) {
			if message.Seq > after {
				page.Messages = append(page.Messages, message)
				page.Next = message.Seq
			}
		}
		if !sendEvent(ctx, out, client.Event{Page: page}) {
			return
		}
		after = page.Next
		for event := range b.HTTP.Stream(ctx, conv, b.Secure.Cursor(conv)) {
			if event.Err != nil {
				if !sendEvent(ctx, out, event) || event.Fatal {
					return
				}
				continue
			}
			if err = b.syncAll(ctx, conv, peer); err != nil {
				sendEvent(ctx, out, client.Event{Err: err, Fatal: true})
				return
			}
			page = protocol.Page{Messages: []protocol.Message{}, Next: after}
			for _, message := range b.Secure.Archive(conv) {
				if message.Seq > after {
					page.Messages = append(page.Messages, message)
					page.Next = message.Seq
				}
			}
			if len(page.Messages) > 0 {
				if !sendEvent(ctx, out, client.Event{Page: page}) {
					return
				}
				after = page.Next
			}
		}
	}()
	return out
}
