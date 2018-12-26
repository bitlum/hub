package lightning

import "github.com/bitlum/hub/common/broadcast"

// UpdatesStreamer is an entity which send updates about lightning network node
// updates.
type UpdatesStreamer interface {
	// RegisterOnUpdates returns register which returns updates about
	// lightning client local network topology changes, about attempts of
	// propagating the payment through the lightning node, about fee changes etc.
	RegisterOnUpdates() *broadcast.Receiver
}

// UpdateChannelClosing is sent when channel was putted in closing state.
// This update might be sent several times, because some information inside
// update might change.
type UpdateChannelClosing struct {
	*ChannelStateClosing
}

func (u *UpdateChannelClosing) String() string {
	return "channel_closing"
}

// UpdateChannelClosed is sent
type UpdateChannelClosed struct {
	*ChannelStateClosed
}

func (u *UpdateChannelClosed) String() string {
	return "channel_closed"
}

// UpdateChannelOpening is used as notifications from lightning client or
// network that channel started to opening, and wait for blockchain confirmation.
type UpdateChannelOpening struct {
	*ChannelStateOpening
}

func (u *UpdateChannelOpening) String() string {
	return "channel_opening"
}

// UpdateChannelOpened is used as notifications from lightning client
// that channel has been opened.
type UpdateChannelOpened struct {
	*ChannelStateOpened
}

func (u *UpdateChannelOpened) String() string {
	return "channel_opened"
}

// UpdatePayment is sent whenever we receive or send lightning network
// payment.
type UpdatePayment struct {
	*Payment
}

// UpdateForwardPayment is sent whenever we successfully propagate payment in
// lightning network.
type UpdateForwardPayment struct {
	*ForwardPayment
}

func (u *UpdatePayment) String() string {
	return "payment"
}
