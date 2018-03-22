package emulation

import (
	"sync"
	"time"

	"errors"
)

// blockNotifier is used to mock the block generation notifications,
// it is needed to properly emulate lightning network.
type blockNotifier struct {
	sync.Mutex
	quit chan struct{}

	notifications chan struct{}
	subscriptions map[int64]chan struct{}

	blockTicker       *time.Ticker
	subscriptionIndex int64
	commands          chan interface{}
}

func newBlockNotifier(blockGenDuration time.Duration) *blockNotifier {
	return &blockNotifier{
		notifications: make(chan struct{}, 10),
		blockTicker:   time.NewTicker(blockGenDuration),
		subscriptions: make(map[int64]chan struct{}),
		quit:          make(chan struct{}),
		commands:      make(chan interface{}, 1),
	}
}

// StartNotifying start goroutine which generates the block notifications and resend
// them to all subscribers.
//
// NOTE: Should run as goroutine.
func (n *blockNotifier) Start() {
	for {
		select {
		case ntf := <-n.notifications:
			// Resend notification about block to every
			// subscribed on notification client.
			for _, s := range n.subscriptions {
				s <- ntf
			}
		case <-n.blockTicker.C:
			n.MineBlock()
		case c := <-n.commands:
			switch cmd := c.(type) {
			case *subscribeCmd:
				channel := make(chan struct{}, 5)

				n.subscriptionIndex++
				n.subscriptions[n.subscriptionIndex] = channel

				cmd.errChan <- nil
				cmd.sChan <- &Subscription{
					C:  channel,
					id: n.subscriptionIndex,
				}
			case *removeSubscriptionCmd:
				delete(n.subscriptions, cmd.s.id)
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
	n.notifications <- struct{}{}
}

type Subscription struct {
	C  <-chan struct{}
	id int64
}

type subscribeCmd struct {
	sChan   chan *Subscription
	errChan chan error
}

// Subscribe subscribe on new block generation notification.
func (n *blockNotifier) Subscribe() (*Subscription, error) {
	cmd := &subscribeCmd{
		sChan:   make(chan *Subscription, 1),
		errChan: make(chan error, 1),
	}

	select {
	case <-n.quit:
		return nil, errors.New("block notifier has stopped")
	case n.commands <- cmd:
	}

	return <-cmd.sChan, <-cmd.errChan
}

type removeSubscriptionCmd struct {
	s *Subscription
}

// RemoveSubscription...
func (n *blockNotifier) RemoveSubscription(s *Subscription) {
	select {
	case <-n.quit:
		return
	case n.commands <- &removeSubscriptionCmd{
		s: s,
	}:
	}
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
