package router

import "time"

// Router aka payment provider, aka hub, aka lightning network node.
// This interface gives as unified way of managing different implementations of
// lightning network daemons. With this interface we could control hub/router
// and force it into the state of equilibrium - the state with maximum
// efficiency. The exact efficiency metrics depends on chosen strategy.
type Router interface {
	// SendPayment makes the payment on behalf of router. In the context of
	// lightning network hub manager this hook might be used for future
	// off-chain channel re-balancing tactics.
	SendPayment(userID UserID, amount ChannelUnit) error

	// OpenChannel opens the channel with the given user.
	OpenChannel(id UserID, funds ChannelUnit) error

	// CloseChannel closes the specified channel.
	CloseChannel(id ChannelID) error

	// UpdateChannel updates the number of locked funds in the specified
	// channel.
	UpdateChannel(id ChannelID, funds ChannelUnit) error

	// SetFee updates the fee which router takes for routing the users
	// payments.
	SetFee(fee uint64) error

	// ReceiveUpdates returns updates about router local network topology
	// changes, about attempts of propagating the payment through the
	// router, about fee changes etc.
	ReceiveUpdates() <-chan interface{}

	// Network returns the information about the current local network router
	// topology.
	Network() ([]*Channel, error)

	// FreeBalance returns the amount of funds at router disposal.
	FreeBalance() (ChannelUnit, error)

	// PendingBalance returns the amount of funds which in the process of
	// being accepted by blockchain.
	PendingBalance() (ChannelUnit, error)

	// AverageChangeUpdateDuration average time which is needed the change of
	// state to ba updated over blockchain.
	AverageChangeUpdateDuration() (time.Duration, error)
}

// ChannelID uniquely identifies the channel in the lightning network.
type ChannelID uint64

// UserID uniquely identifies the user in the local lightning network.
type UserID uint64

// ChannelUnit represent the number of funds locked by the participant.
type ChannelUnit int64

// Channel represent the Lightning Network channel.
type Channel struct {
	ChannelID     ChannelID
	UserID        UserID
	UserBalance   ChannelUnit
	RouterBalance ChannelUnit
	IsPending     bool
}

type UpdateChannelClosing struct {
	UserID    UserID
	ChannelID ChannelID

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee ChannelUnit
}

type UpdateChannelClosed struct {
	UserID    UserID
	ChannelID ChannelID

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee ChannelUnit
}

// UpdateChannelUpdating is used to notify that one of the participants
// decided to splice in or splice out some portion of their money from the
// channel.
//
// NOTE: On 11.03.2018 this is not yet possible in the Bitcoin Lightning
// Network, channel might be either opened or closed.
type UpdateChannelUpdating struct {
	UserID    UserID
	ChannelID ChannelID

	UserBalance   ChannelUnit
	RouterBalance ChannelUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee ChannelUnit
}

// UpdateChannelUpdated is used to notify that one of the participants
// decided to splice in or splice out some portion of their money from the
// channel.
//
// NOTE: On 11.03.2018 this is not yet possible in the Bitcoin Lightning
// Network, channel might be either opened or closed.
type UpdateChannelUpdated struct {
	UserID    UserID
	ChannelID ChannelID

	UserBalance   ChannelUnit
	RouterBalance ChannelUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee ChannelUnit
}

// UpdateChannelOpening is used as notifications from router or network that
// channel started to opening, and wait for blockchain confirmation.
type UpdateChannelOpening struct {
	UserID        UserID
	ChannelID     ChannelID
	UserBalance   ChannelUnit
	RouterBalance ChannelUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee ChannelUnit
}

// UpdateChannelOpened is used as notifications from router or network that
// channel has been opened.
type UpdateChannelOpened struct {
	UserID        UserID
	ChannelID     ChannelID
	UserBalance   ChannelUnit
	RouterBalance ChannelUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee ChannelUnit
}

type UpdatePayment struct {
	Status    string
	Sender    uint64
	Receiver  uint64
	ChannelID uint64
	Amount    uint64

	// Earned is the number of funds which router earned by making this payment.
	// In case of re-balancing router will pay the fee, for that reason this
	// number will be negative.
	Earned int64
}


const (
	Successful = "successful"

	// InsufficientFunds means that router haven't posses/locked enough funds
	// with receiver user to route through the payment.
	InsufficientFunds = "insufficient_funds"

	// ExternalFail means that receiver failed to receive payment because of
	// the unknown to us reason.
	ExternalFail = "external_fail"
)

const (
	// UserInitiator is used when close or update or open was initiated from
	// the user side.
	UserInitiator = "user"

	// RouterInitiator is used when channel close or update or open was
	// initiated by the router side.
	RouterInitiator = "router"
)
