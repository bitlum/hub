package emulation

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
)

// RouterEmulation is an implementation of router. Router interface which
// completely detached from real lightning network daemon and emulates it
// activity.
type RouterEmulation struct {
	freeBalance router.ChannelUnit
	errChan     chan error
	network     *emulationNetwork
	fee         uint64
}

// Runtime check that RouterEmulation implements router.Router interface.
var _ router.Router = (*RouterEmulation)(nil)

// NewRouter creates new entity of emulator router and start grpc server which l
func NewRouter(freeBalance router.ChannelUnit) *RouterEmulation {
	errChan := make(chan error)

	n := newEmulationNetwork()
	r := &RouterEmulation{
		freeBalance: freeBalance,
		errChan:     errChan,
		network:     n,
	}
	n.router = r

	return r
}

// Done is used to notify external subsystems that emulator router stopped
// working.
func (r *RouterEmulation) Done() chan error {
	return r.errChan
}

// Stop...
func (r *RouterEmulation) Start(host, port string) {
	r.network.start(host, port)
}

// Stop...
func (r *RouterEmulation) Stop() {
	r.network.stop()
	close(r.errChan)
}

// SendPayment makes the payment on behalf of router. In the context of
// lightning network hub manager this hook might be used for future
// off-chain channel re-balancing tactics.
func (r *RouterEmulation) SendPayment(userID router.UserID,
	amount router.ChannelUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	// TODO(andrew.shvv) Implement for rebalancing
	return errors.Errorf("not implemented")
}

// OpenChannel opens the channel with the given user.
func (r *RouterEmulation) OpenChannel(userID router.UserID,
	funds router.ChannelUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	r.network.channelIndex++
	chanID := router.ChannelID(r.network.channelIndex)

	if _, ok := r.network.users[userID]; ok {
		// TODO(andrew.shvv) add multiple channels support
		return errors.Errorf("multiple channels unsupported")
	}

	c := &router.Channel{
		ChannelID:     chanID,
		UserID:        userID,
		UserBalance:   0,
		RouterBalance: funds,
	}

	r.network.users[userID] = c
	r.network.channels[chanID] = c

	r.network.updates <- &router.UpdateChannelOpened{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   router.ChannelUnit(c.UserBalance),
		RouterBalance: router.ChannelUnit(c.RouterBalance),

		// TODO(andrew.shvv) Add work with fee
		Fee: 0,
	}

	return nil
}

// CloseChannel closes the specified channel.
func (r *RouterEmulation) CloseChannel(id router.ChannelID) error {
	r.network.Lock()
	defer r.network.Unlock()

	if _, ok := r.network.channels[id]; !ok {
		return errors.Errorf("unable to find channel with %v id: %v", id)
	}

	// TODO(andrew.shvv) add multiple channels support
	delete(r.network.channels, id)
	for userID, channel := range r.network.users {
		if channel.ChannelID == id {
			delete(r.network.users, userID)
			r.freeBalance += channel.RouterBalance
			return nil
		}
	}

	return nil
}

// UpdateChannel updates the number of locked funds in the specified
// channel.
func (r *RouterEmulation) UpdateChannel(id router.ChannelID,
	funds router.ChannelUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	if funds > r.freeBalance {
		return errors.Errorf("insufficient free funds")
	}

	channel, ok := r.network.channels[id]
	if !ok {
		return errors.Errorf("unable to find the channel with %v id", id)
	}

	channel.RouterBalance += funds
	r.freeBalance -= funds

	return nil
}

// ReceiveUpdates returns updates about router local network topology
// changes, about attempts of propagating the payment through the
// router, about fee changes etc.
func (r *RouterEmulation) ReceiveUpdates() <-chan interface{} {
	r.network.Lock()
	defer r.network.Unlock()

	return r.network.updates
}

// Network returns the information about the current local network router
// topology.
func (r *RouterEmulation) Network() ([]*router.Channel, error) {
	r.network.Lock()
	defer r.network.Unlock()

	var channels []*router.Channel
	for _, channel := range r.network.channels {
		channels = append(channels, &router.Channel{
			ChannelID:     channel.ChannelID,
			UserID:        channel.UserID,
			UserBalance:   channel.UserBalance,
			RouterBalance: channel.RouterBalance,
		})
	}

	return nil, nil
}

// FreeBalance returns the amount of funds at router disposal.
func (r *RouterEmulation) FreeBalance() (router.ChannelUnit, error) {
	r.network.Lock()
	defer r.network.Unlock()

	return r.freeBalance, nil
}

// FreeBalance returns the amount of funds at router disposal.
func (r *RouterEmulation) SetFee(fee uint64) error {
	r.network.Lock()
	defer r.network.Unlock()

	r.fee = fee
	return nil
}
