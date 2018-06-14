package router

import (
	"time"
)

// ChannelID uniquely identifies the channel in the lightning network.
type ChannelID string

// UserID uniquely identifies the user in the local lightning network.
type UserID string

// BalanceUnit represent the number of funds.
type BalanceUnit int64

// ChannelState
type ChannelState struct {
	Time time.Time
	Name ChannelStateName

	// Status determines is this channel is active for being used for routing
	// the payments.
	Status ChannelStatus
}

type ChannelStateName string

const (
	// ChannelOpening denotes that channel open request has been sent in
	// blockchain network, and that we wait for its approval.
	ChannelOpening ChannelStateName = "opening"

	// ChannelOpened denotes that channel open request was approved in blockchain,
	// and it could be used for routing the payments.
	ChannelOpened ChannelStateName = "opened"

	// ChannelClosing denotes that channel close request has been sent in
	// blockchain network, and that we wait for its approval.
	// Channel couldn't be used for routing payments,
	// and funds in this channel are still couldn't be used.
	ChannelClosing ChannelStateName = "closing"

	// ChannelClosed denotes that channel close request was approved in blockchain,
	// and couldn't be used for routing payment anymore, locked on our side
	// funds now are back in wallet.
	ChannelClosed ChannelStateName = "closed"

	// ChannelUpdating denotes that channel overall capacity is updating,
	// either decreasing or increasing. During this update previous channel
	// should stay in operational mode i.e. being able to route payments.
	ChannelUpdating ChannelStateName = "updating"
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

type ChannelInitiator string

const (
	// UserInitiator is used when close or update or open was initiated from
	// the user side.
	UserInitiator ChannelInitiator = "user"

	// RouterInitiator is used when channel close or update or open was
	// initiated by the router side.
	RouterInitiator ChannelInitiator = "router"
)

// Channel represent the Lightning Network channel.
type Channel struct {
	ChannelID ChannelID
	UserID    UserID

	FundingAmount BalanceUnit
	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Initiator side which initiated open of the channel.
	Initiator ChannelInitiator

	// CloseFee is the number of funds which are needed to close the channel
	// and release locked funds, might change with time, because of the
	// commitment transaction size and fee rate in the network.
	CloseFee BalanceUnit

	// States...
	States []*ChannelState
}

func NewChannel(channelID ChannelID, userID UserID, fundingAmount, userBalance,
routerBalance, closeFee BalanceUnit, initiator ChannelInitiator) *Channel {
	return &Channel{
		ChannelID:     channelID,
		UserID:        userID,
		FundingAmount: fundingAmount,
		UserBalance:   userBalance,
		RouterBalance: routerBalance,
		Initiator:     initiator,
		CloseFee:      closeFee,
	}
}

func (c *Channel) CurrentState() *ChannelState {
	return c.States[len(c.States)-1]
}

func (c *Channel) PrevState() *ChannelState {
	if len(c.States) >= 2 {
		return c.States[len(c.States)-2]
	} else {
		return nil
	}
}

func (c *Channel) SetOpeningState() {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelOpening,
		Status: ChannelNonActive,
	})
}

func (c *Channel) SetOpenedState() {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelOpened,
		Status: ChannelActive,
	})
}

func (c *Channel) SetUpdatingState() {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelUpdating,
		Status: ChannelNonActive,
	})
}

func (c *Channel) SetClosingState() {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelClosing,
		Status: ChannelNonActive,
	})
}

func (c *Channel) SetClosedState() {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelClosed,
		Status: ChannelNonActive,
	})
}

func (c *Channel) IsPending() bool {
	// TODO(andrew.shvv) Channel is not pending if it is in update mode,
	// because of the dynamic mode, for more info watch lightning labs conner
	// l2 summit watchtower video.
	currentState := c.CurrentState()
	return currentState.Name == ChannelOpening ||
		currentState.Name == ChannelClosing ||
		currentState.Name == ChannelUpdating
}

func (c *Channel) IsActive() bool {
	return c.CurrentState().Status == ChannelActive
}

func (c *Channel) FundingFee() BalanceUnit {
	if c.Initiator == RouterInitiator {
		return c.FundingAmount - (c.RouterBalance + c.UserBalance)
	}

	return 0
}
