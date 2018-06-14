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
