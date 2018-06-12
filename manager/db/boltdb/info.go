package boltdb

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/coreos/bbolt"
	"encoding/json"
	"encoding/binary"
)

var (
	infoBucket     = []byte("info")
	paymentsBucket = []byte("payments")
	peersKey       = []byte("peers")
	peerInfoKey    = []byte("peer_info")
)

// Runtime check to ensure that DB implements lnd.SyncStorage interface.
var _ router.InfoStorage = (*DB)(nil)

// StorePayment saves the payment.
func (d *DB) StorePayment(payment *router.DbPayment) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(paymentsBucket)
		if err != nil {
			return err
		}

		index, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		// TODO(andrew.shvv) replace with binary representation?
		data, err := json.Marshal(payment)
		if err != nil {
			return err
		}

		var key [8]byte
		binary.BigEndian.PutUint64(key[:], index)
		return bucket.Put(key[:], data)
	})
	return nil
}

// UpdatePeers updates information about set of online and peers
// connected to the hub.
func (d *DB) UpdatePeers(peers []*router.DbPeer) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(infoBucket)
		if err != nil {
			return err
		}

		// TODO(andrew.shvv) replace with binary representation?
		data, err := json.Marshal(peers)
		if err != nil {
			return err
		}

		return bucket.Put(peersKey, data)
	})
	return nil
}

// UpdateInfo updates information about the hub lightning network node.
func (d *DB) UpdateInfo(info *router.DbInfo) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(infoBucket)
		if err != nil {
			return err
		}

		// TODO(andrew.shvv) replace with binary representation?
		data, err := json.Marshal(info)
		if err != nil {
			return err
		}

		return bucket.Put(peerInfoKey, data)
	})
	return nil
}

// Payments returns the payments happening inside the hub local network,
func (d *DB) Payments() ([]*router.DbPayment, error) {
	var payments []*router.DbPayment

	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(paymentsBucket)
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			// Skip buckets
			if v == nil {
				return nil
			}

			payment := &router.DbPayment{}
			if err := json.Unmarshal(v, payment); err != nil {
				return err
			}

			payments = append(payments, payment)
			return nil
		})
	})

	return payments, err
}

// Peers return the peers active and connected to the hub.
func (d *DB) Peers() ([]*router.DbPeer, error) {
	peers := make([]*router.DbPeer, 0)

	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(infoBucket)
		if bucket == nil {
			return nil
		}

		// TODO(andrew.shvv) replace with binary representation?

		data := bucket.Get(peersKey)
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, &peers)
	})

	return peers, err
}

// Info returns hub lighting network node information.
func (d *DB) Info() (*router.DbInfo, error) {
	info := &router.DbInfo{}

	err := d.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(infoBucket)
		if bucket == nil {
			return nil
		}

		// TODO(andrew.shvv) replace with binary representation?

		data := bucket.Get(peerInfoKey)
		if data == nil {
			return nil
		}

		return json.Unmarshal(data, info)
	})

	return info, err
}
