package worker

import (
	"context"
	"encoding/json"
	"sync"
)

type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[chan []byte]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: map[chan []byte]struct{}{},
	}
}

func (b *Broadcaster) Subscribe(ctx context.Context) <-chan []byte {
	ch := make(chan []byte, 8)

	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subscribers, ch)
		close(ch)
		b.mu.Unlock()
	}()

	return ch
}

func (b *Broadcaster) Publish(payload any) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}
	b.PublishRaw(raw)
}

func (b *Broadcaster) PublishRaw(raw []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- raw:
		default:
		}
	}
}
