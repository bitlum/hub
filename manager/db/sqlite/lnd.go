package sqlite

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/hub/manager/router/lnd"
)

// Runtime check to ensure that DB implements lnd.RouterStorage interface.
var _ lnd.RouterStorage = (*DB)(nil)

// GetUserIDByChannelID returns user id by channel id, i.e. channel point.
func (d *DB) GetUserIDByChannelID(id router.ChannelID) (router.UserID, error) {
	channel := Channel{ID: string(id)}
	if err := d.Find(&channel).Error; err != nil {
		return "", err
	}

	return router.UserID(channel.UserID), nil
}
