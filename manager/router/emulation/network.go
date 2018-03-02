package emulation

import (
	"github.com/bitlum/hub/manager/router"

	"golang.org/x/net/context"
)

// emulationNetwork is used to emulate activity of local lightning network.
// This structure is an implementation of gRPC service, it was done in order to
// be able to emulate activity by third-party subsystems. external third-party
// commands.
type emulationNetwork struct {
	channels     map[router.ChannelID]*router.Channel
	users        map[router.UserID]*router.Channel
	updates      chan interface{}
	channelIndex uint64
	router       *routerEmulation
}

// Runtime check that routerEmulation implements EmulatorServer interface.
var _ EmulatorServer = (*emulationNetwork)(nil)

// SendPayment is used to emulate the activity of one user sending payment to
// another within the local router network.
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) SendPayment(_ context.Context, req *SendPaymentRequest) (
	*SendPaymentResponse, error) {

	var paymentFailed bool

	if req.FirstUser != 0 {
		channel := n.users[router.UserID(req.FirstUser)]
		channel.UserBalance -= req.Amount
		channel.RouterBalance += req.Amount
		defer func() {
			if paymentFailed {
				channel.UserBalance += req.Amount
				channel.RouterBalance -= req.Amount
			}
		}()

		if channel.UserBalance < 0 {
			paymentFailed = true
			return nil, nil
		}
	}

	if req.SecondUser != 0 {
		channel := n.users[router.UserID(req.SecondUser)]
		channel.RouterBalance -= req.Amount - req.Fee
		channel.UserBalance += req.Amount - req.Fee
		defer func() {
			if paymentFailed {
				channel.RouterBalance += req.Amount - req.Fee
				channel.UserBalance -= req.Amount - req.Fee
			}
		}()

		if channel.RouterBalance < 0 {
			paymentFailed = true
			return nil, nil
		}
	}

	n.updates <- router.UpdatePayment{
		// TODO(andrew.shvv) Populate with data
	}

	return &SendPaymentResponse{}, nil
}

// OpenChannel is used to emulate that user has opened the channel with the
// router.
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) OpenChannel(_ context.Context, req *OpenChannelRequest) (
	*OpenChannelResponse, error) {

	n.channelIndex++
	chanID := router.ChannelID(n.channelIndex)
	userID := router.UserID(req.UserId)

	c := &router.Channel{
		ChannelID:     chanID,
		UserID:        userID,
		UserBalance:   req.LockedByUser,
		RouterBalance: 0,
	}

	n.users[userID] = c
	n.channels[chanID] = c

	n.updates <- router.UpdateChannelOpened{
		// TODO(andrew.shvv) Populate with data
	}

	return &OpenChannelResponse{}, nil
}

// CloseChannel is used to emulate that user has closed the channel with the
// router.
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) CloseChannel(_ context.Context, req *CloseChannelRequest) (
	*CloseChannelResponse, error) {
	chanID := router.ChannelID(req.ChanId)
	c := n.channels[chanID]

	delete(n.channels, chanID)
	delete(n.users, c.UserID)

	n.updates <- router.UpdateChannelClosed{
		// TODO(andrew.shvv) Populate with data
	}

	// TODO(andrew.shvv) better persist layer separation
	n.router.freeBalance += c.RouterBalance

	return nil, nil
}
