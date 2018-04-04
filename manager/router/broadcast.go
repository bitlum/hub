package router

// Broadcast is an element which contains the link on next
// broadcast element. This structure implements the broadcast queue.
// This structure contains the writen in the broadcaster value, and has a link
// on the next channel, which will be used for pushing the next broadcast object.
// This allows receivers to wait for the next element.
type Broadcast struct {
	next  chan Broadcast;
	value interface{};
}

// Broadcaster is an implementation of one to many object broadcasting
// without blocking of one of the readers/listeners/receivers haven't able to
// read the message.
//
// NOTE: Such implementation might potentially be the source of memory leak,
// if one the receivers unable to read the messages.
type Broadcaster struct {
	commands chan *cmdInitReceiver
	sendChan chan interface{}
}

// NewBroadcaster creates a new broadcaster object.
func NewBroadcaster() Broadcaster {
	brodcaster := Broadcaster{
		commands: make(chan *cmdInitReceiver),
		sendChan: make(chan interface{}),
	}

	go func() {
		// Broadcast channel is used as an last element in the linked list,
		// it will be update during each write. This channel is used as a
		// storage of broadcast object, it contains zero elements if it last,
		// and one element if not last.
		lastElem := make(chan Broadcast, 1)

		for {
			select {
			case v := <-brodcaster.sendChan:
				if v == nil {
					lastElem <- Broadcast{}
					return
				}

				// Create new broadcast element an push it to the "list" by
				newElem := make(chan Broadcast, 1)
				lastElem <- Broadcast{next: newElem, value: v}
				lastElem = newElem

			case cmd := <-brodcaster.commands:
				// Init receiver with the last element in the queue.
				cmd.resp <- Receiver{
					broadcasts: lastElem,
				}
			}
		}
	}()

	return brodcaster
}

// cmdInitReceiver is used to initialise the new receiver in thread safe manner.
type cmdInitReceiver struct {
	resp chan Receiver
}

// Listen start listening to the broadcasts.
func (b Broadcaster) Listen() Receiver {
	cmd := &cmdInitReceiver{
		resp: make(chan Receiver, 0),
	}

	b.commands <- cmd
	return <-cmd.resp
}

// Write broadcast a value to all listeners/receivers.
func (b Broadcaster) Write(v interface{}) { b.sendChan <- v }

// Receiver is an entity which is receiving messages,
// broadcast by the broadcaster.
type Receiver struct {
	broadcasts chan Broadcast
}

// Read a value that has been broadcast, waiting until one is available if
// necessary.
//
// NOTE: Value of the broadcast element might be accessed only by one of the
// receivers, so this operation is thread-safe.
func (r *Receiver) Read() chan interface{} {
	c := make(chan interface{}, 1)

	go func() {
		// Take the broadcast element for the channel, such action services as the
		// mutex, because other receivers on this stage are waiting for the element
		// to be in channel.
		b := <-r.broadcasts
		v := b.value

		// Put element again in the channel, so that other receiver could take it,
		// if they want.
		r.broadcasts <- b

		// Update channel with the next, if this channel contains the broadcast
		// element than it will be taken on the next read.
		r.broadcasts = b.next

		c <- v
	}()

	return c
}
