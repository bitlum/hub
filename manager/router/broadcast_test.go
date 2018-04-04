package router

import (
	"testing"
	"fmt"
	"time"
	"strconv"
)

func TestBroadcaster(t *testing.T) {
	broadcaster := NewBroadcaster()

	firstReceiver := broadcaster.Listen()

	go func() {
		for i := 1; i < 101; i++ {
			broadcaster.Write(strconv.Itoa(i))
		}
	}()

	go func() {
		for {
			fmt.Println("first:", <-firstReceiver.Read())
		}
	}()

	time.Sleep(time.Second * 10)
}
