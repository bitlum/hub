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

// SetState is used to receive equilibrium state from third-party optimisation
// subsystem and depending on optimisation strategy make changes in the
// topology of the router.
//
// NOTE: Part of the ManagerServer interface.
func (h *Hub) SetState(_ context.Context, req *SetStateRequest) (
	*SetStateResponse, error) {

	currentNetwork, err := h.router.Network()
	if err != nil {
		return nil, errors.Errorf("unable to get router topology: %v", err)
	}

	equilibriumNetwork := make([]*router.Channel, len(req.Channels))
	for i, c := range req.Channels {
		equilibriumNetwork[i] = &router.Channel{
			ChannelID:     router.ChannelID(c.ChannelId),
			UserID:        router.UserID(c.UserId),
			RouterBalance: router.ChannelUnit(c.RouterBalance),
			UserBalance:   0,
		}
	}

	actions := h.strategy.GenerateActions(currentNetwork, equilibriumNetwork)
	for _, changeState := range actions {
		if err := changeState(h.router); err != nil {
			return nil, errors.Errorf("unable to apply change state "+
				"function to the router: %v", err)

		}
	}

	return nil, nil
}
