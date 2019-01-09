package lnd

import (
	"encoding/hex"
	"github.com/bitlum/graphql-go/errors"
	"github.com/bitlum/hub/common"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/btcsuite/btcutil"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/shopspring/decimal"
	"time"
)

// Runtime check to ensure that Client implements
// lightning. PaymentClient interface.
var _ lightning.PaymentClient = (*Client)(nil)

// QueryRoutes returns list of routes from to the given lnd node,
// and insures the the capacity of the channels is sufficient.
//
// NOTE: Part of the lightning.PaymentClient interface.
func (c *Client) QueryRoutes(invoiceStr string, amount btcutil.Amount,
	maxRoutes int32) ([]*lightning.Route, error) {

	return nil, errors.Errorf("not implemented")

	//m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
	//	c.cfg.MetricsBackend)
	//defer m.Finish()
	//
	//netParams, err := GetParams(c.cfg.Net)
	//if err != nil {
	//	m.AddError(metrics.HighSeverity)
	//	return nil, errors.Errorf("unable to get net params: %v", err)
	//}
	//
	//invoice, err := zpay32.Decode(invoiceStr, netParams)
	//if err != nil {
	//	m.AddError(metrics.LowSeverity)
	//	return nil, errors.Errorf("unable decode invoice: %v", err)
	//}
	//
	//// If node has route hints, than it means that destination in the invoice
	//// is not reachable by default, and there is private channel,
	//// which actually leads to final node.
	//var routes []*lnrpc.Route
	//if invoice.RouteHints == nil {
	//	pubKey := hex.EncodeToString(invoice.Destination.SerializeCompressed())
	//	req := &lnrpc.QueryRoutesRequest{
	//		PubKey:    pubKey,
	//		Amt:       int64(amount),
	//		NumRoutes: int32(maxRoutes),
	//	}
	//
	//	resp, err := c.rpc.QueryRoutes(timeout(10), req)
	//	if err != nil {
	//		return nil, errors.Errorf("unable to query routes: %v", err)
	//	}
	//
	//	routes = resp.Routes
	//
	//} else {
	//	// TODO(andrew.shvv) Add support for query routes private channels
	//	return nil, errors.Errorf("payment to private channels " +
	//		"are not supported")
	//}
	//
	//routes := make([]*lightning.Route, len(resp.Routes))
	//for i, route := range resp.Routes {
	//	var nodes []lightning.Node
	//
	//	routes[i] = &lightning.Route{
	//
	//	}
	//}
	//
	//return routes, nil
}

// EstimateFee estimate fee for the payment with the given sending
// amount, to the given node.
//
// NOTE: Part of the lightning.PaymentClient interface.
func (c *Client) EstimateFee(invoiceStr string,
	amount btcutil.Amount) (decimal.Decimal, error) {

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	if invoiceStr == "" {
		// If invoice is not specified that we unable to understand where
		// payment is going, for that reason estimate fee based on
		// previous payment experience.
		return c.averageFee.Round(8), nil

	} else {
		netParams, err := getParams(c.cfg.Net)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			return decimal.Zero, err
		}

		invoice, err := zpay32.Decode(invoiceStr, netParams)
		if err != nil {
			m.AddError(metrics.LowSeverity)
			return decimal.Zero, errors.Errorf("unable decode invoice: %v",
				err)
		}

		// If amount is not specified than we unable to understand what
		// fee we should expect, for that reason return the average one.
		var amount int64
		if invoice.MilliSat == nil {
			return c.averageFee.Round(8), nil
		} else {
			amount = int64(invoice.MilliSat.ToSatoshis())
		}

		pubKey := hex.EncodeToString(invoice.Destination.SerializeCompressed())
		req := &lnrpc.QueryRoutesRequest{
			PubKey: pubKey,
			Amt:    amount,
			FeeLimit: &lnrpc.FeeLimit{
				Limit: &lnrpc.FeeLimit_Percent{
					Percent: 3,
				},
			},
			NumRoutes: 10,
		}

		// Fee to our own node is zero.
		if pubKey == string(c.lightningNodeUserID) {
			return decimal.Zero, nil
		}

		resp, err := c.rpc.QueryRoutes(timeout(10), req)
		if err != nil {
			m.AddError(metrics.LowSeverity)
			return decimal.Zero, err
		}

		// Calculate average route fee from received routes
		var averageFeeSat decimal.Decimal
		for _, route := range resp.Routes {
			averageFeeSat = averageFeeSat.Add(decimal.New(route.TotalFees, 0))
		}
		averageFeeSat = averageFeeSat.Div(decimal.New(int64(len(resp.Routes)), 0))

		// Convert satoshis to bitcoin
		averageFeeBtc := averageFeeSat.Div(decimal.New(btcutil.SatoshiPerBitcoin, 0))
		return averageFeeBtc.Round(8), nil
	}
}

// SendPayment sends payment with the given payment hash to the given node
// in the route.
//
// NOTE: Part of the lightning.PaymentClient interface.
func (c *Client) SendPaymentToRoute(route *lightning.Route,
	paymentHash lightning.PaymentHash) (*lightning.Payment, error) {
	return nil, errors.Errorf("%v is not implemented",
		common.GetFunctionName())
}

// CreateInvoice is used to create lightning network invoice.
//
// NOTE: Part of the lightning.PaymentClient interface.
func (c *Client) CreateInvoice(receipt string, amount btcutil.Amount,
	description string) (string, *zpay32.Invoice, error) {

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(), c.cfg.MetricsBackend)
	defer m.Finish()

	expirationTime := time.Minute * 15
	invoiceReq := &lnrpc.Invoice{
		Receipt: []byte(receipt),
		Value:   int64(amount),
		Memo:    description,
		Expiry:  int64(expirationTime.Seconds()),
	}

	invoiceResp, err := c.rpc.AddInvoice(timeout(10), invoiceReq)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return "", nil, err
	}

	// Check that invoice is valid, and that amount which we are sending is
	// corresponding to what we expect.
	netParams, err := getParams(c.cfg.Net)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return "", nil, err
	}

	invoice, err := zpay32.Decode(invoiceResp.PaymentRequest, netParams)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return "", nil, err
	}

	return invoiceResp.PaymentRequest, invoice, nil
}

// ValidateInvoice takes the encoded lightning network invoice and ensure
// its valid.
//
// NOTE: Part of the lightning.PaymentClient interface.
func (c *Client) ValidateInvoice(invoiceStr string,
	amount btcutil.Amount) (*zpay32.Invoice, error) {

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	netParams, err := getParams(c.cfg.Net)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return nil, errors.Errorf("unable load network params: %v", err)
	}

	invoice, err := zpay32.Decode(invoiceStr, netParams)
	if err != nil {
		m.AddError(metrics.LowSeverity)
		return nil, errors.Errorf("unable decode invoice: %v", err)
	}

	// Only if amount is specified we need to check that it is the same as in
	// invoice. This is needed in case of wallet would like to decode
	// invoice, without checking amount.
	if amount != 0 {
		if invoice.MilliSat != nil {
			if invoice.MilliSat.ToSatoshis() != amount {
				m.AddError(metrics.LowSeverity)
				return nil, errors.Errorf("wrong amount "+
					"received(%v) and in invoice(%v)", amount,
					invoice.MilliSat.ToSatoshis())
			}
		}
	}

	return invoice, nil
}

// ListPayments returns list of incoming and outgoing payment.
//
// NOTE: Part of the lightning.PaymentClient interface.
func (c *Client) ListPayments(asset string, status lightning.PaymentStatus,
	direction lightning.PaymentDirection, system lightning.PaymentSystem) (
	[]*lightning.Payment, error) {

	select {
	case <-c.startedTrigger:
	}

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	incomingPayments, err := fetchInvoicePayments(c.rpc)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable to fetch incoming payments: %v", err)
		log.Error(err)
		return nil, err
	}

	outgoingPayments, err := fetchOutgoingPayments(c.rpc)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable to fetch outgoing payments: %v", err)
		log.Error(err)
		return nil, err
	}

	var payments []*lightning.Payment
	for _, incomingPayment := range incomingPayments {
		if !incomingPayment.Settled {
			continue
		}

		payments = append(payments, &lightning.Payment{
			PaymentID:   "",
			Receiver:    c.lightningNodeUserID,
			UpdatedAt:   incomingPayment.SettleDate,
			Status:      lightning.Completed,
			Direction:   lightning.Incoming,
			System:      lightning.External,
			Amount:      btcutil.Amount(incomingPayment.AmtPaidSat),
			MediaFee:    0,
			PaymentHash: lightning.PaymentHash(hex.EncodeToString(incomingPayment.RHash)),
		})
	}

	for _, outgoingPayment := range outgoingPayments {
		payments = append(payments, &lightning.Payment{
			PaymentID:   "",
			Receiver:    lightning.NodeID(outgoingPayment.Path[len(outgoingPayment.Path)-1]),
			UpdatedAt:   outgoingPayment.CreationDate,
			Status:      lightning.Completed,
			Direction:   lightning.Outgoing,
			System:      lightning.External,
			Amount:      btcutil.Amount(outgoingPayment.Value),
			MediaFee:    btcutil.Amount(outgoingPayment.Fee),
			PaymentHash: lightning.PaymentHash(outgoingPayment.PaymentHash),
		})
	}

	var filteredPayment []*lightning.Payment
	for _, payment := range payments {
		if payment.Status != status && status != lightning.AllStatuses {
			continue
		}

		if payment.Direction != direction && direction != lightning.AllDirections {
			continue
		}

		if payment.System != system && system != lightning.AllSystems {
			continue
		}

		filteredPayment = append(filteredPayment, payment)
	}

	return filteredPayment, nil
}

// ListForwardPayments returns list of forward payments which were routed
// thorough lightning node.
//
// NOTE: Part of the lightning.PaymentClient interface.
func (c *Client) ListForwardPayments() ([]*lightning.ForwardPayment, error) {
	select {
	case <-c.startedTrigger:
	}

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	events, err := fetchForwardingPayments(c.rpc, 0)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable to get forwarding events: %v", err)
		log.Error(err)
		return nil, err
	}

	var forwardPayments []*lightning.ForwardPayment
	for _, event := range events {

		fromChannel, err := c.cfg.Storage.GetChannelAdditionalInfoByShortID(event.ChanIdIn)
		if err != nil {
			// TODO(andrew.shvv) cache might no be in sync
			m.AddError(metrics.HighSeverity)
			log.Errorf("unable to get sender id by short"+
				" chan id(%v): %v", event.ChanIdIn, err)
			continue
		}

		toChannel, err := c.cfg.Storage.GetChannelAdditionalInfoByShortID(event.ChanIdOut)
		if err != nil {
			// TODO(andrew.shvv) cache might no be in sync
			m.AddError(metrics.HighSeverity)
			log.Errorf("unable to get receiver id by"+
				" short chan id(%v): %v", event.ChanIdOut, err)
			continue
		}

		forwardPayments = append(forwardPayments, &lightning.ForwardPayment{
			FromNode:       fromChannel.NodeID,
			ToNode:         toChannel.NodeID,
			FromChannel:    fromChannel.ChannelID,
			ToChannel:      toChannel.ChannelID,
			IncomingAmount: btcutil.Amount(event.AmtIn),
			OutgoingAmount: btcutil.Amount(event.AmtOut),
			ForwardFee:     btcutil.Amount(event.Fee),
			Time:           int64(event.Timestamp),
		})
	}

	return forwardPayments, nil
}

// PaymentByInvoice returns payment by given lightning network invoice.
func (c *Client) PaymentByInvoice(invoice string) (*lightning.Payment, error) {
	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	return nil, errors.Errorf("unsupported method: %v", common.GetFunctionName())
}
