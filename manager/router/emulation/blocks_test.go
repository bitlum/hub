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

	s := notifier.Subscribe()
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

	s := notifier.Subscribe()
	defer notifier.RemoveSubscription(s)

	select {
	case <-s.C:
	case <-time.After(2 * delay):
		t.Fatalf("haven't received notification")
	}
}
