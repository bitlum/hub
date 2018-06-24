package router

import (
	"github.com/bitlum/hub/manager/common/broadcast"
)

// Router aka payment provider, aka hub, aka lightning network node.
// This interface gives as unified way of managing different implementations of
// lightning network daemons. With this interface we could control hub/router
// and force it into the state of equilibrium - the state with maximum
// efficiency. The exact efficiency metrics depends on chosen strategy.
type Router interface {
	// SendPayment makes the payment on behalf of router. In the context of
	// lightning network hub manager this hook might be used for future
	// off-chain channel re-balancing tactics.
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

	// RegisterOnUpdates returns register which returns updates about router
	// local network topology changes, about attempts of propagating the payment
	// through the router, about fee changes etc.
	RegisterOnUpdates() *broadcast.Receiver

	// Network returns the information about the current local network router
	// topology.
	Network() ([]*Channel, error)

	// FreeBalance returns the amount of funds at router disposal.
	FreeBalance() (BalanceUnit, error)

	// PendingBalance returns the amount of funds which in the process of
	// being accepted by blockchain.
	PendingBalance() (BalanceUnit, error)

	// Done returns error if router stopped working for some reason,
	// and nil if it was stopped.
	Done() chan error

	// Asset returns asset with which corresponds to this router.
	Asset() string
}
