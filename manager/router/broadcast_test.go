package router

import (
	"testing"
	"time"
)

func TestBroadcasterOrder(t *testing.T) {
	broadcaster := NewBroadcaster()

	receiver := broadcaster.Listen()
	defer receiver.Stop()

	for i := 0; i < 20000; i++ {
		broadcaster.Write(i)

		v, ok := <-receiver.Read()
		if !ok {
			t.Fatalf("receiver was closed")
		}

		j := v.(int)

		if i != j {
			t.Fatalf("wrong order, expected: %v, received: %v", i, j)
		}
	}
}

func TestBroadcasterDoubleRead(t *testing.T) {
	broadcaster := NewBroadcaster()

	receiver := broadcaster.Listen()
	defer receiver.Stop()

	broadcaster.Write(&struct{}{})

	select {
	case <-receiver.Read():
	case <-time.After(time.Millisecond * 50):
		t.Fatalf("haven't received writen data")
	}

	select {
	case <-receiver.Read():
		t.Fatalf("received unxpected data")
	case <-time.After(time.Millisecond * 50):
	}
}
