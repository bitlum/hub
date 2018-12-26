package lnd

import (
	"github.com/bitlum/hub/lightning"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
)

// ErrorChannelInfoNotFound...
var ErrorChannelInfoNotFound = errors.Errorf("ErrorChannelInfoNotFound")

// InfoStorage is the storage which is needed for keeping the info about
// lightning client channel synchronisation state.
type InfoStorage interface {
	// ...
	UpdateChannelAdditionalInfo(*ChannelAdditionalInfo) error

	// ...
	GetChannelAdditionalInfoByID(lightning.ChannelID) (*ChannelAdditionalInfo, error)

	// ...
	GetChannelAdditionalInfoByShortID(uint64) (*ChannelAdditionalInfo, error)
}

// ChannelAdditionalInfo is used to store additional data about channel
// transition states, so that they could be used later to properly populate
// response from lightning network client.
type ChannelAdditionalInfo struct {
	// Save user id <> short channel id index in order to retrieve it
	// later on payment notifications without query lightning network
	// daemon.
	NodeID    lightning.NodeID
	ChannelID lightning.ChannelID

	// ShortChannelID is the identificator which is used to uniquely identify
	// channel.
	//
	// NOTE: Channel point if the tuple of (tx_id, tx_index) it is used inside lnd
	// and in some cases returned from rpc responses. Short channel id the
	// tuple of (block_number, tx_index, tx_position). Both of this
	// identifications equally identify channel. We use channel point as
	// channel id.
	ShortChannelID uint64

	// OpeningTime is the time when channel has been moved in opening state.
	OpeningTime int64

	// OpeningInitiator is the initiator of channel open.
	OpeningInitiator lightning.ChannelInitiator

	// OpeningCommitFees is the initial estimation of how much initiator of
	// channel open need to pay on cooperative channel close.
	OpeningCommitFees btcutil.Amount

	// OpeningFees is the how much funds has been spent on channel open.
	OpeningFees btcutil.Amount

	// OpeningRemoteBalance is the initial balance of remote party locked in
	// channel.
	OpeningRemoteBalance btcutil.Amount

	// OpeningLocalBalance is our initially balance locked in channel.
	OpeningLocalBalance btcutil.Amount

	// OpenTime is time when channel has been moved in open state.
	OpenTime int64

	// OpenCommitFees is current estimation of how much funds are need to
	// cooperatively close the channel.
	OpenCommitFees btcutil.Amount

	// OpenRemoteBalance is the current balance on which is locked remote
	// side in channel
	OpenRemoteBalance btcutil.Amount

	// OpenLocalBalance is the current balance on which is locked by us
	// in channel
	OpenLocalBalance btcutil.Amount

	// OpenStuckBalance is the number of funds which are locked in pending
	// payments.
	OpenStuckBalance btcutil.Amount

	// ClosingTime time when channel has moved in closing state.
	ClosingTime int64

	// ClosingFees is how much funds has been spend to close the channel.
	ClosingFees btcutil.Amount

	// ClosingRemoteBalance is how much funds has been locked on remote side
	// on the moment of closing the channel.
	ClosingRemoteBalance btcutil.Amount

	// ClosingLocalBalance is how much funds has been locked on local side
	// on the moment of closing the channel.
	ClosingLocalBalance btcutil.Amount

	// SwipeFees is how much finds we spent on swiping pending htlc.
	SwipeFees btcutil.Amount

	// CloseTime is time when channel has been closed, i.e. we returned all
	// money back to wallet.
	CloseTime int64

	// State last synced state.
	State lightning.ChannelStateName
}
