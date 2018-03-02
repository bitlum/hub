package emulation

import "github.com/bitlum/hub/manager/router"

type routerEmulation struct {
	network     *emulationNetwork
	freeBalance uint64
}

// Runtime check that routerEmulation implements router.Router interface.
var _ router.Router = (*routerEmulation)(nil)

func NewRouterEmulator(freeBalance uint64) *routerEmulation {
	r := &routerEmulation{
		freeBalance: freeBalance,
	}

	r.network = &emulationNetwork{
		channels: make(map[router.ChannelID]*router.Channel),
		users:    make(map[router.UserID]*router.Channel),
		updates:  make(chan interface{}),
		router:   r,
	}

	return r
}

// SendPayment makes the payment on behalf of router. In the context of
// lightning network hub manager this hook might be used for future
// off-chain channel re-balancing tactics.
func (e *routerEmulation) SendPayment(userID router.UserID,
	amount uint64) error {
	return nil
}

// OpenChannel opens the channel with the given user.
func (e *routerEmulation) OpenChannel(id router.UserID,
	funds router.LockedFunds) error {
	// TODO(andrew.shvv) implement it
	return nil
}

// CloseChannel closes the specified channel.
func (e *routerEmulation) CloseChannel(id router.ChannelID) error {
	// TODO(andrew.shvv) implement it
	return nil
}

// UpdateChannel updates the number of locked funds in the specified
// channel.
func (e *routerEmulation) UpdateChannel(id router.UserID,
	funds router.LockedFunds) error {
	// TODO(andrew.shvv) implement it
	return nil
}

// ReceiveUpdates returns updates about router local network topology
// changes, about attempts of propagating the payment through the
// router, about fee changes etc.
func (e *routerEmulation) ReceiveUpdates() <-chan interface{} {
	// TODO(andrew.shvv) implement it
	return nil
}
