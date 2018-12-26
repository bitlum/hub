package lightning

import (
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
)

// ChannelID uniquely identifies the channel in the lightning network.
// Currently lightning network channel point is used as channel id.
type ChannelID string

// ChannelInitiator is the side which initiated the open or close of channel.
type ChannelInitiator string

const (
	// RemoteInitiator is used when close or open was initiated from
	// the remote side.
	RemoteInitiator ChannelInitiator = "remote"

	// LocalInitiator is used when channel close or open was
	// initiated by the local side.
	LocalInitiator ChannelInitiator = "local"
)

// Channel represent the Lightning Network channel.
type Channel struct {
	ChannelID ChannelID
	NodeID    NodeID

	State  ChannelStateName
	States map[ChannelStateName]interface{}
}

// ...
func (c *Channel) CurrentState() ChannelStateName {
	return c.State
}

// ...
func (c *Channel) OpeningTime() (int64, error) {
	openingState, ok := c.States[ChannelOpening]
	if !ok {
		return 0, errors.Errorf("channel state not found")
	}

	return openingState.(*ChannelStateOpening).CreationTime, nil
}

// ...
func (c *Channel) ClosingTime() (int64, error) {
	state, ok := c.States[ChannelClosing]
	if !ok {
		return 0, errors.Errorf("channel state not found")
	}

	return state.(*ChannelStateClosing).CreationTime, nil
}

// ...
func (c *Channel) StuckBalance() (btcutil.Amount, error) {
	state, ok := c.States[ChannelOpened]
	if !ok {
		return 0, errors.Errorf("channel state not found")
	}

	return state.(*ChannelStateOpened).StuckBalance, nil
}

// ...
func (c *Channel) LimboBalance() (btcutil.Amount, error) {
	state, ok := c.States[ChannelClosing]
	if !ok {
		return 0, errors.Errorf("channel state not found")
	}

	return state.(*ChannelStateClosing).LockedBalance, nil
}

// ...
func (c *Channel) SwipeFee() (btcutil.Amount, error) {
	state, ok := c.States[ChannelClosing]
	if !ok {
		return 0, errors.Errorf("channel state not found")
	}

	return state.(*ChannelStateClosing).SwipeFee, nil
}

// ...
func (c *Channel) OpenFee() (btcutil.Amount, error) {
	state, ok := c.States[ChannelOpening]
	if !ok {
		return 0, errors.Errorf("channel state not found")
	}

	return state.(*ChannelStateOpening).OpenFee, nil
}

// ...
func (c *Channel) CommitFee() (btcutil.Amount, error) {
	switch c.State {

	// Commitment transaction fee might be exist on opening state.
	case ChannelOpening:
		state, ok := c.States[ChannelOpening]
		if !ok {
			return 0, errors.Errorf("channel state not found")
		}

		return state.(*ChannelStateOpening).CommitFee, nil

	// With time commitment transaction fee might change of the commitment
	// transaction size, after channel is in the closing or closed state,
	// commitment fee, doesn't make a lot of sense anymore.
	case ChannelOpened:
		state, ok := c.States[ChannelOpened]
		if !ok {
			return 0, errors.Errorf("channel state not found")
		}

		return state.(*ChannelStateOpened).CommitFee, nil

	default:
		return 0, errors.Errorf("unhandled channel state(%v)", c.State)
	}
}

// ...
func (c *Channel) CloseFee() (btcutil.Amount, error) {
	state, ok := c.States[ChannelClosing]
	if !ok {
		return 0, errors.Errorf("channel state not found")
	}

	return state.(*ChannelStateClosing).CloseFee, nil
}

// ...
func (c *Channel) IsActive() bool {
	switch c.State {
	case ChannelOpened:
		state, ok := c.States[ChannelOpened]
		if !ok {
			return false
		}

		return state.(*ChannelStateOpened).IsActive
	}

	return false
}
