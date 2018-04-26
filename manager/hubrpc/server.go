package hubrpc

import (
	"golang.org/x/net/context"
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
)

// Hub is an implementation of gRPC server which receive the message from
// external optimisation subsystem and apply those changes to the local
// router accordingly with initialised re-balancing strategy.
type Hub struct {
	router   router.Router
	strategy router.RouterStateStrategy
}

// NewHub creates new instance of the Hub.
func NewHub(r router.Router, s router.RouterStateStrategy) *Hub {
	return &Hub{
		router:   r,
		strategy: s,
	}
}

// Runtime check that Hub implements the hubrpc.ManagerServer interface.
var _ ManagerServer = (*Hub)(nil)

// UpdateLink is used to update router link in accordance with givein
// request. Link might just one channel, or might be the set of
// channels between user and router. This hook is used by third-parties
// to put new equilibrium state.
func (h *Hub) UpdateLink(_ context.Context,
	req *UpdateLinkRequest) (*UpdateLinkResponse, error) {

	currentNetwork, err := h.router.Network()
	if err != nil {
		return nil, errors.Errorf("unable to get router topology: %v", err)
	}

	// TODO(andrew.shvv) Remove that because we switched to the UpdateChannel
	equilibriumNetwork := make([]*router.Channel, len(currentNetwork))
	for i, c := range currentNetwork {

		// TODO(andrew.shvv) Add work with multiple channels
		if c.UserID == router.UserID(req.UserId) {
			equilibriumNetwork[i] = &router.Channel{
				ChannelID:     router.ChannelID(c.ChannelID),
				UserID:        router.UserID(req.UserId),
				RouterBalance: router.BalanceUnit(req.RouterBalance),
				UserBalance:   router.BalanceUnit(c.UserBalance),
			}
		} else {
			equilibriumNetwork[i] = c
		}
	}

	actions := h.strategy.GenerateActions(currentNetwork, equilibriumNetwork)
	for _, changeState := range actions {
		if err := changeState(h.router); err != nil {
			return nil, errors.Errorf("unable to apply change state "+
				"function to the router: %v", err)

		}
	}

	return &UpdateLinkResponse{}, nil
}
