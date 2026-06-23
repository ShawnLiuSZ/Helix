package agent

import (
	"testing"
	"time"
)

func TestMessageBus_SendReceive(t *testing.T) {
	bus := NewMessageBus()
	ch := bus.Subscribe("agent1")

	bus.Send(BusMessage{
		FromID:  "agent0",
		ToID:    "agent1",
		Content: "hello",
		Type:    "text",
	})

	select {
	case msg := <-ch:
		if msg.Content != "hello" {
			t.Errorf("Content = %q, want %q", msg.Content, "hello")
		}
		if msg.FromID != "agent0" {
			t.Errorf("FromID = %q, want %q", msg.FromID, "agent0")
		}
		if msg.ToID != "agent1" {
			t.Errorf("ToID = %q, want %q", msg.ToID, "agent1")
		}
		if msg.Timestamp.IsZero() {
			t.Error("Timestamp should be set automatically")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestMessageBus_Broadcast(t *testing.T) {
	bus := NewMessageBus()
	ch1 := bus.Subscribe("a")
	ch2 := bus.Subscribe("b")

	bus.Send(BusMessage{
		FromID:  "sender",
		Content: "broadcast",
	})

	for _, ch := range []<-chan BusMessage{ch1, ch2} {
		select {
		case msg := <-ch:
			if msg.Content != "broadcast" {
				t.Errorf("Content = %q, want %q", msg.Content, "broadcast")
			}
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for broadcast")
		}
	}
}

func TestMessageBus_SubscribeUnsubscribe(t *testing.T) {
	bus := NewMessageBus()

	ch := bus.Subscribe("agent1")
	subs := bus.ListSubscribers()
	if len(subs) != 1 || subs[0] != "agent1" {
		t.Errorf("subscribers = %v, want [agent1]", subs)
	}

	bus.Unsubscribe("agent1")

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("channel should be closed after unsubscribe")
		}
	case <-time.After(time.Second):
		t.Fatal("channel not closed after unsubscribe")
	}

	subs = bus.ListSubscribers()
	if len(subs) != 0 {
		t.Errorf("subscribers = %v, want empty", subs)
	}
}
