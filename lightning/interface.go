package lightning

import (
	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/shopspring/decimal"
)

// Client aka payment provider, aka hub, aka lightning network node.
// This interface gives as unified way of managing different implementations of
// lightning network daemons.
type Client interface {
	TopologyClient
	PaymentClient

	// AvailableBalance return the amount of confirmed funds available for account.
	AvailableBalance() (btcutil.Amount, error)

	// PendingBalance returns the amount of funds which in the process of
	// being accepted by blockchain.
	PendingBalance() (btcutil.Amount, error)

	// Info returns the information about our lnd node.
	Info() (*Info, error)

	// Asset returns asset with which corresponds to this lightning client.
	Asset() string
}

type TopologyClient interface {
	// Channels returns all lightning network channels which belongs to us.
	Channels() ([]*Channel, error)

	// OpenChannel opens the lightning network channel with the given node.
	OpenChannel(nodeID NodeID, funds btcutil.Amount) error

	// CloseChannel closes the specified lightning network channel.
	CloseChannel(channelID ChannelID) error

	// ConnectToNode connects to node with tcp / ip connection.
	ConnectToNode(nodeID NodeID) error
}

type PaymentClient interface {
	// EstimateFee estimate fee for the payment with the given sending
	// amount, to the given node.
	EstimateFee(invoice string, amount btcutil.Amount) (decimal.Decimal, error)

	// QueryRoutes returns list of routes from to the given lnd node,
	// and insures the the capacity of the channels is sufficient.
	QueryRoutes(invoiceStr string, amount btcutil.Amount,
		maxRoutes int32) ([]*Route, error)

	// SendPayment makes the payment on behalf of lightning node.
	SendPaymentToRoute(route *Route, paymentHash PaymentHash) (*Payment, error)

	// CreateInvoice is used to create lightning network invoice.
	CreateInvoice(receipt string, amount btcutil.Amount,
		description string) (string, *zpay32.Invoice, error)

	// ValidateInvoice takes the encoded lightning network invoice and ensure
	// its valid.
	ValidateInvoice(invoice string, amount btcutil.Amount) (*zpay32.Invoice, error)

	// ListPayments returns list of incoming and outgoing payment.
	ListPayments(asset string, status PaymentStatus,
		direction PaymentDirection, system PaymentSystem) ([]*Payment, error)

	// ListForwardPayments returns list of forward payments which were routed
	// thorough lightning node.
	ListForwardPayments() ([]*ForwardPayment, error)

	// PaymentByInvoice returns payment by given lightning network invoice.
	PaymentByInvoice(invoice string) (*Payment, error)
}
