package emulation

import (
	"testing"
	"time"
)

func TestNotification(t *testing.T) {
	delay := 100 * time.Millisecond
	notifier := newBlockNotifier(delay)

	go notifier.Start()
	defer notifier.Stop()

	s, _ := notifier.Subscribe()
	defer notifier.RemoveSubscription(s)

	select {
	case <-s.C:
	case <-time.After(2 * delay):
		t.Fatalf("haven't received notification")
	}
}

func TestSetDuration(t *testing.T) {
	delay := 100 * time.Millisecond
	notifier := newBlockNotifier(delay / 100)

	go notifier.Start()
	defer notifier.Stop()

	if err := notifier.SetBlockGenDuration(delay); err != nil {
		t.Fatalf("unable to set block gen duration: %v", err)
	}

	s, _ := notifier.Subscribe()
	defer notifier.RemoveSubscription(s)

	select {
	case <-s.C:
	case <-time.After(2 * delay):
		t.Fatalf("haven't received notification")
	}
}

func TestMine(t *testing.T) {
	notifier := newBlockNotifier(time.Hour)

	go notifier.Start()
	defer notifier.Stop()

	s, _ := notifier.Subscribe()
	defer notifier.RemoveSubscription(s)

	notifier.MineBlock()

	select {
	case <-s.C:
	case <-time.After(time.Second):
		t.Fatalf("haven't received notification")
	}
}
