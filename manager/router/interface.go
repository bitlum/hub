package router

// Router aka payment provider, aka hub, aka lightning network node.
// This interface gives as unified way of managing different implementations of
// lightning network daemons. With this interface we could control hub/router
// and force it into the state of equilibrium - the state with maximum
// efficiency. The exact efficiency metrics depends on chosen strategy.
type Router interface {
	// SendPayment makes the payment on behalf of router. In the context of
	// lightning network hub manager this hook might be used for future
	// off-chain channel re-balancing tactics.
	SendPayment(userID UserID, amount uint64) error

	// OpenChannel opens the channel with the given user.
	OpenChannel(id UserID, funds LockedFunds) error

	// CloseChannel closes the specified channel.
	CloseChannel(id ChannelID) error

	// UpdateChannel updates the number of locked funds in the specified
	// channel.
	UpdateChannel(id UserID, funds LockedFunds) error

	// ReceiveUpdates returns updates about router local network topology
	// changes, about attempts of propagating the payment through the
	// router, about fee changes etc.
	ReceiveUpdates() <-chan interface{}
}

// ChannelID uniquely identifies the channel in the lightning network.
type ChannelID uint64

// UserID uniquely identifies the user in the local lightning network.
type UserID uint64

// LockedFunds represent the number of funds locked by the participant.
type LockedFunds uint64

// Channel represent the Lightning Network channel.
type Channel struct {
	ChannelID     ChannelID
	UserID        UserID
	UserBalance   uint64
	RouterBalance uint64
}

type UpdatePayment struct{}
type UpdateChannelClosed struct{}
type UpdateChannelOpened struct{}
type UpdatePaymentUpdated struct{}
