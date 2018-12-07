package sqlite

import (
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/lightning/lnd"
)

// Runtime check to ensure that DB implements lnd.IndexesStorage interface.
var _ lnd.IndexesStorage = (*DB)(nil)

// GetUserIDByShortChanID returns user id by the given lightning network
// specification channel id.
func (d *DB) GetUserIDByShortChanID(shortChanID uint64) (lightning.UserID, error) {
	index := UserIDShortChanIDIndex{}
	if err := d.Where("short_channel_id = ?", shortChanID).
		Find(&index).Error; err != nil {
		return "", err
	}

	return lightning.UserID(index.UserID), nil
}

// AddUserIDToShortChanIDIndex ads pubkey <> short_channel_id mapping
// in storage, so that it could be later retrieved without making
// requests to the lnd daemon.
func (d *DB) AddUserIDToShortChanIDIndex(userID lightning.UserID,
	shortChanID uint64) error {
	return d.Save(&UserIDShortChanIDIndex{
		ShortChannelID: shortChanID,
		UserID:         string(userID),
	}).Error
}

// GetChannelPointByShortChanID returns out internal channel
// identification i.e. channel point by the given lnd short channel id.
func (d *DB) GetChannelPointByShortChanID(shortChanID uint64) (lightning.ChannelID, error) {
	index := ChannelIDShortChanIDIndex{ShortChannelID: shortChanID}
	if err := d.Where("short_channel_id = ?", shortChanID).
		Find(&index).Error; err != nil {
		return "", err
	}

	return lightning.ChannelID(index.ChannelID), nil
}

// AddChannelPointToShortChanIDIndex add index channel_point
// <> short_chan_id mapping.
func (d *DB) AddChannelPointToShortChanIDIndex(channelID lightning.ChannelID,
	shortChanID uint64) error {
	return d.Save(&ChannelIDShortChanIDIndex{
		ShortChannelID: shortChanID,
		ChannelID:      string(channelID),
	}).Error
}
