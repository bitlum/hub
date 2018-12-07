package sqlite

import (
	"github.com/bitlum/hub/manager/router"
)

// Runtime check to ensure that DB implements router.ChannelStorage interface.
var _ router.ChannelStorage = (*DB)(nil)

// UpdateChannel saves channel without saving its states.
//
// NOTE: Part the the router.ChannelStorage interface
func (d *DB) UpdateChannel(channel *router.Channel) error {
	return d.Save(&Channel{
		ID:              string(channel.ChannelID),
		UserID:          string(channel.UserID),
		OpenFee:         int64(channel.OpenFee),
		UserBalance:     int64(channel.UserBalance),
		RouterBalance:   int64(channel.RouterBalance),
		Initiator:       string(channel.Initiator),
		IsUserConnected: channel.IsUserConnected,
		CloseFee:        int64(channel.CloseFee),
	}).Error
}

// RemoveChannel removes the channel and associated with it states.
//
// NOTE: Part the the router.ChannelStorage interface
func (d *DB) RemoveChannel(channel *router.Channel) (err error) {
	tx := d.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	var states []State
	err = d.Model(&State{}).
		Find(&states, "channel_id = ?", channel.ChannelID).
		Error
	if err != nil {
		return err
	}

	err = tx.Model(&State{}).Delete(states).Error
	if err != nil {
		return err
	}

	chanID := string(channel.ChannelID)
	err = tx.Delete(&Channel{ID: chanID}).Error
	return
}

// Channels is used to return previously saved local topology of the
// router.
//
// NOTE: Part the the router.ChannelStorage interface
func (d *DB) Channels() ([]*router.Channel, error) {
	var channels []Channel
	if err := d.Find(&channels).Error; err != nil {
		return nil, err
	}

	routerChannels := make([]*router.Channel, len(channels))

	for i, channel := range channels {
		var states []State
		err := d.Model(&State{}).
			Find(&states, "channel_id = ?", channel.ID).
			Error
		if err != nil {
			return nil, err
		}

		routerStates := make([]*router.ChannelState, len(states))
		for i, state := range states {
			routerStates[i] = &router.ChannelState{
				Time: state.Time,
				Name: router.ChannelStateName(state.Name),
			}
		}

		routerChannels[i] = &router.Channel{
			ChannelID:       router.ChannelID(channel.ID),
			UserID:          router.UserID(channel.UserID),
			OpenFee:         router.BalanceUnit(channel.OpenFee),
			UserBalance:     router.BalanceUnit(channel.UserBalance),
			RouterBalance:   router.BalanceUnit(channel.RouterBalance),
			Initiator:       router.ChannelInitiator(channel.Initiator),
			CloseFee:        router.BalanceUnit(channel.CloseFee),
			IsUserConnected: channel.IsUserConnected,
			States:          routerStates,
		}
	}

	return routerChannels, nil
}

// AddChannelState adds state to the channel's state array. State array
// should be initialised in the Channel object on the stage of getting
// channels.
//
// NOTE: Part the the router.ChannelStorage interface
func (d *DB) AddChannelState(chanID router.ChannelID,
	state *router.ChannelState) error {
	channel := &Channel{ID: string(chanID)}
	return d.Model(channel).Association("States").Append(&State{
		ChannelID: string(chanID),
		Time:      state.Time,
		Name:      string(state.Name),
	}).Error
}
