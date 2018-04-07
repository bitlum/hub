package emulation

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
	"time"
	"strconv"
)

// RouterEmulation is an implementation of router. Router interface which
// completely detached from real lightning network daemon and emulates it
// activity.
type RouterEmulation struct {
	freeBalance    router.BalanceUnit
	pendingBalance router.BalanceUnit
	network        *emulationNetwork
	fee            uint64
}

// Runtime check that RouterEmulation implements router.Router interface.
var _ router.Router = (*RouterEmulation)(nil)

// NewRouter creates new entity of emulator router and start grpc server which l
func NewRouter(freeBalance router.BalanceUnit,
	blockGeneration time.Duration) *RouterEmulation {

	n := newEmulationNetwork(blockGeneration)
	r := &RouterEmulation{
		freeBalance: freeBalance,
		network:     n,
	}
	n.router = r

	return r
}

// Done is used to notify external subsystems that emulator router stopped
// working.
func (r *RouterEmulation) Done() chan error {
	return r.network.errChan
}

// Stop...
func (r *RouterEmulation) Start(host, port string) {
	r.network.start(host, port)
}

// Stop...
func (r *RouterEmulation) Stop() {
	r.network.stop()
}

// SendPayment makes the payment on behalf of router. In the context of
// lightning network hub manager this hook might be used for future
// off-chain channel re-balancing tactics.
func (r *RouterEmulation) SendPayment(userID router.UserID,
	amount router.BalanceUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	// TODO(andrew.shvv) Implement for rebalancing
	return errors.Errorf("not implemented")
}

// OpenChannel opens the channel with the given user.
func (r *RouterEmulation) OpenChannel(userID router.UserID,
	funds router.BalanceUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	r.network.channelIndex++
	id := strconv.FormatUint(r.network.channelIndex, 10)
	chanID := router.ChannelID(id)

	if _, ok := r.network.users[userID]; ok {
		// TODO(andrew.shvv) add multiple channels support
		return errors.Errorf("multiple channels unsupported")
	}

	c := &router.Channel{
		ChannelID:     chanID,
		UserID:        userID,
		UserBalance:   0,
		RouterBalance: funds,
		IsPending:     true,
	}

	r.network.users[userID] = c
	r.network.channels[chanID] = c

	r.network.broadcaster.Write(&router.UpdateChannelOpening{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   router.BalanceUnit(c.UserBalance),
		RouterBalance: router.BalanceUnit(c.RouterBalance),

		// TODO(andrew.shvv) Add work with fee
		Fee: 0,
	})

	log.Tracef("Router opened channel(%v) with user(%v)", chanID, userID)

	// Subscribe on block notification and update channel when block is
	// generated.
	l := r.network.blockNotifier.Listen()

	// Channel is able to operate only after block is generated.
	// Send update that channel is opened only after it is unlocked.
	go func() {
		defer l.Stop()
		<-l.Read()

		r.network.Lock()
		defer r.network.Unlock()

		c.IsPending = false
		r.network.broadcaster.Write(&router.UpdateChannelOpened{
			UserID:        c.UserID,
			ChannelID:     c.ChannelID,
			UserBalance:   router.BalanceUnit(c.UserBalance),
			RouterBalance: router.BalanceUnit(c.RouterBalance),

			// TODO(andrew.shvv) Add work with fee
			Fee: 0,
		})

		log.Tracef("Channel(%v) with user(%v) unlocked", chanID, userID)
	}()

	return nil
}

// CloseChannel closes the specified channel.
func (r *RouterEmulation) CloseChannel(id router.ChannelID) error {
	r.network.Lock()
	defer r.network.Unlock()

	if channel, ok := r.network.channels[id]; !ok {
		return errors.Errorf("unable to find channel with %v id: %v", id)
	} else if channel.IsPending {
		return errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	// TODO(andrew.shvv) add multiple channels support
	for userID, channel := range r.network.users {
		if channel.ChannelID == id {

			r.pendingBalance += channel.RouterBalance

			// Lock the channel and send the closing notification.
			// Wait for block to be generated and only after that remove it
			// from router network.
			channel.IsPending = true
			r.network.broadcaster.Write(&router.UpdateChannelClosing{
				UserID:    userID,
				ChannelID: id,

				// TODO(andrew.shvv) Add work with fee
				Fee: 0,
			})

			log.Tracef("Router closed channel(%v)", id)

			// Subscribe on block notification and return funds when block is
			// generated.
			l := r.network.blockNotifier.Listen()

			// Update router free balance only after block is mined and increase
			// router balance on amount which we locked on our side in this channel.
			go func() {
				defer l.Stop()
				<-l.Read()

				r.network.Lock()
				defer r.network.Unlock()

				delete(r.network.users, userID)
				delete(r.network.channels, id)

				r.pendingBalance -= channel.RouterBalance
				r.freeBalance += channel.RouterBalance

				r.network.broadcaster.Write(&router.UpdateChannelClosed{
					UserID:    userID,
					ChannelID: id,

					// TODO(andrew.shvv) Add work with fee
					Fee: 0,
				})

				log.Tracef("Router received %v money previously locked in"+
					" channel(%v)", channel.RouterBalance, id)
			}()

			break
		}
	}

	return nil
}

// UpdateChannel updates the number of locked funds in the specified
// channel.
func (r *RouterEmulation) UpdateChannel(id router.ChannelID,
	newRouterBalance router.BalanceUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	channel, ok := r.network.channels[id]
	if !ok {
		return errors.Errorf("unable to find the channel with %v id", id)
	} else if channel.IsPending {
		return errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	if newRouterBalance < 0 {
		return errors.New("new balance is lower than zero")
	}

	diff := newRouterBalance - channel.RouterBalance

	if diff > 0 {
		// Number of funds we want to add from our free balance to the
		// channel on router side.
		sliceInFunds := diff

		if sliceInFunds > r.freeBalance {
			return errors.Errorf("insufficient free funds")
		}

		r.freeBalance -= sliceInFunds
		r.pendingBalance += sliceInFunds
	} else {
		// Number of funds we want to get from our channel to the
		// channel on free balance.
		sliceOutFunds := -diff

		// Redundant check, left here just for security if input values would
		if sliceOutFunds > channel.RouterBalance {
			return errors.Errorf("insufficient funds in channel")
		}

		channel.RouterBalance -= sliceOutFunds
		r.pendingBalance += sliceOutFunds
	}

	// During channel update make it locked, so that it couldn't be used by
	// both sides.
	channel.IsPending = true
	r.network.broadcaster.Write(&router.UpdateChannelUpdating{
		UserID:        channel.UserID,
		ChannelID:     channel.ChannelID,
		UserBalance:   channel.UserBalance,
		RouterBalance: channel.RouterBalance,

		// TODO(andrew.shvv) Add work with fee
		Fee: 0,
	})

	// Subscribe on block notification and return funds when block is
	// generated.
	l := r.network.blockNotifier.Listen()

	// Update router free balance only after block is mined and increase
	// router balance on amount which we locked on our side in this channel.
	go func() {
		defer l.Stop()
		<-l.Read()

		r.network.Lock()
		defer r.network.Unlock()

		if diff > 0 {
			// Number of funds we want to add from our pending balance to the
			// channel on router side.
			sliceInFunds := diff

			r.pendingBalance -= sliceInFunds
			channel.RouterBalance += sliceInFunds
		} else {
			// Number of funds we want to get from our pending channel
			// balance to the free balance.
			sliceOutFunds := -diff

			r.pendingBalance -= sliceOutFunds
			r.freeBalance += sliceOutFunds
		}

		log.Tracef("Update channel(%v) balance, old(%v) => new(%v)",
			channel.RouterBalance, newRouterBalance)

		channel.IsPending = false
		r.network.broadcaster.Write(&router.UpdateChannelUpdated{
			UserID:        channel.UserID,
			ChannelID:     channel.ChannelID,
			UserBalance:   channel.UserBalance,
			RouterBalance: channel.RouterBalance,

			// TODO(andrew.shvv) Add work with fee
			Fee: 0,
		})
	}()

	return nil

}

// RegisterOnUpdates returns updates about router local network topology
// changes, about attempts of propagating the payment through the
// router, about fee changes etc.
func (r *RouterEmulation) RegisterOnUpdates() *router.Receiver {
	r.network.Lock()
	defer r.network.Unlock()

	return r.network.broadcaster.Listen()
}

// Network returns the information about the current local network router
// topology.
func (r *RouterEmulation) Network() ([]*router.Channel, error) {
	r.network.Lock()
	defer r.network.Unlock()

	var channels []*router.Channel
	for _, channel := range r.network.channels {
		channels = append(channels, &router.Channel{
			IsPending:     channel.IsPending,
			ChannelID:     channel.ChannelID,
			UserID:        channel.UserID,
			UserBalance:   channel.UserBalance,
			RouterBalance: channel.RouterBalance,
		})
	}

	return channels, nil
}

// FreeBalance returns the amount of funds at router disposal.
func (r *RouterEmulation) FreeBalance() (router.BalanceUnit, error) {
	r.network.Lock()
	defer r.network.Unlock()

	return r.freeBalance, nil
}

// PendingBalance returns the amount of funds which in the process of
// being accepted by blockchain.
func (r *RouterEmulation) PendingBalance() (router.BalanceUnit, error) {
	r.network.Lock()
	defer r.network.Unlock()

	return r.pendingBalance, nil
}

// AverageChangeUpdateDuration average time which is needed the change of
// state to ba updated over blockchain.
func (r *RouterEmulation) AverageChangeUpdateDuration() (time.Duration, error) {
	r.network.Lock()
	defer r.network.Unlock()

	return r.network.blockGeneration, nil
}

func (r *RouterEmulation) SetFee(fee uint64) error {
	r.network.Lock()
	defer r.network.Unlock()

	r.fee = fee
	return nil
}
