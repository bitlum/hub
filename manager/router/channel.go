package router

import (
	"time"
	"github.com/bitlum/hub/manager/common/broadcast"
	"github.com/go-errors/errors"
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

type ChannelConfig struct {
	// Broadcast is used to by the channel to send channel notifications updates
	// to it, usually it is populated by the router broadcaster.
	Broadcaster *broadcast.Broadcaster

	// Storage is used by channel to keep important data persistent.
	Storage ChannelStorage
}

func (c *ChannelConfig) validate() error {
	if c.Broadcaster == nil {
		return errors.Errorf("broadcaster is empty")
	}

	if c.Storage == nil {
		return errors.Errorf("storage is empty")
	}

	return nil
}

// Channel represent the Lightning Network channel.
type Channel struct {
	ChannelID ChannelID
	UserID    UserID

	UserBalance   BalanceUnit
	RouterBalance BalanceUnit

	// Initiator side which initiated open of the channel.
	Initiator ChannelInitiator

	// CloseFee is the number of funds which are needed to close the channel
	// and release locked funds, might change with time, because of the
	// commitment transaction size and fee rate in the network.
	CloseFee BalanceUnit

	// CloseFee is the number of funds which were needed to open the channel
	// and lock funds.
	OpenFee BalanceUnit

	// States...
	States []*ChannelState

	cfg *ChannelConfig
}

func NewChannel(channelID ChannelID, userID UserID, openFee, userBalance,
routerBalance, closeFee BalanceUnit, initiator ChannelInitiator,
	cfg *ChannelConfig) (*Channel, error) {

	c := &Channel{
		ChannelID:     channelID,
		UserID:        userID,
		OpenFee:       openFee,
		UserBalance:   userBalance,
		RouterBalance: routerBalance,
		Initiator:     initiator,
		CloseFee:      closeFee,
	}

	if err := c.SetConfig(cfg); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Channel) Save() error {
	return c.cfg.Storage.AddChannel(c)
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

func (c *Channel) SetOpeningState() error {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelOpening,
		Status: ChannelNonActive,
	})

	err := c.cfg.Storage.AddChannelState(c.ChannelID, c.CurrentState())
	if err != nil {
		return errors.Errorf("unable save channel: %v", err)
	}

	c.cfg.Broadcaster.Write(&UpdateChannelOpening{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   c.UserBalance,
		RouterBalance: c.RouterBalance,
		Fee:           c.FundingFee(),
	})

	return nil
}

func (c *Channel) SetOpenedState() error {
	lastStateTime := c.CurrentState().Time

	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelOpened,
		Status: ChannelActive,
	})

	err := c.cfg.Storage.AddChannelState(c.ChannelID, c.CurrentState())
	if err != nil {
		return errors.Errorf("unable to save channel state: %v", err)
	}

	c.cfg.Broadcaster.Write(&UpdateChannelOpened{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   c.UserBalance,
		RouterBalance: c.RouterBalance,
		Fee:           c.FundingFee(),
		Duration:      time.Now().Sub(lastStateTime),
	})

	return nil
}

func (c *Channel) SetUpdatingState(fee BalanceUnit) error {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelUpdating,
		Status: ChannelNonActive,
	})

	err := c.cfg.Storage.AddChannelState(c.ChannelID, c.CurrentState())
	if err != nil {
		return errors.Errorf("unable save channel: %v", err)
	}

	c.cfg.Broadcaster.Write(&UpdateChannelUpdating{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   c.UserBalance,
		RouterBalance: c.RouterBalance,
		Fee:           fee,
	})

	return nil
}

func (c *Channel) SetUpdatedState(fee BalanceUnit) error {
	lastStateTime := c.CurrentState().Time

	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelOpened,
		Status: ChannelActive,
	})

	err := c.cfg.Storage.AddChannelState(c.ChannelID, c.CurrentState())
	if err != nil {
		return errors.Errorf("unable to save channel state: %v", err)
	}

	c.cfg.Broadcaster.Write(&UpdateChannelUpdated{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   c.UserBalance,
		RouterBalance: c.RouterBalance,
		Fee:           fee,
		Duration:      time.Now().Sub(lastStateTime),
	})

	return nil
}

func (c *Channel) SetClosingState() error {
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelClosing,
		Status: ChannelNonActive,
	})

	err := c.cfg.Storage.AddChannelState(c.ChannelID, c.CurrentState())
	if err != nil {
		return errors.Errorf("unable save channel: %v", err)
	}

	c.cfg.Broadcaster.Write(&UpdateChannelClosing{
		UserID:    c.UserID,
		ChannelID: c.ChannelID,
		Fee:       c.CloseFee,
	})

	return nil
}

func (c *Channel) SetClosedState() error {
	lastStateTime := c.CurrentState().Time
	c.States = append(c.States, &ChannelState{
		Time:   time.Now(),
		Name:   ChannelClosed,
		Status: ChannelNonActive,
	})

	// NOTE: If remove this, be aware of double notification
	if err := c.cfg.Storage.RemoveChannel(c); err != nil {
		return errors.Errorf("unable remove channel: %v", err)
	}

	c.cfg.Broadcaster.Write(&UpdateChannelClosed{
		UserID:    c.UserID,
		ChannelID: c.ChannelID,
		Fee:       c.CloseFee,
		Duration:  time.Now().Sub(lastStateTime),
	})

	return nil
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
		return c.OpenFee
	}

	return 0
}

func (c *Channel) SetConfig(cfg *ChannelConfig) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	c.cfg = &(*cfg)
	return nil
}
