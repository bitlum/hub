package emulation

import (
	"sync"
	"time"

	"errors"
	"github.com/bitlum/hub/manager/router"
)

// blockNotifier is used to mock the block generation notifications,
// it is needed to properly emulate lightning network.
type blockNotifier struct {
	sync.Mutex
	quit chan struct{}

	router.Broadcaster

	blockTicker *time.Ticker
	commands    chan interface{}
}

func newBlockNotifier(blockGenDuration time.Duration) *blockNotifier {
	return &blockNotifier{
		Broadcaster: router.NewBroadcaster(),
		blockTicker: time.NewTicker(blockGenDuration),
		quit:        make(chan struct{}),
		commands:    make(chan interface{}, 1),
	}
}

// StartNotifying start goroutine which generates the block notifications and resend
// them to all subscribers.
//
// NOTE: Should run as goroutine.
func (n *blockNotifier) Start() {
	for {
		select {
		case <-n.blockTicker.C:
			n.MineBlock()
		case c := <-n.commands:
			switch cmd := c.(type) {
			case *setBlockGenCmd:
				n.blockTicker.Stop()
				n.blockTicker = time.NewTicker(cmd.t)
			}
		case <-n.quit:
			return
		}
	}
}

// Stop...
func (n *blockNotifier) Stop() {
	close(n.quit)
}

// MineBlock is used to trigger the block generation notification.
func (n *blockNotifier) MineBlock() {
	log.Tracef("Block generated/mined")
	n.Write(struct{}{})
}

type setBlockGenCmd struct {
	t time.Duration
}

// SetBlockGenDuration is used to set the new block generation duration time,
// to make it faster or slower.
func (n *blockNotifier) SetBlockGenDuration(duration time.Duration) error {
	select {
	case <-n.quit:
		return errors.New("block notifier has stopped")
	case n.commands <- &setBlockGenCmd{
		t: duration,
	}:
	}

	return nil
}
