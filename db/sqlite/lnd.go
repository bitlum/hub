package sqlite

import (
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/lightning/lnd"
)

// Runtime check to ensure that DB implements lnd.ClientStorage interface.
var _ lnd.ClientStorage = (*DB)(nil)

// GetUserIDByChannelID returns user id by channel id, i.e. channel point.
func (d *DB) GetUserIDByChannelID(id lightning.ChannelID) (lightning.UserID, error) {
	channel := Channel{ID: string(id)}
	if err := d.Find(&channel).Error; err != nil {
		return "", err
	}

	return lightning.UserID(channel.UserID), nil
}
