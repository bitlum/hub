package hubrpc

import (
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/optimisation"
	"github.com/go-errors/errors"
	"golang.org/x/net/context"
)

// Hub is an implementation of gRPC server which receive the message from
// external optimisation subsystem and apply those changes to the local
// lightning node accordingly with initialised re-balancing strategy.
type Hub struct {
	client   lightning.Client
	strategy optimisation.NodeStateStrategy
}

// NewHub creates new instance of the Hub.
func NewHub(client lightning.Client, s optimisation.NodeStateStrategy) *Hub {
	return &Hub{
		client:   client,
		strategy: s,
	}
}

// Runtime check that Hub implements the hubrpc.ManagerServer interface.
var _ ManagerServer = (*Hub)(nil)

// UpdateLink is used to update client link in accordance with givein
// request. Link might just one channel, or might be the set of
// channels between user and lightning. This hook is used by third-parties
// to put new equilibrium state.
func (h *Hub) UpdateLink(_ context.Context,
	req *UpdateLinkRequest) (*UpdateLinkResponse, error) {

	currentNetwork, err := h.client.Channels()
	if err != nil {
		return nil, errors.Errorf("unable to get client topology: %v", err)
	}

	// TODO(andrew.shvv) Remove that because we switched to the UpdateChannel
	equilibriumNetwork := make([]*lightning.Channel, len(currentNetwork))
	for i, c := range currentNetwork {

		// TODO(andrew.shvv) Add work with multiple channels
		if c.UserID == lightning.UserID(req.UserId) {
			equilibriumNetwork[i] = &lightning.Channel{
				ChannelID:     lightning.ChannelID(c.ChannelID),
				UserID:        lightning.UserID(req.UserId),
				LocalBalance:  lightning.BalanceUnit(req.LocalBalance),
				RemoteBalance: lightning.BalanceUnit(c.RemoteBalance),
			}
		} else {
			equilibriumNetwork[i] = c
		}
	}

	actions := h.strategy.GenerateActions(currentNetwork, equilibriumNetwork)
	for _, changeState := range actions {
		if err := changeState(h.client); err != nil {
			return nil, errors.Errorf("unable to apply change state "+
				"function to the client: %v", err)

		}
	}

	return &UpdateLinkResponse{}, nil
}

//
// SetPaymentFeeBase sets base number of milli units (i.e milli satoshis in
// Bitcoin) which will be taken for every payment forwarding.
func (h *Hub) SetPaymentFeeBase(_ context.Context,
	req *SetPaymentFeeBaseRequest) (*SetPaymentFeeBaseResponse, error) {
	err := h.client.SetFeeBase(req.PaymentFeeBase)
	return &SetPaymentFeeBaseResponse{}, err
}

//
// SetPaymentFeeProportional sets the number of milli units (i.e milli
// satoshis in Bitcoin) which will be taken for every killounit of
// payment amount.
func (h *Hub) SetPaymentFeeProportional(_ context.Context,
	req *SetPaymentFeeProportionalRequest) (*SetPaymentFeeProportionalResponse,
	error) {
	err := h.client.SetFeeProportional(req.PaymentFeeProportional)
	return &SetPaymentFeeProportionalResponse{}, err
}
