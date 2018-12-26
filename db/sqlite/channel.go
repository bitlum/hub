package sqlite

// Runtime check to ensure that DB implements lightning.ChannelStorage interface.
// var _ lightning.ChannelStorage = (*DB)(nil)

// UpdateChannel saves channel without saving its states.
//
// NOTE: Part the the lightning.ChannelStorage interface
//func (d *DB) UpdateChannel(channel *lightning.Channel) error {
//	return d.Save(&Channel{
//		ID:              string(channel.ChannelID),
//		UserID:          string(channel.NodeID),
//		OpenFee:         int64(channel.OpenFee),
//		RemoteBalance:   int64(channel.RemoteBalance),
//		LocalBalance:    int64(channel.LocalBalance),
//		Initiator:       string(channel.Initiator),
//		IsUserConnected: channel.IsUserConnected,
//		CloseFee:        int64(channel.CommitFee),
//	}).Error
//}
//
//// RemoveChannel removes the channel and associated with it states.
////
//// NOTE: Part the the lightning.ChannelStorage interface
//func (d *DB) RemoveChannel(channel *lightning.Channel) (err error) {
//	tx := d.Begin()
//	defer func() {
//		if err != nil {
//			tx.Rollback()
//		} else {
//			tx.Commit()
//		}
//	}()
//
//	var states []State
//	err = d.Model(&State{}).
//		Find(&states, "channel_id = ?", channel.ChannelID).
//		Error
//	if err != nil {
//		return err
//	}
//
//	err = tx.Model(&State{}).Delete(states).Error
//	if err != nil {
//		return err
//	}
//
//	chanID := string(channel.ChannelID)
//	err = tx.Delete(&Channel{ID: chanID}).Error
//	return
//}
//
//// Channels is used to return previously saved local topology of the
//// lightning.
////
//// NOTE: Part the the lightning.ChannelStorage interface
//func (d *DB) Channels() ([]*lightning.Channel, error) {
//	var channels []Channel
//	if err := d.Find(&channels).Error; err != nil {
//		return nil, err
//	}
//
//	nodeChannels := make([]*lightning.Channel, len(channels))
//
//	for i, channel := range channels {
//		var states []State
//		err := d.Model(&State{}).
//			Find(&states, "channel_id = ?", channel.ID).
//			Error
//		if err != nil {
//			return nil, err
//		}
//
//		channelStates := make([]*lightning.ChannelState, len(states))
//		for i, state := range states {
//			channelStates[i] = &lightning.ChannelState{
//				Time: state.Time,
//				Name: lightning.ChannelStateName(state.Name),
//			}
//		}
//
//		nodeChannels[i] = &lightning.Channel{
//			ChannelID:       lightning.ChannelID(channel.ID),
//			NodeID:          lightning.NodeID(channel.UserID),
//			OpenFee:         btcutil.Amount(channel.OpenFee),
//			RemoteBalance:   btcutil.Amount(channel.RemoteBalance),
//			LocalBalance:    btcutil.Amount(channel.LocalBalance),
//			Initiator:       lightning.ChannelInitiator(channel.Initiator),
//			CommitFee:       btcutil.Amount(channel.CloseFee),
//			IsUserConnected: channel.IsUserConnected,
//			State:           channelStates,
//		}
//	}
//
//	return nodeChannels, nil
//}
//
//// AddChannelState adds state to the channel's state array. State array
//// should be initialised in the Channel object on the stage of getting
//// channels.
////
//// NOTE: Part the the lightning.ChannelStorage interface
//func (d *DB) AddChannelState(chanID lightning.ChannelID,
//	state *lightning.ChannelState) error {
//	channel := &Channel{ID: string(chanID)}
//	return d.Model(channel).Association("State").Append(&State{
//		ChannelID: string(chanID),
//		Time:      state.Time,
//		Name:      string(state.Name),
//	}).Error
//}
