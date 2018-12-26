package hubrpc

import (
	"encoding/hex"
	"github.com/bitlum/hub/common"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/rpc"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	"github.com/shopspring/decimal"
	"golang.org/x/net/context"
	"math/rand"
)

type Config struct {
	// Client...
	Client lightning.Client

	// MetricsBackend...
	MetricsBackend rpc.MetricsBackend
}

// Hub is an implementation of gRPC server which receive the message from
// external optimisation subsystem and apply those changes to the local
// lightning node accordingly with initialised re-balancing strategy.
type Hub struct {
	cfg *Config
}

// NewHub creates new instance of the Hub.
func NewHub(cfg *Config) *Hub {
	return &Hub{
		cfg: cfg,
	}
}

// Runtime check that Hub implements the hubrpc.ManagerServer interface.
var _ HubServer = (*Hub)(nil)

//
// CreateInvoice is used to create lightning network invoice in which
// will be used to receive money from external lightning network entity.
func (h *Hub) CreateInvoice(ctx context.Context,
	req *CreateInvoiceRequest) (*CreateInvoiceResponse, error) {

	m := rpc.NewMetric(common.GetFunctionName(), h.cfg.MetricsBackend)
	defer m.Finish()

	requestID := rand.Int()

	log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(req))

	var resp *CreateInvoiceResponse

	// Ensure that even if amount is not specified we treat it as zero
	// value.
	if req.Amount == "" {
		req.Amount = "0"
	}

	amountSat, err := common.BtcStrToSatoshi(req.Amount)
	if err != nil {
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	paymentRequest, invoice, err := h.cfg.Client.CreateInvoice("bitlum",
		btcutil.Amount(amountSat), req.Description)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	resp = &CreateInvoiceResponse{
		CreationDate: common.ConvertTimeToMilliSeconds(invoice.Timestamp),
		Expiry:       common.ConvertDurationToMilliSeconds(invoice.Expiry()),
		Invoice:      paymentRequest,
	}

	log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(resp))

	return resp, nil
}

//
// ValidateInvoice is used to validate lightning invoice on belonging to
// proper network as well as amount inside invoice.
func (h *Hub) ValidateInvoice(ctx context.Context,
	req *ValidateInvoiceRequest) (*ValidateInvoiceResponse, error) {

	m := rpc.NewMetric(common.GetFunctionName(), h.cfg.MetricsBackend)
	defer m.Finish()

	requestID := rand.Int()

	log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(req))

	if req.Amount == "" {
		req.Amount = "0"
	}

	amountSat, err := common.BtcStrToSatoshi(req.Amount)
	if err != nil {
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	invoice, err := h.cfg.Client.ValidateInvoice(req.Invoice, btcutil.Amount(amountSat))
	if err != nil {
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	var description string
	if invoice.Description != nil {
		description = *invoice.Description
	}

	var invoiceAmount decimal.Decimal
	if invoice.MilliSat != nil {
		amountSat := invoice.MilliSat.ToSatoshis()
		invoiceAmount = common.Sat2DecAmount(btcutil.Amount(amountSat))
	}

	var fallbackAddress string
	if invoice.FallbackAddr != nil {
		fallbackAddress = invoice.FallbackAddr.String()
	}

	var destination string
	if invoice.Destination != nil {
		destination = hex.EncodeToString(invoice.Destination.SerializeCompressed())
	}

	resp := &ValidateInvoiceResponse{
		Invoice: &Invoice{
			Memo:         description,
			Value:        invoiceAmount.Round(8).String(),
			CreationDate: common.ConvertTimeToMilliSeconds(invoice.Timestamp),
			Expiry:       common.ConvertDurationToMilliSeconds(invoice.Expiry()),
			FallbackAddr: fallbackAddress,
			Destination:  destination,
		},
	}

	log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(resp))

	return resp, nil
}

//
// Balance is used to determine balance.
func (h *Hub) Balance(ctx context.Context, req *BalanceRequest) (*BalanceResponse,
	error) {

	m := rpc.NewMetric(common.GetFunctionName(), h.cfg.MetricsBackend)
	defer m.Finish()

	requestID := rand.Int()

	log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(req))

	resp := &BalanceResponse{}

	available, err := h.cfg.Client.AvailableBalance()
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	pending, err := h.cfg.Client.PendingBalance()
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	resp.Balances = append(resp.Balances, &Balance{
		Available: common.Sat2DecAmount(available).String(),
		Pending:   common.Sat2DecAmount(pending).String(),
	})

	log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(resp))

	return resp, nil
}

// EstimateFee estimates the fee of the outgoing payment.
func (h *Hub) EstimateFee(ctx context.Context,
	req *EstimateFeeRequest) (*EstimateFeeResponse, error) {

	m := rpc.NewMetric(common.GetFunctionName(), h.cfg.MetricsBackend)
	defer m.Finish()

	requestID := rand.Int()

	log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(req))

	var resp *EstimateFeeResponse

	amountSat, err := common.BtcStrToSatoshi(req.Amount)
	if err != nil {
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	fee, err := h.cfg.Client.EstimateFee(req.Invoice, btcutil.Amount(amountSat))
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	resp = &EstimateFeeResponse{
		MediaFee: fee.String(),
	}

	log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(resp))

	return resp, nil
}

// SendPayment sends payment to the given recipient, ensures in the validity
// of the receipt as well as the account has enough money for doing that.
func (h *Hub) SendPayment(ctx context.Context,
	req *SendPaymentRequest) (*Payment, error) {

	return nil, errors.Errorf("not implemented")
	//m := rpc.NewMetric(common.GetFunctionName(), h.cfg.MetricsBackend)
	//defer m.Finish()
	//
	//requestID := rand.Int()
	//
	//log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
	//	requestID, convertProtoMessage(req))
	//
	//var (
	//	resp    *Payment
	//	payment *lightning.Payment
	//	err     error
	//)
	//
	//if req.Amount == "" {
	//	req.Amount = "0"
	//}
	//
	//amountSat, err := common.BtcStrToSatoshi(req.Amount)
	//if err != nil {
	//	log.Errorf("command(%v), id(%v), error: %v",
	//		common.GetFunctionName(), requestID, err)
	//	m.AddError(metrics.LowSeverity)
	//	return nil, err
	//}
	//
	//payment, err = h.cfg.Client.SendPaymentToRoute(req.Invoice,
	//	btcutil.Amount(amountSat))
	//if err != nil {
	//	err := newErrInternal(err.Error())
	//	log.Errorf("command(%v), id(%v), error: %v", common.GetFunctionName(),
	//		requestID, err)
	//	m.AddError(metrics.LowSeverity)
	//	return nil, err
	//}
	//
	//resp, err = convertPaymentToProto(payment)
	//if err != nil {
	//	err := newErrInternal(err.Error())
	//	log.Errorf("command(%v), id(%v), error: %v",
	//		common.GetFunctionName(), requestID, err)
	//	m.AddError(metrics.LowSeverity)
	//	return nil, err
	//}
	//
	//log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
	//	requestID, convertProtoMessage(resp))
	//
	//return resp, nil
}

// PaymentByID is used to fetch the information about payment, by the
// given system payment id.
func (h *Hub) PaymentByID(ctx context.Context, req *PaymentByIDRequest) (*Payment,
	error) {
	return nil, errors.Errorf("endpoint is not implemented")

	///requestID := rand.Int()
	//
	//log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
	//	requestID, convertProtoMessage(req))
	//
	//payment, err := s.paymentsStore.PaymentByID(req.PaymentId)
	//if err != nil {
	//	err := newErrInternal(err.Error())
	//	log.Errorf("command(%v), id(%v), error: %v",
	//		common.GetFunctionName(), requestID, err)
	//	s.metrics.AddError(common.GetFunctionName(), string(metrics.LowSeverity))
	//	return nil, err
	//}
	//
	//resp, err := convertPaymentToProto(payment)
	//if err != nil {
	//	err := newErrInternal(err.Error())
	//	log.Errorf("command(%v), id(%v), error: %v",
	//		common.GetFunctionName(), requestID, err)
	//	s.metrics.AddError(common.GetFunctionName(), string(metrics.LowSeverity))
	//	return nil, err
	//}
	//
	//log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
	//	requestID, convertProtoMessage(resp))
	//
	//return resp, err
}

//
// PaymentByReceipt is used to fetch the information about payment, by the
// given receipt.
func (h *Hub) PaymentByInvoice(ctx context.Context,
	req *PaymentByInvoiceRequest) (*Payment, error) {

	m := rpc.NewMetric(common.GetFunctionName(), h.cfg.MetricsBackend)
	defer m.Finish()

	requestID := rand.Int()

	log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(req))

	payment, err := h.cfg.Client.PaymentByInvoice(req.Invoice)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	protoPayment, err := convertPaymentToProto(payment)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	resp := protoPayment

	log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(resp))

	return resp, nil
}

//
// ListPayments returns list of payment which were registered by the
// system.
func (h *Hub) ListPayments(ctx context.Context,
	req *ListPaymentsRequest) (*ListPaymentsResponse, error) {

	m := rpc.NewMetric(common.GetFunctionName(), h.cfg.MetricsBackend)
	defer m.Finish()

	requestID := rand.Int()

	log.Tracef("command(%v), id(%v), request(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(req))

	var (
		asset     string
		status    lightning.PaymentStatus
		direction lightning.PaymentDirection
		system    lightning.PaymentSystem
		err       error
	)

	if req.Direction != PaymentDirection_DIRECTION_NONE {
		direction, err = ConvertPaymentDirectionFromProto(req.Direction)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), id(%v), error: %v",
				common.GetFunctionName(), requestID, err)
			m.AddError(metrics.LowSeverity)
			return nil, err
		}
	}

	if req.System != PaymentSystem_SYSTEM_NONE {
		system, err = ConvertPaymentSystemFromProto(req.System)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), id(%v), error: %v",
				common.GetFunctionName(), requestID, err)
			m.AddError(metrics.LowSeverity)
			return nil, err
		}
	}

	if req.Status != PaymentStatus_STATUS_NONE {
		status, err = ConvertPaymentStatusFromProto(req.Status)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), id(%v), error: %v",
				common.GetFunctionName(), requestID, err)
			m.AddError(metrics.LowSeverity)
			return nil, err
		}
	}

	payments, err := h.cfg.Client.ListPayments(asset, status, direction, system)
	if err != nil {
		err := newErrInternal(err.Error())
		log.Errorf("command(%v), id(%v), error: %v",
			common.GetFunctionName(), requestID, err)
		m.AddError(metrics.LowSeverity)
		return nil, err
	}

	var protoPayments []*Payment
	for _, payment := range payments {
		protoPayment, err := convertPaymentToProto(payment)
		if err != nil {
			err := newErrInternal(err.Error())
			log.Errorf("command(%v), id(%v), error: %v",
				common.GetFunctionName(), requestID, err)
			m.AddError(metrics.LowSeverity)
			return nil, err
		}

		protoPayments = append(protoPayments, protoPayment)
	}

	resp := &ListPaymentsResponse{
		Payments: protoPayments,
	}

	log.Tracef("command(%v), id(%v), response(%v)", common.GetFunctionName(),
		requestID, convertProtoMessage(resp))

	return resp, nil
}

// NodesStats return statistical data about node, and sort nodes by
// internal ranking algorithm.
func (h *Hub) CheckNodeStats(ctx context.Context,
	req *CheckNodeStatsRequest) (*CheckNodeStatsResponse, error) {
	return nil, errors.Errorf("endpoint is no supported")
}
