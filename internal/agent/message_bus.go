package agent

import (
	"sync"
	"time"
)

type BusMessage struct {
	FromID    string
	ToID      string
	Content   string
	Type      string
	Timestamp time.Time
}

type MessageBus struct {
	mu          sync.RWMutex
	subscribers map[string]chan BusMessage
}

func NewMessageBus() *MessageBus {
	return &MessageBus{
		subscribers: make(map[string]chan BusMessage),
	}
}

func (b *MessageBus) Subscribe(agentID string) <-chan BusMessage {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan BusMessage, 64)
	b.subscribers[agentID] = ch
	return ch
}

func (b *MessageBus) Unsubscribe(agentID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subscribers[agentID]; ok {
		close(ch)
		delete(b.subscribers, agentID)
	}
}

func (b *MessageBus) Send(msg BusMessage) {
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	if msg.ToID != "" {
		if ch, ok := b.subscribers[msg.ToID]; ok {
			select {
			case ch <- msg:
			default:
			}
		}
		return
	}

	for _, ch := range b.subscribers {
		if ch == nil {
			continue
		}
		select {
		case ch <- msg:
		default:
		}
	}
}

func (b *MessageBus) ListSubscribers() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subs := make([]string, 0, len(b.subscribers))
	for id := range b.subscribers {
		subs = append(subs, id)
	}
	return subs
}
