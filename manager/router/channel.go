package router

import (
	"time"
	"github.com/bitlum/hub/manager/common/broadcast"
	"github.com/go-errors/errors"
)

// ChannelID uniquely identifies the channel in the lightning network.
type ChannelID string

// BalanceUnit represent the number of funds.
type BalanceUnit int64

// ChannelState
type ChannelState struct {
	Time int64
	Name ChannelStateName
}

func (s ChannelState) String() string {
	return string(s.Name)
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

// ChannelConfig contains all external replaceable subsystems.
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

	// IsUserConnected is used to determine is user connected with tcp/ip
	// connection to the hub, which means that this channel could be used for
	// payments.
	IsUserConnected bool

	// States is the array of states which this channel went thorough.
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

// Save is used to save the channel in the database, without saving it states.
func (c *Channel) Save() error {
	return c.cfg.Storage.AddChannel(c)
}

// CurrentState return current state of payment channel.
func (c *Channel) CurrentState() *ChannelState {
	return c.States[len(c.States)-1]
}

// SetOpeningState...
func (c *Channel) SetOpeningState() error {
	c.States = append(c.States, &ChannelState{
		Time: time.Now().UnixNano(),
		Name: ChannelOpening,
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

// SetOpenedState...
func (c *Channel) SetOpenedState() error {
	lastStateTime := c.CurrentState().Time

	c.States = append(c.States, &ChannelState{
		Time: time.Now().UnixNano(),
		Name: ChannelOpened,
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
		Duration:      time.Now().UnixNano() - lastStateTime,
	})

	return nil
}

// SetUpdatingState puts channel in the state of being updating,
// depending on static or dynamic channel update it either could be used for
// forwarding payment or couldn't be used. For more information go to
// lightning mailing list.
func (c *Channel) SetUpdatingState(fee BalanceUnit) error {
	c.States = append(c.States, &ChannelState{
		Time: time.Now().UnixNano(),
		Name: ChannelUpdating,
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

// SetUpdatedState put channel in the state of being updated,
// which means that it again open and eligible for forwarding payments.
func (c *Channel) SetUpdatedState(fee BalanceUnit) error {
	lastStateTime := c.CurrentState().Time

	c.States = append(c.States, &ChannelState{
		Time: time.Now().UnixNano(),
		Name: ChannelOpened,
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
		Duration:      time.Now().UnixNano() - lastStateTime,
	})

	return nil
}

// SetClosingState...
func (c *Channel) SetClosingState() error {
	c.States = append(c.States, &ChannelState{
		Time: time.Now().UnixNano(),
		Name: ChannelClosing,
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

// SetClosedState...
func (c *Channel) SetClosedState() error {
	lastStateTime := c.CurrentState().Time
	c.States = append(c.States, &ChannelState{
		Time: time.Now().UnixNano(),
		Name: ChannelClosed,
	})

	// NOTE: If remove this, be aware of double notification
	if err := c.cfg.Storage.RemoveChannel(c); err != nil {
		return errors.Errorf("unable remove channel: %v", err)
	}

	c.cfg.Broadcaster.Write(&UpdateChannelClosed{
		UserID:    c.UserID,
		ChannelID: c.ChannelID,
		Fee:       c.CloseFee,
		Duration:  time.Now().UnixNano() - lastStateTime,
	})

	return nil
}

// IsPending returns is the channel going thorough the stage of being accepted
// in blockchain. It either updating, opening or closing.
func (c *Channel) IsPending() bool {
	// TODO(andrew.shvv) Channel is not pending if it is in update mode,
	// because of the dynamic mode, for more info watch lightning labs conner
	// l2 summit watchtower video.
	currentState := c.CurrentState()
	return currentState.Name == ChannelOpening ||
		currentState.Name == ChannelClosing ||
		currentState.Name == ChannelUpdating
}

// IsConnected returns does this channel could be used for receiving and sending
// payment. For channel to be active it should be in proper state and user of
// this channel should be connected to hub.
func (c *Channel) IsActive() bool {
	return c.IsUserConnected && !c.IsPending()
}

// FundingFee is the amount of money which was spent to open this channel.
func (c *Channel) FundingFee() BalanceUnit {
	if c.Initiator == RouterInitiator {
		return c.OpenFee
	}

	// If user is initiator of this channel than we are
	// not paying funding fee.
	return 0
}

// SetConfig is used to set config which is used for using the external
// subsystems by channel.
func (c *Channel) SetConfig(cfg *ChannelConfig) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	// Copy config file
	c.cfg = &(*cfg)
	return nil
}

// SetUserConnected sets user of this channel as being connected,
// which means that hub and user could exchange protocol message and use
// channel for payments.
func (c *Channel) SetUserConnected(isConnected bool) error {
	c.IsUserConnected = isConnected
	return c.cfg.Storage.UpdateChannel(c)
}
