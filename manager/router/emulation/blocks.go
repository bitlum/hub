package emulation

import (
	"sync"
	"time"
	"github.com/bitlum/graphql-go/errors"
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
}

func newBlockNotifier(blockGenDuration time.Duration) *blockNotifier {
	return &blockNotifier{
		notifications: make(chan struct{}, 10),
		blockTicker:   time.NewTicker(blockGenDuration),
		subscriptions: make(map[int64]chan struct{}),
		quit:          make(chan struct{}),
	}
}

// Start start goroutine which generates the block notifications and resend
// them to all subscribers.
//
// NOTE: Should run as goroutine.
func (n *blockNotifier) Start() {
	// Func is used to resend notification about block to every
	// subscribed on notification client.
	notifySubscribers := func(ntf struct{}) {
		n.Lock()
		defer n.Unlock()

		for _, s := range n.subscriptions {
			go func(s chan struct{}) {
				select {
				case s <- ntf:
				case <-time.After(time.Second):
				}
			}(s)
		}
	}

	for {
		select {
		case ntf := <-n.notifications:
			notifySubscribers(ntf)
		case <-n.blockTicker.C:
			n.notifications <- struct{}{}
		case <-n.quit:
			return
		}
	}
}

// Stop...
func (n *blockNotifier) Stop() {
	close(n.quit)
}

type Subscription struct {
	C  <-chan struct{}
	id int64
}

// Subscribe subscribe on new block generation notification.
func (n *blockNotifier) Subscribe() *Subscription {
	select {
	case <-n.quit:
		return nil
	default:
	}

	n.Lock()
	defer n.Unlock()

	c := make(chan struct{})

	n.subscriptionIndex++
	n.subscriptions[n.subscriptionIndex] = c
	return &Subscription{
		C:  c,
		id: n.subscriptionIndex,
	}
}

// RemoveSubscription...
func (n *blockNotifier) RemoveSubscription(s *Subscription) {
	select {
	case <-n.quit:
		return
	default:
	}

	n.Lock()
	defer n.Unlock()

	delete(n.subscriptions, s.id)
}

// SetBlockGenDuration is used to set the new block generation duration time,
// to make it faster or slower.
func (n *blockNotifier) SetBlockGenDuration(duration time.Duration) error {
	select {
	case <-n.quit:
		return errors.Errorf("notifier was stopped")
	default:
	}

	n.Lock()
	defer n.Unlock()

	n.blockTicker.Stop()
	n.blockTicker = time.NewTicker(duration)
	return nil
}
