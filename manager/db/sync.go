package db

import (
	"github.com/bitlum/hub/manager/router/lnd"
	"github.com/coreos/bbolt"
	"encoding/json"
)

var (
	syncBucket = []byte("sync")
	stateKey   = []byte("state_key")
	timeKey    = []byte("time_key")
)

// Runtime check to ensure that DB implements lnd.SyncStorage interface.
var _ lnd.SyncStorage = (*DB)(nil)

// PutLastForwardingTime is used to save last forward pagination time
// which was used for getting forwarding events. With this we avoid
// processing of the same forwarding events twice.
//
// NOTE: Part of the lnd.DB interface.
func (d *DB) PutLastForwardingTime(time int64) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(syncBucket)
		if err != nil {
			return err
		}

		// TODO(andrew.shvv) replace with binary representation?
		data, err := json.Marshal(time)
		if err != nil {
			return err
		}

		return bucket.Put(timeKey, data)
	})
}

// LastForwardingTime return last lnd forwarding pagination time of
// which were preceded by the hub.
//
// NOTE: Part of the lnd.DB interface.
func (d *DB) LastForwardingTime() (int64, error) {
	var syncTime int64

	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(syncBucket)
		if bucket == nil {
			return nil
		}

		data := bucket.Get(timeKey)
		if data == nil {
			// Equal to returning zero time
			return nil
		}

		// TODO(andrew.shvv) replace with binary representation?
		return json.Unmarshal(data, &syncTime)
	})

	return syncTime, err
}

// PutChannelsState is used to save the local topology of the router,
// in order to later determine what has changed.
//
// NOTE: Part of the lnd.DB interface.
func (d *DB) PutChannelsState(state map[string]string) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(syncBucket)
		if err != nil {
			return err
		}

		// TODO(andrew.shvv) replace with binary representation?
		data, err := json.Marshal(state)
		if err != nil {
			return err
		}

		return bucket.Put(stateKey, data)
	})
}

// ChannelsState is used to return previously saved local topology of the
// router.
//
// NOTE: Part of the lnd.DB interface.
func (d *DB) ChannelsState() (map[string]string, error) {
	var state map[string]string

	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(syncBucket)
		if bucket == nil {
			return nil
		}

		data := bucket.Get(stateKey)
		if data == nil {
			// Equal to returning zero time
			return nil
		}

		// TODO(andrew.shvv) replace with binary representation?
		return json.Unmarshal(data, &state)
	})

	return state, err
}
