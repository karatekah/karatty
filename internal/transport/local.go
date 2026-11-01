package transport

import (
	"context"
	"time"

	"github.com/karatekah/termlink/internal/chat"
	"github.com/karatekah/termlink/pkg/client"
	"github.com/karatekah/termlink/pkg/protocol"
)

// Local is an authenticated adapter, never a parallel messaging implementation.
type Local struct {
	Chat  *chat.Service
	Token string
}

func (l Local) List(ctx context.Context) ([]protocol.Conversation, error) {
	return l.Chat.List(ctx, l.Token)
}
func (l Local) History(ctx context.Context, c string, a int64, n int) (protocol.Page, error) {
	return l.Chat.History(ctx, l.Token, c, a, n)
}
func (l Local) Send(ctx context.Context, c string, r protocol.Send) (protocol.Message, error) {
	return l.Chat.Send(ctx, l.Token, c, r)
}
func (l Local) Direct(ctx context.Context, u string) (protocol.Conversation, error) {
	return l.Chat.Direct(ctx, l.Token, u)
}
func (l Local) DirectRequests(ctx context.Context) ([]protocol.DirectRequest, error) {
	return l.Chat.DirectRequests(ctx, l.Token)
}
func (l Local) DeclineDirect(ctx context.Context, u string) error {
	return l.Chat.DeclineDirect(ctx, l.Token, u)
}
func (l Local) BlockDirect(ctx context.Context, u string) error {
	return l.Chat.BlockDirect(ctx, l.Token, u)
}
func (l Local) UnblockDirect(ctx context.Context, u string) error {
	return l.Chat.UnblockDirect(ctx, l.Token, u)
}
func (l Local) Stream(ctx context.Context, conv string, after int64) <-chan client.Event {
	out := make(chan client.Event, 1)
	go func() {
		defer close(out)
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				query, cancel := context.WithTimeout(ctx, 5*time.Second)
				p, err := l.History(query, conv, after, 100)
				cancel()
				if err != nil {
					select {
					case out <- client.Event{Err: err, Fatal: true}:
					case <-ctx.Done():
					}
					return
				}
				if len(p.Messages) > 0 {
					select {
					case out <- client.Event{Page: p}:
						after = p.Next
					case <-ctx.Done():
						return
					case <-time.After(5 * time.Second):
						return
					}
				}
			}
		}
	}()
	return out
}
