package lnd

import "github.com/bitlum/hub/manager/router"

// RouterStorage is the storage which is needed for keeping the info about hub
// synchronisation state.
type RouterStorage interface {
	router.ChannelStorage
	router.UserStorage
	router.PaymentStorage
	router.InfoStorage

	IndexesStorage
	SyncStorage

	// GetUserIDByChannelID returns user id by channel id, i.e. channel point.
	GetUserIDByChannelID(router.ChannelID) (router.UserID, error)
}

type SyncStorage interface {
	// PutLastForwardingIndex is used to save last forward pagination index
	// which was used for getting forwarding events. With this we avoid
	// processing of the same forwarding events twice.
	PutLastForwardingIndex(uint32) error

	// LastForwardingIndex return last lnd forwarding pagination index of
	// which were preceded by the hub.
	LastForwardingIndex() (uint32, error)

	// PutLastOutgoingPaymentTime saves last outgoing payment time, which is
	// used to properly synchronise payment table of our lightning network
	// node.
	PutLastOutgoingPaymentTime(int64) error

	// LastOutgoingPaymentTime returns last outgoing payment time, which is
	// used to properly synchronise payment table of our lightning network
	// node.
	LastOutgoingPaymentTime() (int64, error)
}

// IndexesStorage this is lnd specific index storage which maps lightning
// network short channel id, with lnd channel point. This storage is needed
// because in some cases lnd returns short channel id, and in some
// cases its channel point.
//
// NOTE: Channel point if the tuple of (tx_id, tx_index) it is used inside lnd
// and in some cases returned from rpc responses. Short channel id id the
// tuple of (block_number, tx_index, tx_position). Both of this
// identifications uniquely identify channel. We use channel point as
// channel id.
type IndexesStorage interface {
	// GetUserIDByShortChanID returns user id by the given lightning network
	// specification channel id.
	GetUserIDByShortChanID(shortChanID uint64) (router.UserID, error)

	// AddUserIDToShortChanIDIndex ads pubkey <> short_channel_id mapping
	// in storage, so that it could be later retrieved without making
	// requests to the lnd daemon.
	AddUserIDToShortChanIDIndex(userID router.UserID,
		shortChanID uint64) error

	// GetChannelPointByShortChanID returns out internal channel
	// identification i.e. channel point by the given lnd short channel id.
	GetChannelPointByShortChanID(shortChanID uint64) (router.ChannelID, error)

	// AddChannelPointToShortChanIDIndex add index channel_point
	// <> short_chan_id mapping.
	AddChannelPointToShortChanIDIndex(channelID router.ChannelID,
		shortChanID uint64) error
}
