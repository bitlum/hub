package router

import (
	"testing"
	"fmt"
	"strconv"
	"time"
)

func TestBroadcaster(t *testing.T) {
	broadcaster := NewBroadcaster()

	firstReceiver := broadcaster.Listen()
	secondReceiver := broadcaster.Listen()

	go func() {
		for i := 1; i < 10001; i++ {
			broadcaster.Write(strconv.Itoa(i))
		}
	}()

	go func() {
		for {
			fmt.Println("first:", <-firstReceiver.Read())
		}
	}()

	go func() {
		for {
			fmt.Println("second:", <-secondReceiver.Read())
		}
	}()

	time.Sleep(time.Second * 10)
}
