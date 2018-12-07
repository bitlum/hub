package lightning

import (
	"github.com/bitlum/hub/common/broadcast"
)

// Client aka payment provider, aka hub, aka lightning network node.
// This interface gives as unified way of managing different implementations of
// lightning network daemons.
type Client interface {
	// SendPayment makes the payment on behalf of lightning node.
	SendPayment(userID UserID, amount BalanceUnit) error

	// OpenChannel opens the channel with the given user.
	OpenChannel(id UserID, funds BalanceUnit) error

	// CloseChannel closes the specified channel.
	CloseChannel(id ChannelID) error

	// UpdateChannel updates the number of locked funds in the specified
	// channel.
	UpdateChannel(id ChannelID, funds BalanceUnit) error

	// SetFeeBase sets base number of milli units (i.e milli satoshis in
	// Bitcoin) which will be taken for every forwarding payment.
	SetFeeBase(feeBase int64) error

	// SetFeeProportional sets the number of milli units (i.e milli
	// satoshis in Bitcoin) which will be taken for every killo-unit of
	// forwarding payment amount as a forwarding fee.
	SetFeeProportional(feeProportional int64) error

	// RegisterOnUpdates returns register which returns updates about
	// lightning client local network topology changes, about attempts of
	// propagating the payment through the lightning node, about fee changes etc.
	RegisterOnUpdates() *broadcast.Receiver

	// Channels returns all channels which are connected to lightning node.
	Channels() ([]*Channel, error)

	// Users return all users which connected or were connected to lightning
	// node with payment channel.
	Users() ([]*User, error)

	// FreeBalance returns the amount of funds at lightning node disposal.
	FreeBalance() (BalanceUnit, error)

	// PendingBalance returns the amount of funds which in the process of
	// being accepted by blockchain.
	PendingBalance() (BalanceUnit, error)

	// Done returns error if lightning client stopped working for some reason,
	// and nil if it was stopped.
	Done() chan error

	// Asset returns asset with which corresponds to this lightning client.
	Asset() string
}
