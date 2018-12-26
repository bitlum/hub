package lightning

import (
	"github.com/btcsuite/btcutil"
)

type ChannelStateName string

const (
	// ChannelOpening denotes that channel open request has been sent in
	// blockchain network, and that we wait for its approval.
	ChannelOpening ChannelStateName = "opening"

	// ChannelOpened denotes that channel open request was approved in
	// blockchain, and it could be used for routing the payments.
	ChannelOpened ChannelStateName = "opened"

	// ChannelClosing denotes that channel close request has been sent in
	// blockchain network, and that we wait for its approval.
	// Channel couldn't be used for routing payments, and funds in this
	// channel are still couldn't be used.
	ChannelClosing ChannelStateName = "closing"

	// ChannelClosed denotes that channel close request was approved in blockchain,
	// and couldn't be used for routing payment anymore, locked on our side
	// funds now are back in wallet.
	ChannelClosed ChannelStateName = "closed"
)

// ChannelStateOpening denotes that channel open request has been sent in
// blockchain network, and that we wait for its approval.
type ChannelStateOpening struct {
	ChannelID    ChannelID
	CreationTime int64

	// CommitFee is an estimated amount of funds which will be needed to
	// close the channel and unlock funds.
	CommitFee btcutil.Amount

	// OpenFee is the number of funds which were taken by miners to proceed
	// operation of channel open.
	OpenFee btcutil.Amount

	// RemoteBalance funds which were locked from remote side on initial step
	// of channel creation.
	RemoteBalance btcutil.Amount

	// LocalBalance funds which where locked by from our side on initial step
	// of channel creation.
	LocalBalance btcutil.Amount

	// Initiator side which initiated open of the channel.
	Initiator ChannelInitiator
}

// ChannelStateOpened denotes that channel open request was approved in
// blockchain, and it could be used for routing the payments.
type ChannelStateOpened struct {
	ChannelID    ChannelID
	CreationTime int64

	// CommitFee is an estimated amount of funds which will be needed to
	// close the channel and unlock funds.
	CommitFee btcutil.Amount

	// RemoteBalance funds which are locked from remote side on current state
	// of channel.
	RemoteBalance btcutil.Amount

	// LocalBalance funds which are locked from our side on current state
	// of channel.
	LocalBalance btcutil.Amount

	// IsActive is channel could be used for sending / forwarding payments,
	// for that our node has to be connected to peer with tcp/ip connection.
	IsActive bool

	// StuckBalance funds which are stuck in pending htlc.
	StuckBalance btcutil.Amount
}

// ChannelStateClosing denotes that channel close request has been sent in
// blockchain network, and that we wait for its approval. Channel couldn't be
// used for routing payments, and funds in this channel are still couldn't be
// used.
type ChannelStateClosing struct {
	ChannelID    ChannelID
	CreationTime int64

	// CloseFee is the number of funds which where needed to close the channel.
	CloseFee btcutil.Amount

	// SwipeFee is the number of funds which where needed to swipe
	// pending htlc.
	SwipeFee btcutil.Amount

	// RemoteBalance funds which will be given back to remote side when channel
	// will be closed.
	RemoteBalance btcutil.Amount

	// LocalBalance funds which will be given back to us when channel will be
	// closed.
	LocalBalance btcutil.Amount

	// LockedBalance funds which are stuck in the network, until lightning
	// network client will retrieve them, they couldn't be used.
	LockedBalance btcutil.Amount
}

// ChannelStateClosed denotes that channel close request was approved in blockchain,
// and couldn't be used for routing payment anymore, locked on our side
// funds now are back in wallet.
type ChannelStateClosed struct {
	ChannelID    ChannelID
	CreationTime int64

	// CloseFee fee which were paid by initiator of channel for closing this
	// channel.
	CloseFee btcutil.Amount

	// LocalBalance funds which were be given back to us when channel has be
	// closed.
	LocalBalance btcutil.Amount
}
