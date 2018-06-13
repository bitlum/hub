package router

import (
	"time"
	"github.com/bitlum/hub/manager/common/broadcast"
)

// Router aka payment provider, aka hub, aka lightning network node.
// This interface gives as unified way of managing different implementations of
// lightning network daemons. With this interface we could control hub/router
// and force it into the state of equilibrium - the state with maximum
// efficiency. The exact efficiency metrics depends on chosen strategy.
type Router interface {
	// SendPayment makes the payment on behalf of router. In the context of
	// lightning network hub manager this hook might be used for future
	// off-chain channel re-balancing tactics.
	SendPayment(userID UserID, amount BalanceUnit) error

	// OpenChannel opens the channel with the given user.
	OpenChannel(id UserID, funds BalanceUnit) error

	// CloseChannel closes the specified channel.
	CloseChannel(id ChannelID) error

	// UpdateChannel updates the number of locked funds in the specified
	// channel.
	UpdateChannel(id ChannelID, funds BalanceUnit) error

	// SetFeeBase sets base number of milli units (i.e milli satoshis in
	// Bitcoin) which will be taken for every forwarding payment.
	SetFeeBase(feeBase int64) error

	// SetFeeProportional sets the number of milli units (i.e milli
	// satoshis in Bitcoin) which will be taken for every killo-unit of
	// forwarding payment amount as a forwarding fee.
	SetFeeProportional(feeProportional int64) error

	// RegisterOnUpdates returns register which returns updates about router
	// local network topology changes, about attempts of propagating the payment
	// through the router, about fee changes etc.
	RegisterOnUpdates() *broadcast.Receiver

	// Network returns the information about the current local network router
	// topology.
	Network() ([]*Channel, error)

	// FreeBalance returns the amount of funds at router disposal.
	FreeBalance() (BalanceUnit, error)

	// PendingBalance returns the amount of funds which in the process of
	// being accepted by blockchain.
	PendingBalance() (BalanceUnit, error)

	// AverageChangeUpdateDuration average time which is needed the change of
	// state to ba updated over blockchain.
	AverageChangeUpdateDuration() (time.Duration, error)

	// Done returns error if router stopped working for some reason,
	// and nil if it was stopped.
	Done() chan error

	// Asset returns asset with which corresponds to this router.
	Asset() string
}

// ChannelID uniquely identifies the channel in the lightning network.
type ChannelID string

// UserID uniquely identifies the user in the local lightning network.
type UserID string

// BalanceUnit represent the number of funds.
type BalanceUnit int64

// Channel represent the Lightning Network channel.
type Channel struct {
	ChannelID     ChannelID
	UserID        UserID
	UserBalance   BalanceUnit
	RouterBalance BalanceUnit
	State         ChannelState
	Status        ChannelStatus
	Initiator     string
	CloseFee      BalanceUnit
}

func (c *Channel) IsPending() bool {
	// TODO(andrew.shvv) Channel is not pending if it is in update mode,
	// because of the dynamic mode, for more info watch lightning labs conner
	// l2 summit watchtower video.
	return c.State == ChannelOpening ||
		c.State == ChannelClosing ||
		c.State == ChannelUpdating
}

func (c *Channel) IsActive() bool {
	return c.Status == ChannelActive
}

type UpdateChannelClosing struct {
	UserID    UserID
	ChannelID ChannelID

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

type UpdateChannelClosed struct {
	UserID    UserID
	ChannelID ChannelID

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration time.Duration
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

	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
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

	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration time.Duration
}

// UpdateChannelOpening is used as notifications from router or network that
// channel started to opening, and wait for blockchain confirmation.
type UpdateChannelOpening struct {
	UserID        UserID
	ChannelID     ChannelID
	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit
}

// UpdateChannelOpened is used as notifications from router or network that
// channel has been opened.
type UpdateChannelOpened struct {
	UserID        UserID
	ChannelID     ChannelID
	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Fee which was taken by blockchain decentralised computer / mainers or
	// some other form of smart contract manager from initiator of the
	// channel. By initiator we means the side which created the channel.
	Fee BalanceUnit

	// Duration is a period of time which was needed to proceed the channel
	// update by blockchain decentralised computer.
	Duration time.Duration
}

type UpdatePayment struct {
	Status PaymentStatus
	Type   PaymentType

	Sender   UserID
	Receiver UserID

	Amount BalanceUnit

	// Earned is the number of funds which router earned by making this payment.
	// In case of re-balancing router will pay the fee, for that reason this
	// number will be negative.
	Earned BalanceUnit
}

// UpdateLinkAverageUpdateDuration is used when router wants to notify that
// the average link update time has changed.
type UpdateLinkAverageUpdateDuration struct {
	AverageUpdateDuration time.Duration
}

type ChannelState string

const (
	// ChannelOpening denotes that channel open request has been sent in
	// blockchain network, and that we wait for its approval.
	ChannelOpening ChannelState = "opening"

	// ChannelOpened denotes that channel open request was approved in blockchain,
	// and it could be used for routing the payments.
	ChannelOpened ChannelState = "opened"

	// ChannelClosing denotes that channel close request has been sent in
	// blockchain network, and that we wait for its approval.
	// Channel couldn't be used for routing payments,
	// and funds in this channel are still couldn't be used.
	ChannelClosing ChannelState = "closing"

	// ChannelClosed denotes that channel close request was approved in blockchain,
	// and couldn't be used for routing payment anymore, locked on our side
	// funds now are back in wallet.
	ChannelClosed ChannelState = "closed"

	// ChannelUpdating denotes that channel overall capacity is updating,
	// either decreasing or increasing. During this update previous channel
	// should stay in operational mode i.e. being able to route payments.
	ChannelUpdating ChannelState = "updating"
)

// ChannelStatus identifies does the channel could be used for routing the
// payments. Channel is active when we have a tcp connection with remote
// node, and channel in the operational mode.
type ChannelStatus string

const (
	ChannelActive    ChannelStatus = "active"
	ChannelNonActive ChannelStatus = "nonactive"
)

type PaymentStatus string

const (
	Successful PaymentStatus = "successful"

	// InsufficientFunds means that router haven't posses/locked enough funds
	// with receiver user to route through the payment.
	InsufficientFunds PaymentStatus = "insufficient_funds"

	// ExternalFail means that receiver failed to receive payment because of
	// the unknown to us reason.
	ExternalFail PaymentStatus = "external_fail"
)

type PaymentType string

const (
	// Outgoing is the payment which was sent from the router.
	Outgoing PaymentType = "outgoing"

	// Incoming is the payment which was sent from user to router.
	Incoming PaymentType = "incoming"

	// Forward is the payment which was send from uer to user.
	Forward PaymentType = "forward"
)

const (
	// UserInitiator is used when close or update or open was initiated from
	// the user side.
	UserInitiator = "user"

	// RouterInitiator is used when channel close or update or open was
	// initiated by the router side.
	RouterInitiator = "router"
)
