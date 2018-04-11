package db

import (
	"github.com/bitlum/hub/manager/router/lnd"
	"github.com/coreos/bbolt"
	"encoding/json"
)

var (
	lndBucket = []byte("lnd")
	stateKey  = []byte("state_key")
	indexKey  = []byte("index_key")
)

// Runtime check to ensure that Connector implements common.LightningConnector
// interface.
var _ lnd.DB = (*DB)(nil)

// PutLastForwardingIndex is used to save last forward pagination index
// which was used for getting forwarding events. With this we avoid
// processing of the same forwarding events twice.
//
// NOTE: Part of the lnd.DB interface.
func (d *DB) PutLastForwardingIndex(index uint32) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(lndBucket)
		if err != nil {
			return err
		}

		// TODO(andrew.shvv) replace with binary representation?
		data, err := json.Marshal(index)
		if err != nil {
			return err
		}

		return bucket.Put(indexKey, data)
	})
}

// LastForwardingIndex return last lnd forwarding pagination index of
// which were preceded by the hub.
//
// NOTE: Part of the lnd.DB interface.
func (d *DB) LastForwardingIndex() (uint32, error) {
	var index uint32

	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(lndBucket)
		if bucket == nil {
			return nil
		}

		data := bucket.Get(indexKey)
		if data == nil {
			// Equal to returning zero index
			return nil
		}

		// TODO(andrew.shvv) replace with binary representation?
		return json.Unmarshal(data, &index)
	})

	return index, err
}

// PutChannelsState is used to save the local topology of the router,
// in order to later determine what has changed.
//
// NOTE: Part of the lnd.DB interface.
func (d *DB) PutChannelsState(state map[string]string) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(lndBucket)
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
		bucket := tx.Bucket(lndBucket)
		if bucket == nil {
			return nil
		}

		data := bucket.Get(stateKey)
		if data == nil {
			// Equal to returning zero index
			return nil
		}

		// TODO(andrew.shvv) replace with binary representation?
		return json.Unmarshal(data, &state)
	})

	return state, err
}
