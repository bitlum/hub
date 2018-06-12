package sqlite

import (
	"github.com/bitlum/hub/manager/router/lnd"
)

// Runtime check to ensure that DB implements lnd.SyncStorage interface.
var _ lnd.SyncStorage = (*DB)(nil)

// PutLastForwardingIndex is used to save last forward pagination index
// which was used for getting forwarding events. With this we avoid
// processing of the same forwarding events twice.
//
// NOTE: Part of the lnd interface.
func (d *DB) PutLastForwardingIndex(index uint32) error {
	// Create the state sync db entry if there is no one.
	state := &Counters{}
	if err := d.FirstOrCreate(state).Error; err != nil {
		return err
	}

	// Update entry with new index
	state.ForwardIndex = index
	return d.Save(state).Error
}

// LastForwardingIndex return last lnd forwarding pagination index of
// which were preceded by the hub.
//
// NOTE: Part of the lnd interface.
func (d *DB) LastForwardingIndex() (uint32, error) {
	state := &Counters{}
	return state.ForwardIndex, d.FirstOrCreate(state).Error
}

// PutChannelsState is used to save the local topology of the router,
// in order to later determine what has changed.
//
// NOTE: Part of the lnd interface.
func (d *DB) PutChannelsState(state map[string]string) error {
	for chanID, state := range state {
		channel := &Channel{
			ID:    chanID,
			State: state,
		}

		if err := d.Save(channel).Error; err != nil {
			return err
		}
	}

	return nil
}

// ChannelsState is used to return previously saved local topology of the
// router.
//
// NOTE: Part of the lnd interface.
func (d *DB) ChannelsState() (map[string]string, error) {
	var channels []Channel
	db := d.Find(&channels)
	if err := db.Error; err != nil {
		return nil, err
	}

	state := make(map[string]string, len(channels))
	for _, c := range channels {
		state[c.ID] = c.State
	}

	return state, nil
}
