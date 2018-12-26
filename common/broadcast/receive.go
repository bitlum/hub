package broadcast

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// Receiver is an entity which listens for broadcast update, receives in
// non-blocking way and re-sends the update to the end user, when he/she will
// be ready to receive them.
type Receiver struct {
	stopped     int32
	broadcaster *Broadcaster

	// outgoing channel for sending data to the user
	outgoing chan interface{}

	// incoming channel for the receiving data from the broadcaster
	incoming chan interface{}

	quit chan struct{}
	wg   sync.WaitGroup

	// pendingBroadcasts is a list which is need so that broadcaster could be
	// non-blocking. Every update which broadcaster sends goes into this
	// list, and only after that broadcasts are sent to the user.
	pendingBroadcasts *list.List
}

func newReceiver(b *Broadcaster) *Receiver {
	return &Receiver{
		broadcaster: b,

		outgoing: make(chan interface{}),
		incoming: make(chan interface{}),

		quit:              make(chan struct{}),
		pendingBroadcasts: list.New(),
	}
}

// start is used by the broadcaster only to kick off the listening process.
func (l *Receiver) start() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()

		var infiniteDuration = time.Hour * 999999

		// There is no any data initially, for that reason we shouldn't
		// fall thorough in the processing select case.
		var awaitDuration = infiniteDuration

		for {
			select {
			case <-l.quit:
				return

			case v := <-l.incoming:
				// Put new data in the data list
				l.pendingBroadcasts.PushBack(v)

				// Notify receiver loop that the new data has arrived
				awaitDuration = 0

			case <-time.After(awaitDuration):
				var next *list.Element

			iteration:
				for e := l.pendingBroadcasts.Front(); e != nil; e = next {
					// Fetch new data and try to send it to the user
					select {
					case <-l.quit:
						return
					case l.outgoing <- e.Value:
						// If e is removed from the list then call of e.Next()
						// in the next loop will return nil. Therefore, need
						// to assign e.Next() to the next before deleting e.
						next = e.Next()
						l.pendingBroadcasts.Remove(e)
						awaitDuration = 0
					case <-time.After(time.Millisecond * 10):
						// If user is not waiting for the data than we should
						// skip loop to preserve the order.
						break iteration
					}
				}

				if l.pendingBroadcasts.Len() == 0 {
					awaitDuration = infiniteDuration
				} else {
					// If not all data was sent to the user,
					// than try again after some time. This is done in order
					// to not hammer processor.
					if awaitDuration < time.Millisecond*200 {
						awaitDuration += time.Millisecond * 20
					}
				}
			}
		}
	}()
}

// write is used by the broadcaster only to send the new data to the receiver
func (l *Receiver) write(value interface{}) {
	// Avoid "send on close" panic
	select {
	case l.incoming <- value:
	case <-l.quit:
		return
	}
}

// Stop unsubscribe from the broadcaster updates, notify user that receiver
// was closed by closing the outgoing channel, and exit the processing
// goroutine.
func (l *Receiver) Stop() {
	if !atomic.CompareAndSwapInt32(&l.stopped, 0, 1) {
		return
	}

	// Signal to the broadcaster that we are no longer interested in the
	// receiving updates.
	l.broadcaster.unsubscribe(l)

	// Signal to the main loop to exit
	close(l.quit)
	l.wg.Wait()

	// Notifying end user that receiver was closed
	close(l.incoming)
	close(l.outgoing)
}

// Read is used to read the data which were broadcast.
func (l *Receiver) Read() <-chan interface{} {
	return l.outgoing
}
