package sqlite

import (
	"github.com/bitlum/hub/lightning"
)

// Runtime check to ensure that DB implements lnd.InfoStorage interface.
// var _ lightning.InfoStorage = (*DB)(nil)

// UpdateInfo updates information about the hub lightning network node.
func (d *DB) UpdateInfo(info *lightning.Info) error {
	d.nodeInfo = info
	return nil
}

// Info returns hub lighting network node information.
func (d *DB) Info() (*lightning.Info, error) {
	return d.nodeInfo, nil
}
