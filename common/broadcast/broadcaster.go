package broadcast

import (
	"sync"
	"sync/atomic"
)

// Broadcaster is an entity which send message to several receivers in
// non-blocking way.
type Broadcaster struct {
	stopped int32

	// receivers list of subscribed receivers
	receivers map[*Receiver]struct{}

	// commands is a broadcaster commands stream which is used to avoid mutexes
	commands chan interface{}

	// data is a written pendingBroadcasts stream
	data chan interface{}

	quit chan struct{}
	wg   sync.WaitGroup
}

func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		receivers: make(map[*Receiver]struct{}),
		quit:      make(chan struct{}),
		commands:  make(chan interface{}),
		data:      make(chan interface{}),
	}

	b.wg.Add(1)
	go func() {
		defer func() {
			b.wg.Done()
		}()

		for {
			select {
			case command := <-b.commands:
				switch cmd := command.(type) {
				case subscribeCmd:
					receiver := newReceiver(b)
					b.receivers[receiver] = struct{}{}
					receiver.start()
					cmd.resp <- receiver

				case ubsubscribeCmd:
					delete(b.receivers, cmd.receiver)
				}
			case data := <-b.data:
				// Send data to all receivers, as far as receivers have built-in
				// pendingBroadcasts list, this write will not stuck.
				for l, _ := range b.receivers {
					l.write(data)
				}
			case <-b.quit:
				return
			}
		}
	}()

	return b
}

// Write is used to send data to all receivers.
func (b *Broadcaster) Write(data interface{}) {
	select {
	case b.data <- data:
	case <-b.quit:
		return
	}
}

func (b *Broadcaster) Stop() {
	if !atomic.CompareAndSwapInt32(&b.stopped, 0, 1) {
		return
	}

	// Signal main goroutine to exit
	close(b.quit)
	b.wg.Wait()

	close(b.data)
}

// subscribeCmd is used to make a subscriptions and return the newly created
// receiver.
type subscribeCmd struct {
	resp chan *Receiver
}

type ubsubscribeCmd struct {
	receiver *Receiver
}

func (b *Broadcaster) Subscribe() *Receiver {
	cmd := subscribeCmd{
		resp: make(chan *Receiver, 1),
	}

	select {
	case b.commands <- cmd:
	case <-b.quit:
		return newReceiver(b)
	}

	select {
	case l := <-cmd.resp:
		return l
	case <-b.quit:
		return newReceiver(b)
	}
}

// unsubscribe is used by the receiver only, to unsubscribe of the broadcaster
// when stop function is called.
func (b *Broadcaster) unsubscribe(l *Receiver) {
	cmd := ubsubscribeCmd{
		receiver: l,
	}

	select {
	case b.commands <- cmd:
	case <-b.quit:
		return
	}
}
