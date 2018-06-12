package sqlite

import (
	"github.com/bitlum/hub/manager/router"
)

// Runtime check to ensure that DB implements lnd.SyncStorage interface.
var _ router.InfoStorage = (*DB)(nil)

// StorePayment saves the payment.
func (d *DB) StorePayment(p *router.DbPayment) error {
	payment := &Payment{
		FromPeer: p.FromPeer,
		ToPeer:   p.ToPeer,
		Amount:   p.Amount,
		Status:   p.Status,
		Type:     p.Type,
		Time:     p.Time,
	}

	return d.Save(payment).Error
}

// UpdatePeers updates information about set of online and peers
// connected to the hub.
func (d *DB) UpdatePeers(peers []*router.DbPeer) error {
	for _, peer := range peers {
		if err := d.Save(&Peer{
			ID:           peer.PubKey,
			Alias:        peer.Alias,
			LockedByPeer: peer.LockedByPeer,
			LockedByHub:  peer.LockedByHub,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}

// UpdateInfo updates information about the hub lightning network node.
func (d *DB) UpdateInfo(info *router.DbInfo) error {
	d.nodeInfo = info
	return nil
}

// Payments returns the payments happening inside the hub local network,
func (d *DB) Payments() ([]*router.DbPayment, error) {
	var payments []Payment
	db := d.Find(&payments)
	if err := db.Error; err != nil {
		return nil, err
	}

	pmts := make([]*router.DbPayment, len(payments))
	for i, p := range payments {
		pmts[i] = &router.DbPayment{
			FromPeer: p.FromPeer,
			ToPeer:   p.ToPeer,
			Amount:   p.Amount,
			Status:   p.Status,
			Type:     p.Type,
			Time:     p.Time,
		}
	}

	return pmts, nil
}

// Peers return the peers active and connected to the hub.
func (d *DB) Peers() ([]*router.DbPeer, error) {
	var peers []Peer
	db := d.Find(&peers)
	if err := db.Error; err != nil {
		return nil, err
	}

	prs := make([]*router.DbPeer, len(peers))
	for i, p := range peers {
		prs[i] = &router.DbPeer{
			PubKey:       p.ID,
			Alias:        p.Alias,
			LockedByPeer: p.LockedByPeer,
			LockedByHub:  p.LockedByHub,
		}
	}

	return prs, nil
}

// OurInfo returns hub lighting network node information.
func (d *DB) Info() (*router.DbInfo, error) {
	return d.nodeInfo, nil
}
