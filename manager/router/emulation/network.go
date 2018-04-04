package emulation

import (
	"github.com/bitlum/hub/manager/router"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"github.com/go-errors/errors"
	"sync"
	"time"
	"strconv"
)

// emulationNetwork is used to emulate activity of users in router local
// lightning network. This structure is an implementation of gRPC service, it
// was done in order to be able to emulate activity by third-party subsystems.
type emulationNetwork struct {
	sync.Mutex

	channels     map[router.ChannelID]*router.Channel
	users        map[router.UserID]*router.Channel
	broadcaster  router.Broadcaster
	channelIndex uint64
	grpcServer   *grpc.Server
	errChan      chan error
	router       *RouterEmulation

	blockNotifier   *blockNotifier
	blockGeneration time.Duration
}

func newEmulationNetwork(blockGeneration time.Duration) *emulationNetwork {
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	network := &emulationNetwork{
		channels:        make(map[router.ChannelID]*router.Channel),
		users:           make(map[router.UserID]*router.Channel),
		broadcaster:     router.NewBroadcaster(),
		errChan:         make(chan error),
		grpcServer:      grpcServer,
		blockNotifier:   newBlockNotifier(blockGeneration),
		blockGeneration: blockGeneration,
	}

	RegisterEmulatorServer(grpcServer, network)

	return network
}

// Runtime check that RouterEmulation implements EmulatorServer interface.
var _ EmulatorServer = (*emulationNetwork)(nil)

// start...
func (n *emulationNetwork) start(host, port string) {
	go func() {
		addr := net.JoinHostPort(host, port)
		log.Infof("Start listening on: %v", addr)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			fail(n.errChan, "gRPC server unable to listen on %s", addr)
			return
		}
		defer lis.Close()

		log.Infof("Start gRPC serving on: %v", addr)
		if err := n.grpcServer.Serve(lis); err != nil {
			fail(n.errChan, "gRPC server unable to serve on %s", addr)
			return
		}

		log.Infof("Stop gRPC serving on: %v", addr)
	}()

	go n.blockNotifier.Start()
}

// stop gracefully stops the emulate network.
func (n *emulationNetwork) stop() {
	n.grpcServer.Stop()
	n.blockNotifier.Stop()
	close(n.errChan)
}

// done is used to notify other subsystem that service stop working.
func (n *emulationNetwork) done() chan error {
	return n.errChan
}

// SendPayment is used to emulate the activity of one user sending payment to
// another within the local router network.
//
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) SendPayment(_ context.Context, req *SendPaymentRequest) (
	*SendPaymentResponse, error) {

	n.Lock()
	defer n.Unlock()

	var paymentFailed bool

	if req.Receiver == "" && req.Sender == "" {
		return nil, errors.Errorf("both receiver and sender are zero")
	}

	if req.Sender != "" {
		// TODO(andrew.shvv) add multiple channels support
		channel, ok := n.users[router.UserID(req.Sender)]
		if !ok {
			return nil, errors.Errorf("unable to find sender with %v id",
				req.Sender)
		} else if channel.IsPending {
			return nil, errors.Errorf("channel %v is locked",
				channel.ChannelID)
		}

		channel.UserBalance -= router.ChannelUnit(req.Amount)
		channel.RouterBalance += router.ChannelUnit(req.Amount)
		defer func() {
			if paymentFailed {
				channel.UserBalance += router.ChannelUnit(req.Amount)
				channel.RouterBalance -= router.ChannelUnit(req.Amount)
			}
		}()

		if channel.UserBalance < 0 {
			paymentFailed = true

			// In the case of real system such information wouldn't be
			// accessible to us, for that return error, emulating user wallet
			// experience.
			return nil, errors.New("insufficient user balance to " +
				"make a payment")
		}
	}

	if req.Receiver != "" {
		// TODO(andrew.shvv) add multiple channels support
		channel, ok := n.users[router.UserID(req.Receiver)]
		if !ok {
			return nil, errors.Errorf("unable to find receiver with %v id",
				req.Sender)
		} else if channel.IsPending {
			return nil, errors.Errorf("channel %v is locked",
				channel.ChannelID)
		}

		channel.RouterBalance -= router.ChannelUnit(req.Amount - n.router.fee)
		channel.UserBalance += router.ChannelUnit(req.Amount - n.router.fee)
		defer func() {
			if paymentFailed {
				channel.RouterBalance += router.ChannelUnit(req.Amount + n.router.fee)
				channel.UserBalance -= router.ChannelUnit(req.Amount + n.router.fee)
			}
		}()

		if channel.RouterBalance < 0 {
			paymentFailed = true

			// As far as this is emulation we shouldn't return error,
			// but instead notify another subsystem about error,
			// so that it might be written in log for example an later examined.
			n.broadcaster.Write(&router.UpdatePayment{
				Status:   router.InsufficientFunds,
				Sender:   router.UserID(req.Sender),
				Receiver: router.UserID(req.Receiver),
				Amount:   req.Amount,
			})

			return &SendPaymentResponse{}, nil
		}
	}

	n.broadcaster.Write(&router.UpdatePayment{
		Status:   router.Successful,
		Sender:   router.UserID(req.Sender),
		Receiver: router.UserID(req.Receiver),
		Amount:   req.Amount,

		// TODO(andrew.shvv) Add earned
		Earned: 0,
	})

	return &SendPaymentResponse{}, nil
}

// OpenChannel is used to emulate that user has opened the channel with the
// router.
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) OpenChannel(_ context.Context, req *OpenChannelRequest) (
	*OpenChannelResponse, error) {

	n.Lock()
	defer n.Unlock()

	n.channelIndex++
	chanID := router.ChannelID(n.channelIndex)
	userID := router.UserID(req.UserId)

	if _, ok := n.users[userID]; ok {
		// TODO(andrew.shvv) add multiple channels support
		return nil, errors.Errorf("multiple channels with the same " +
			"user id unsupported")
	}

	c := &router.Channel{
		ChannelID:     chanID,
		UserID:        userID,
		UserBalance:   router.ChannelUnit(req.LockedByUser),
		RouterBalance: 0,
		IsPending:     true,
	}

	n.users[userID] = c
	n.channels[chanID] = c

	n.broadcaster.Write(&router.UpdateChannelOpening{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   router.ChannelUnit(c.UserBalance),
		RouterBalance: router.ChannelUnit(c.RouterBalance),

		// TODO(andrew.shvv) Add work with fee
		Fee: 0,
	})

	log.Tracef("User(%v) opened channel(%v) with router", userID, chanID)

	// Subscribe on block notification and update channel when block is
	// generated.
	s, err := n.blockNotifier.Subscribe()
	if err != nil {
		return nil, errors.Errorf("unable to send update payment: %v", err)
	}

	// Channel is able to operate only after block is generated.
	// Send update that channel is opened only after it is unlocked.
	go func() {
		defer n.blockNotifier.RemoveSubscription(s)
		<-s.C

		n.Lock()
		defer n.Unlock()

		c.IsPending = false
		n.broadcaster.Write(&router.UpdateChannelOpened{
			UserID:        c.UserID,
			ChannelID:     c.ChannelID,
			UserBalance:   router.ChannelUnit(c.UserBalance),
			RouterBalance: router.ChannelUnit(c.RouterBalance),

			// TODO(andrew.shvv) Add work with fee
			Fee: 0,
		})

		log.Tracef("Channel(%v) with user(%v) unlocked", chanID, userID)
	}()

	return &OpenChannelResponse{
		ChannelId: strconv.FormatUint(n.channelIndex, 10),
	}, nil
}

// CloseChannel is used to emulate that user has closed the channel with the
// router.
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) CloseChannel(_ context.Context, req *CloseChannelRequest) (
	*CloseChannelResponse, error) {

	n.Lock()
	defer n.Unlock()

	chanID := router.ChannelID(req.ChannelId)
	channel, ok := n.channels[chanID]
	if !ok {
		return nil, errors.Errorf("unable to find the channel with %v id", chanID)
	} else if channel.IsPending {
		return nil, errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	// Increase the pending balance till block is generated.
	n.router.pendingBalance += channel.RouterBalance
	channel.IsPending = true
	n.broadcaster.Write(&router.UpdateChannelClosing{
		UserID:    channel.UserID,
		ChannelID: channel.ChannelID,

		// TODO(andrew.shvv) Add work with fee
		Fee: 0,
	})

	log.Tracef("User(%v) closed channel(%v)", channel.UserID, channel.ChannelID)

	// Subscribe on block notification and return funds when block is
	// generated.
	s, err := n.blockNotifier.Subscribe()
	if err != nil {
		return nil, errors.Errorf("unable to increase router free "+
			"balance: %v", err)
	}

	// Update router free balance only after block is mined and increase
	// router balance on amount which we locked on our side in this channel.
	go func() {
		defer n.blockNotifier.RemoveSubscription(s)
		<-s.C

		n.Lock()
		defer n.Unlock()

		n.broadcaster.Write(&router.UpdateChannelClosed{
			UserID:    channel.UserID,
			ChannelID: channel.ChannelID,

			// TODO(andrew.shvv) Add work with fee
			Fee: 0,
		})

		delete(n.channels, chanID)
		delete(n.users, channel.UserID)

		n.router.pendingBalance -= channel.RouterBalance
		n.router.freeBalance += channel.RouterBalance

		log.Tracef("Router received %v money previously locked in"+
			" channel(%v)", channel.RouterBalance, channel.ChannelID)
	}()

	return &CloseChannelResponse{}, nil
}

// SetBlockGenDuration is used to set the time which is needed for blokc
// to be generated time. This would impact channel creation, channel
// update and channel close.
func (n *emulationNetwork) SetBlockGenDuration(_ context.Context,
	req *SetBlockGenDurationRequest) (*SetBlockGenDurationResponse, error) {

	n.Lock()
	defer n.Unlock()

	n.blockGeneration = time.Duration(req.Duration)
	d := time.Millisecond * n.blockGeneration
	if err := n.blockNotifier.SetBlockGenDuration(d); err != nil {
		return nil, errors.Errorf("unable set block generation duration: %v", err)
	}

	return &SetBlockGenDurationResponse{}, nil
}
