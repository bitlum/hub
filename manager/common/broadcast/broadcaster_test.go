package broadcast

import (
	"testing"
	"time"
	"strconv"
)

func TestBroadcaster_Broadcast(t *testing.T) {
	b := NewBroadcaster()
	defer b.Stop()

	l := b.Subscribe()
	defer l.Stop()

	b.Write("kek")

	select {
	case s := <-l.Read():
		if s != "kek" {
			t.Fatalf("wrong data")
		}
	case <-time.After(time.Second):
		t.Fatalf("data not received")
	}
}

func TestBroadcaster_Order(t *testing.T) {
	b := NewBroadcaster()
	defer b.Stop()

	r := b.Subscribe()
	defer r.Stop()

	n := 100000
	go func() {
		for i := 0; i < n; i++ {
			b.Write(strconv.Itoa(i))
		}
	}()

	for i := 0; i < n; i++ {
		select {
		case s := <-r.Read():
			expected := strconv.Itoa(i)
			if s != expected {
				t.Fatalf("wrong data: %v, expected: %v", s, expected)
			}
		case <-time.After(time.Second):
			t.Fatalf("data not received")
		}
	}
}

func TestBroadcaster_SecondListener(t *testing.T) {
	b := NewBroadcaster()
	defer b.Stop()

	r1 := b.Subscribe()
	defer r1.Stop()

	b.Write("kek1")

	r2 := b.Subscribe()
	defer r2.Stop()

	b.Write("kek2")

	select {
	case s := <-r1.Read():
		if s != "kek1" {
			t.Fatalf("wrong data")
		}
	case <-time.After(time.Second):
		t.Fatalf("data not received")
	}

	select {
	case s := <-r2.Read():
		if s != "kek2" {
			t.Fatalf("wrong data")
		}
	case <-time.After(time.Second):
		t.Fatalf("data not received")
	}

	select {
	case s := <-r1.Read():
		if s != "kek2" {
			t.Fatalf("wrong data")
		}
	case <-time.After(time.Second):
		t.Fatalf("data not received")
	}

	select {
	case s := <-r2.Read():
		t.Fatalf("unexpected data: %v", s)
	case <-time.After(time.Millisecond * 100):
	}
}

func TestBroadcaster_SendAfterStop(t *testing.T) {
	b := NewBroadcaster()
	defer b.Stop()

	b.Write("kek")

	r := b.Subscribe()
	r.Stop()

	for i := 0; i < 10000; i++ {
		b.Write("kek")
	}
}

func TestBroadcaster_ReadAfterListenerStop(t *testing.T) {
	b := NewBroadcaster()
	defer b.Stop()

	b.Write("kek")

	r := b.Subscribe()
	r.Stop()

	select {
	case s := <-r.Read():
		if s != nil {
			t.Fatalf("unexpected data: %v", s)
		}
	}
}

func TestBroadcaster_BroadcasterDoubleStop(t *testing.T) {
	b := NewBroadcaster()
	b.Stop()
	b.Stop()
}

func TestBroadcaster_ReceiverDoubleStop(t *testing.T) {
	b := NewBroadcaster()
	defer b.Stop()

	r := b.Subscribe()
	r.Stop()
	r.Stop()
}

func TestBroadcaster_SubscribeAfterStop(t *testing.T) {
	b := NewBroadcaster()
	b.Stop()

	r := b.Subscribe()
	select {
	case <-r.Read():
	default:
	}
}
