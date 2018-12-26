package router

import (
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
)

// Router responsibilities:
// 1. Send payments:
// 	1.1 Send fake payment before sending normal one.
//
// 2. Track number of pending payments.
//
// 3. Tracking good / bad payment attempts:
//	3.1 For every node / channel track fail which were associated with it as
//	well as success payments.
//	3.2 Use internal good / bad metrics to filter suggested by lnd routes.
//
type Config struct {
	// Client...
	Client lightning.Client

	// MetricsBackend...
	Metrics crypto.MetricsBackend

	// Net...
	Net string
}

func (c Config) validate() error {
	if c.Client == nil {
		return errors.Errorf("client should be specified")
	}

	if c.Metrics == nil {
		return errors.Errorf("metrics backend should be specified")
	}

	if c.Net == "" {
		return errors.Errorf("net should be specified")
	}

	return nil
}

type Router struct {
	cfg Config
}

func NewRouter(cfg Config) (*Router, error) {
	if err := cfg.validate(); err != nil {
		return nil, errors.Errorf("config validate failed: %v", err)
	}

	return &Router{
		cfg: cfg,
	}, nil
}

func (r *Router) SendPayment(invoiceStr string, inputAmountSat btcutil.Amount) (
	*lightning.Payment, error) {
	// 1. Decode invoice and get public key.
	// 2. Ask lightning client to give you routes to destination node.
	//	What kind of information we could give additionally here which
	//	might help router to make decision?
	//  1. Percentage of time node being online.
	//  2.
	// 3. Query routes to the given node.
	// 4. Fetch success / fail rates for nodes in the route.
	// 5. Make attempts to do fail payment until success.
	// 6.

	//var nodeID lightning.NodeID
	//var paymentHash lightning.PaymentHash
	//var payment *lightning.Payment
	//
	//routes, err := r.cfg.Client.QueryRoutes(nodeID, inputAmountSat, 2)
	//if err != nil {
	//	return nil, errors.Errorf("unable query routes: %v", err)
	//}
	//
	//// TODO(andrew.shvv) Sort routes based on:
	//// 1. Previous history of attempts
	//// 2. Success / failure rate of nodes/channels
	//
	//var chosenRoute *lightning.Route
	//for _, route := range routes {
	//	fakePaymentHash := generateRandomPaymentHash()
	//	// TODO(andrew.shvv) Generate fake random payment hash
	//	// TODO(andrew.shvv) Change route to send minimal amount,
	//	//  so that payment in transition would now affect as a lot
	//
	//	err := r.cfg.Client.SendPaymentToRoute(route, fakePaymentHash)
	//	if err != nil {
	//		spew.Dump(err)
	//		return nil, err
	//		// Identify type of error, and act accordingly
	//		// IF cltv error - try again
	//		// IF temporary error - try again
	//		// IF unknown hash - than exit loop and use this route
	//		// IF payment in transition - than payment has failed
	//
	//		// Save payment attempt
	//		//
	//	}
	//}
	//
	//if chosenRoute == nil {
	//	// route hasn't been found - failure
	//	// Change payment status in fail
	//	return nil, errors.New("unable to find route")
	//}
	//
	//for {
	//	err := r.cfg.Client.SendPaymentToRoute(chosenRoute, paymentHash)
	//	if err != nil {
	//		spew.Dump(err)
	//		return nil, err
	//		// Identify type of error, and act accordingly
	//		// IF cltv error - try again for 10 minutes
	//		// IF temporary error - try again for 10 minutes
	//		// IF payment in transition - than payment is pending
	//		// Save successful failed attempt
	//	}
	//
	//	// Change payment status to completed
	//	// Save successful payment attempt
	//	break
	//}
	//
	//m := crypto.NewMetric("BTC", common.GetFunctionName(),
	//	r.cfg.MetricsBackend)
	//defer m.Finish()
	//
	//// Check that invoice is valid, and that amount which we are sending is
	//// corresponding to what we expect.
	//netParams, err := GetParams(r.cfg.Net)
	//if err != nil {
	//	m.AddError(metrics.HighSeverity)
	//	return nil, err
	//}
	//
	//invoice, err := zpay32.Decode(invoiceStr, netParams)
	//if err != nil {
	//	m.AddError(metrics.LowSeverity)
	//	return nil, err
	//}
	//
	//// If amount wasn't specified during invoice creation that amount field
	//// will be equal to nil.
	//var invoiceAmountSat int64
	//if invoice.MilliSat != nil {
	//	invoiceAmountSat = int64(invoice.MilliSat.ToSatoshis())
	//}
	//
	//var amountToSendSat int64
	//if invoiceAmountSat != 0 && inputAmountSat == 0 {
	//	// User hasn't specified amount, but in encoded in the invoice.
	//	amountToSendSat = invoiceAmountSat
	//
	//} else if invoiceAmountSat == 0 && inputAmountSat != 0 {
	//	// Amount is not encoded in the invoice, which means that service could
	//	// send every amount which user has specified.
	//	amountToSendSat = int64(inputAmountSat)
	//
	//} else if invoiceAmountSat != 0 && inputAmountSat != 0 {
	//	// If both amounts are specified that we should check that they are
	//	// equal.
	//	if int64(inputAmountSat) != invoiceAmountSat {
	//		m.AddError(metrics.LowSeverity)
	//		return nil, errors.Errorf("amount are not equal: invoice amount("+
	//			"%v), and input amount(%v)", btcutil.Amount(inputAmountSat),
	//			btcutil.Amount(invoiceAmountSat))
	//	}
	//	amountToSendSat = int64(inputAmountSat)
	//
	//} else {
	//	m.AddError(metrics.LowSeverity)
	//	return nil, errors.Errorf("invoice and user amount are not specified")
	//}
	//
	//var mediaFee decimal.Decimal
	//paymentHash := lightning.PaymentHash(hex.EncodeToString(invoice.PaymentHash[:]))
	//receiverNodeAddr := hex.EncodeToString(invoice.Destination.
	//	SerializeCompressed())
	//
	//if receiverNodeAddr == r.nodeAddr {
	//	// If we try to send payment to ourselves, than lightning network daemon
	//	// will fail, for that reason we handle this and pretend as if payment
	//	// was actually has been made.
	//	payment := &connectors.Payment{
	//		PaymentID: generatePaymentID(invoiceStr, connectors.Incoming),
	//		UpdatedAt: connectors.NowInMilliSeconds(),
	//		Status:    connectors.Completed,
	//		Direction: connectors.Incoming,
	//		System:    connectors.External,
	//		Receipt:   invoiceStr,
	//		Asset:     connectors.BTC,
	//		Media:     connectors.Lightning,
	//		Amount:    sat2DecAmount(btcutil.Amount(amountToSendSat)),
	//		MediaFee:  mediaFee,
	//		MediaID:   paymentHash,
	//	}
	//
	//	if err := c.cfg.PaymentStore.SavePayment(payment); err != nil {
	//		m.AddError(metrics.HighSeverity)
	//		return nil, errors.Errorf("unable add payment in store: %v", err)
	//	}
	//} else {
	//	// Send payment to the recipient and wait for it to be received.
	//	req := &lnrpc.SendRequest{
	//		Amt:            amountToSendSat,
	//		PaymentRequest: invoiceStr,
	//		FeeLimit: &lnrpc.FeeLimit{
	//			Limit: &lnrpc.FeeLimit_Percent{
	//				Percent: 3,
	//			},
	//		},
	//	}
	//
	//	// TODO(andrew.shvv) Use async version and return waiting payment after
	//	// 3-5 seconds.
	//	resp, err := c.cfg.Client.SendPaymentSync(context.Background(), req)
	//	if err != nil {
	//		m.AddError(metrics.HighSeverity)
	//		return nil, errors.Errorf("unable to send payment: %v", err)
	//	}
	//
	//	if resp.PaymentError != "" {
	//		m.AddError(metrics.HighSeverity)
	//		return nil, errors.Errorf("unable to send payment: %v", resp.PaymentError)
	//	}
	//
	//	mediaFee = sat2DecAmount(btcutil.Amount(resp.PaymentRoute.TotalFees))
	//	c.averageFee = c.averageFee.Add(mediaFee).Div(decimal.NewFromFloat(2.0))
	//}
	//
	//payment := &connectors.Payment{
	//	PaymentID: generatePaymentID(invoiceStr, connectors.Outgoing),
	//	UpdatedAt: connectors.NowInMilliSeconds(),
	//	Status:    connectors.Completed,
	//	System:    connectors.External,
	//	Direction: connectors.Outgoing,
	//	Receipt:   invoiceStr,
	//	Asset:     connectors.BTC,
	//	Media:     connectors.Lightning,
	//	Amount:    sat2DecAmount(btcutil.Amount(amountToSendSat)),
	//	MediaFee:  mediaFee,
	//	MediaID:   paymentHash,
	//}
	//
	//if err := c.cfg.PaymentStore.SavePayment(payment); err != nil {
	//	m.AddError(metrics.HighSeverity)
	//	return nil, errors.Errorf("unable add payment in store: %v", err)
	//}
	//
	//log.Infof("Send payment %v", spew.Sdump(payment))
	//
	//return payment, nil
	//
	//return payment, nil
	return nil, errors.Errorf("unsupported operation")
}
