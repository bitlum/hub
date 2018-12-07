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
	state.LastForwardIndex = index
	return d.Save(state).Error
}

// LastForwardingIndex return last lnd forwarding pagination index of
// which were preceded by the hub.
//
// NOTE: Part of the lnd interface.
func (d *DB) LastForwardingIndex() (uint32, error) {
	state := &Counters{}
	return state.LastForwardIndex, d.FirstOrCreate(state).Error
}

// PutLastOutgoingPaymentTime saves last outgoing payment time, which is
// used to properly synchronise payment table of our lightning network
// node.
//
// NOTE: Part of the lnd interface.
func (d *DB) PutLastOutgoingPaymentTime(lastTime int64) error {
	// Create the state sync db entry if there is no one.
	state := &Counters{}
	if err := d.FirstOrCreate(state).Error; err != nil {
		return err
	}

	// Update entry with new time
	state.LastOutgoingPaymentTime = lastTime
	return d.Save(state).Error
}

// LastOutgoingPaymentTime returns last outgoing payment time, which is
// used to properly synchronise payment table of our lightning network
// node.
//
// NOTE: Part of the lnd interface.
func (d *DB) LastOutgoingPaymentTime() (int64, error) {
	state := &Counters{}
	return state.LastOutgoingPaymentTime, d.FirstOrCreate(state).Error
}
