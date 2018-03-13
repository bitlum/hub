package hubrpc

import (
	"github.com/bitlum/hub/manager/hubrpc"
	"golang.org/x/net/context"
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
)

// Hub is an implementation of gRPC server which receive the message from
// external optimisation subsystem and apply those changes to the local
// router accordingly with initialised re-balancing strategy.
type Hub struct {
	router   router.Router
	strategy router.RebalancingStrategy
}

// NewHub creates new instance of the Hub.
func NewHub(r router.Router, s router.RebalancingStrategy) *Hub {
	return &Hub{
		router:   r,
		strategy: s,
	}
}

// Runtime check that Hub implements the hubrpc.ManagerServer interface.
var _ hubrpc.ManagerServer = (*Hub)(nil)

// SetState is used to receive equilibrium state from third-party optimisation
// subsystem and depending on optimisation strategy make changes in the
// topology of the router.
//
// NOTE: Part of the hubrpc.ManagerServer interface.
func (h *Hub) SetState(_ context.Context, req *hubrpc.SetStateRequest) (
	*hubrpc.SetStateResponse, error) {

	currentNetwork, err := h.router.Network()
	if err != nil {
		return nil, errors.Errorf("unable to get router topology: %v", err)
	}

	equilibriumNetwork := make([]*router.Channel, len(req.Channels))
	for i, c := range req.Channels {
		equilibriumNetwork[i] = &router.Channel{
			ChannelID:     router.ChannelID(c.ChannelId),
			UserID:        router.UserID(c.UserId),
			UserBalance:   0,
			RouterBalance: c.RouterBalance,
		}
	}

	actions := h.strategy.GenerateActions(currentNetwork, equilibriumNetwork)

	for _, a := range actions {
		switch a.(type) {
		// TODO(andrew.shvv) Add actions
		}
	}

	return nil, nil
}
