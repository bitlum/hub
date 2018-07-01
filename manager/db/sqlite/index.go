package sqlite

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/hub/manager/router/lnd"
)

// Runtime check to ensure that DB implements lnd.IndexesStorage interface.
var _ lnd.IndexesStorage = (*DB)(nil)

// GetUserIDByShortChanID returns user id by the given lightning network
// specification channel id.
func (d *DB) GetUserIDByShortChanID(shortChanID uint64) (router.UserID, error) {
	index := UserIDShortChanIDIndex{ShortChannelID: shortChanID}
	if err := d.Find(&index).Error; err != nil {
		return "", err
	}

	return router.UserID(index.UserID), nil
}

// AddUserIDToShortChanIDIndex ads pubkey <> short_channel_id mapping
// in storage, so that it could be later retrieved without making
// requests to the lnd daemon.
func (d *DB) AddUserIDToShortChanIDIndex(userID router.UserID,
	shortChanID uint64) error {
	return d.Save(&UserIDShortChanIDIndex{
		ShortChannelID: shortChanID,
		UserID:         string(userID),
	}).Error
}

// GetChannelPointByShortChanID returns out internal channel
// identification i.e. channel point by the given lnd short channel id.
func (d *DB) GetChannelPointByShortChanID(shortChanID uint64) (router.ChannelID, error) {
	index := ChannelIDShortChanIDIndex{ShortChannelID: shortChanID}
	if err := d.Find(&index).Error; err != nil {
		return "", err
	}

	return router.ChannelID(index.ChannelID), nil
}

// AddChannelPointToShortChanIDIndex add index channel_point
// <> short_chan_id mapping.
func (d *DB) AddChannelPointToShortChanIDIndex(channelID router.ChannelID,
	shortChanID uint64) error {
	return d.Save(&ChannelIDShortChanIDIndex{
		ShortChannelID: shortChanID,
		ChannelID:      string(channelID),
	}).Error
}
