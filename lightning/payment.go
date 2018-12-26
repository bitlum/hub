package lightning

import (
	"github.com/btcsuite/btcutil"
)

// PaymentHash it is string which uniquely identifies payment in lightning
// network.
type PaymentHash string

type ForwardPayment struct {
	// FromNode adjacent node from which payment has been received.
	FromNode NodeID

	// ToNode adjacent node to which payment will be forwarded.
	ToNode NodeID

	// FromChannel it is channel from which payment has been received.
	FromChannel ChannelID

	// ToChannel it is channel to which payment has been forwarded.
	ToChannel ChannelID

	// IncomingAmount it is amount which we receive from incoming channel,
	// and from which we take fee.
	IncomingAmount btcutil.Amount

	// OutgoingAmount it is amount which propagate farther,
	// and it is lower than incoming amount.
	OutgoingAmount btcutil.Amount

	// ForwardFee fee which is taken by us for payment propagation.
	ForwardFee btcutil.Amount

	// Time time of forwarding the payment.
	Time int64
}

// PaymentStatus denotes the stage of the processing the payment.
type PaymentStatus string

var (
	// AllStatuses denotes all statuses.
	AllStatuses PaymentStatus = ""

	// Waiting means that payment has been created and waiting to be approved
	// for sending.
	Waiting PaymentStatus = "Waiting"

	// Pending means that service is seeing the payment, but it not yet approved
	// from the its POV.
	Pending PaymentStatus = "Pending"

	// Completed in case of outgoing/incoming payment this means that we
	// sent/received the transaction in/from the network and it was confirmed
	// number of times service believe sufficient. In case of the forward
	// transaction it means that we successfully routed it through and
	// earned fee for that.
	Completed PaymentStatus = "Completed"

	// Failed means that services has tried to send payment for couple of
	// times, but without success, and now service gave up.
	Failed PaymentStatus = "Failed"
)

// PaymentDirection denotes the direction of the payment, whether payment is
// going form us to someone else, or form someone else to us.
type PaymentDirection string

var (
	// AllDirections denotes all directions.
	AllDirections PaymentDirection = ""

	// Incoming type of payment which service has received from someone else
	// in the media.
	Incoming PaymentDirection = "Incoming"

	// Outgoing type of payment which service has sent to someone else in the
	// media.
	Outgoing PaymentDirection = "Outgoing"
)

// PaymentSystem denotes is that payment belongs to business logic of payment
// server or it was originated by user / third-party service.
type PaymentSystem string

var (
	// AllSystems denotes all systems.
	AllSystems PaymentSystem = ""

	// Internal type of payment usually services the purpose of payment
	// server itself for stabilisation of system. In lightning it might
	// channel re-balancing of channel, or ping payments.
	Internal PaymentSystem = "Internal"

	// External type of payment which was originated by user / third-party
	// services, this is what usually interesting for external viewer.
	External PaymentSystem = "External"
)

type Payment struct {
	// PaymentID it is unique identificator of the payment generated inside
	// the system.
	PaymentID string

	// Receiver is an identificator of node which received the payment.
	Receiver NodeID

	// UpdatedAt denotes the time when payment object has been last updated.
	UpdatedAt int64

	// Status denotes the stage of the processing the payment.
	Status PaymentStatus

	// Direction denotes the direction of the payment, whether payment is
	// going form us to someone else, or form someone else to us.
	Direction PaymentDirection

	// System denotes is that payment belongs to business logic of payment
	// server or it was originated by user / third-party service.
	System PaymentSystem

	// Amount is the number of funds which receiver gets at the end.
	Amount btcutil.Amount

	// MediaFee is the fee which is taken by the blockchain or lightning
	// network in order to propagate the payment.
	MediaFee btcutil.Amount

	// PaymentHash it is string which uniquely identifies payment in lightning
	// network.
	PaymentHash PaymentHash
}
