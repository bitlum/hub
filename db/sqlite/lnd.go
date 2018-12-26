package sqlite

import (
	"github.com/bitlum/hub/lightning"
)

// Runtime check to ensure that DB implements lnd.ClientStorage interface.
// var _ lnd.ClientStorage = (*DB)(nil)

// GetUserIDByChannelID returns user id by channel id, i.e. channel point.
func (d *DB) GetUserIDByChannelID(id lightning.ChannelID) (lightning.NodeID, error) {
	channel := Channel{ID: string(id)}
	if err := d.Find(&channel).Error; err != nil {
		return "", err
	}

	return lightning.NodeID(channel.UserID), nil
}
