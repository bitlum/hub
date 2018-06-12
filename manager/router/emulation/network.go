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
	"github.com/bitlum/hub/manager/common/broadcast"
)

// emulationNetwork is used to emulate activity of users in router local
// lightning network. This structure is an implementation of gRPC service, it
// was done in order to be able to emulate activity by third-party subsystems.
type emulationNetwork struct {
	sync.Mutex

	channels     map[router.ChannelID]*router.Channel
	users        map[router.UserID]*router.Channel
	broadcaster  *broadcast.Broadcaster
	channelIndex uint64
	grpcServer   *grpc.Server
	errChan      chan error
	router       *RouterEmulation

	blockNotifier   *blockNotifier
	blockGeneration time.Duration
	blockchainFee   int64
}

func newEmulationNetwork(blockGeneration time.Duration) *emulationNetwork {
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	network := &emulationNetwork{
		channels:        make(map[router.ChannelID]*router.Channel),
		users:           make(map[router.UserID]*router.Channel),
		broadcaster:     broadcast.NewBroadcaster(),
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

	if req.Receiver == "" || req.Sender == "" {
		return nil, errors.Errorf("receiver or sender are empty")
	}

	if req.Receiver == req.Sender {
		return nil, errors.Errorf("receiver and sender are equal")
	}

	var incomingAmount int64
	var outgoingAmount int64
	var routerFee int64
	var transferedAmount int64

	if req.Sender == "0" {
		// In this case sender is router (outgoing payment), so incoming
		// amount - amount which is received from user is zero.
		incomingAmount = 0
		outgoingAmount = req.Amount
		transferedAmount = outgoingAmount
	} else if req.Receiver == "0" {
		// In this case receiver is router (incoming payment), so outgoing
		// amount - amount which is send from router to user is zero.
		incomingAmount = req.Amount
		outgoingAmount = 0
		transferedAmount = incomingAmount
	} else {
		// In the case sender and receiver are users (forward payment).
		// Calculate router fee which it takes for making the forwarding payment.
		routerFee = calculateForwardingFee(req.Amount, n.router.feeBase,
			n.router.feeProportion)
		incomingAmount = req.Amount
		outgoingAmount = req.Amount - routerFee
		transferedAmount = outgoingAmount

		if outgoingAmount <= 0 {
			return nil, errors.Errorf("fee is greater than amount")
		}
	}

	if req.Sender != "0" {
		// TODO(andrew.shvv) add multiple channels support
		channel, ok := n.users[router.UserID(req.Sender)]
		if !ok {
			paymentFailed = true
			return nil, errors.Errorf("unable to find sender with %v id",
				req.Sender)
		} else if channel.IsPending {
			paymentFailed = true
			return nil, errors.Errorf("channel %v is locked",
				channel.ChannelID)
		}

		channel.UserBalance -= router.BalanceUnit(incomingAmount)
		channel.RouterBalance += router.BalanceUnit(incomingAmount)
		defer func() {
			if paymentFailed {
				channel.UserBalance += router.BalanceUnit(incomingAmount)
				channel.RouterBalance -= router.BalanceUnit(incomingAmount)
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

	if req.Receiver != "0" {
		// TODO(andrew.shvv) add multiple channels support
		channel, ok := n.users[router.UserID(req.Receiver)]
		if !ok {
			paymentFailed = true
			return nil, errors.Errorf("unable to find receiver with %v id",
				req.Receiver)
		} else if channel.IsPending {
			paymentFailed = true
			return nil, errors.Errorf("channel %v is locked",
				channel.ChannelID)
		}

		channel.RouterBalance -= router.BalanceUnit(outgoingAmount)
		channel.UserBalance += router.BalanceUnit(outgoingAmount)
		defer func() {
			if paymentFailed {
				channel.RouterBalance += router.BalanceUnit(outgoingAmount)
				channel.UserBalance -= router.BalanceUnit(outgoingAmount)
			}
		}()

		if channel.RouterBalance < 0 {
			paymentFailed = true

			// As far as this is emulation we shouldn't return error,
			// but instead notify another subsystem about error,
			// so that it might be written in log for example an later examined.
			n.broadcaster.Write(&router.UpdatePayment{
				Type:     router.Incoming,
				Status:   router.InsufficientFunds,
				Sender:   router.UserID(req.Sender),
				Receiver: router.UserID(req.Receiver),
				Amount:   router.BalanceUnit(transferedAmount),
			})

			return &SendPaymentResponse{}, nil
		}
	}

	n.broadcaster.Write(&router.UpdatePayment{
		Type:     router.Incoming,
		Status:   router.Successful,
		Sender:   router.UserID(req.Sender),
		Receiver: router.UserID(req.Receiver),
		Amount:   router.BalanceUnit(transferedAmount),
		Earned:   router.BalanceUnit(routerFee),
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
	id := strconv.FormatUint(n.channelIndex, 10)
	chanID := router.ChannelID(id)
	userID := router.UserID(req.UserId)

	if _, ok := n.users[userID]; ok {
		// TODO(andrew.shvv) add multiple channels support
		return nil, errors.Errorf("multiple channels with the same " +
			"user id unsupported")
	}

	// Ensure that initiator has enough funds to open and close the channel.
	if req.LockedByUser-n.blockchainFee <= 0 {
		return nil, errors.Errorf("user balance is not sufficient to "+
			"open the channel, need(%v)", n.blockchainFee)
	} else if req.LockedByUser-2*n.blockchainFee <= 0 {
		return nil, errors.Errorf("user balance is not sufficient to "+
			"close the channel after opening, need(%v)", 2*n.blockchainFee)
	}

	// Take fee for opening and closing the channel, from channel initiator,
	// and save close fee so that we could use it later for paying the
	// blockchain.
	openChannelFee := router.BalanceUnit(n.blockchainFee)
	closeChannelFee := router.BalanceUnit(n.blockchainFee)
	userBalance := router.BalanceUnit(req.LockedByUser) - openChannelFee - closeChannelFee
	c := &router.Channel{
		ChannelID:     chanID,
		UserID:        userID,
		UserBalance:   router.BalanceUnit(userBalance),
		RouterBalance: 0,
		IsPending:     true,
		IsActive:      false,
		Initiator:     router.UserInitiator,
		CloseFee:      closeChannelFee,
	}

	n.users[userID] = c
	n.channels[chanID] = c

	n.broadcaster.Write(&router.UpdateChannelOpening{
		UserID:        c.UserID,
		ChannelID:     c.ChannelID,
		UserBalance:   router.BalanceUnit(c.UserBalance),
		RouterBalance: router.BalanceUnit(c.RouterBalance),
		Fee:           openChannelFee,
	})

	log.Tracef("User(%v) opened channel(%v) with router", userID, chanID)

	// Subscribe on block notification and update channel when block is
	// generated.
	l := n.blockNotifier.Subscribe()

	// Channel is able to operate only after block is generated.
	// Send update that channel is opened only after it is unlocked.
	start := time.Now()
	go func() {
		defer l.Stop()
		<-l.Read()

		n.Lock()
		defer n.Unlock()

		c.IsPending = false
		n.broadcaster.Write(&router.UpdateChannelOpened{
			UserID:        c.UserID,
			ChannelID:     c.ChannelID,
			UserBalance:   router.BalanceUnit(c.UserBalance),
			RouterBalance: router.BalanceUnit(c.RouterBalance),
			Fee:           openChannelFee,
			Duration:      time.Now().Sub(start),
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
		Fee:       channel.CloseFee,
	})

	log.Tracef("User(%v) closed channel(%v)", channel.UserID, channel.ChannelID)

	// Subscribe on block notification and return funds when block is
	// generated.
	l := n.blockNotifier.Subscribe()

	// Update router free balance only after block is mined and increase
	// router balance on amount which we locked on our side in this channel.
	start := time.Now()
	go func() {
		defer l.Stop()
		<-l.Read()

		n.Lock()
		defer n.Unlock()

		n.broadcaster.Write(&router.UpdateChannelClosed{
			UserID:    channel.UserID,
			ChannelID: channel.ChannelID,
			Fee:       channel.CloseFee,
			Duration:  time.Now().Sub(start),
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

	n.blockGeneration = time.Millisecond * time.Duration(req.Duration)

	if err := n.blockNotifier.SetBlockGenDuration(n.blockGeneration); err != nil {
		return nil, errors.Errorf("unable set block generation duration: %v", err)
	}

	return &SetBlockGenDurationResponse{}, nil
}

//
// SetBlockchainFee is used to set the fee which blockchain takes for
// making an computation, transaction creation, i.e. channel updates.
func (n *emulationNetwork) SetBlockchainFee(_ context.Context,
	req *SetBlockchainFeeRequest) (*SetBlockchainFeeResponse, error) {
	n.Lock()
	defer n.Unlock()

	n.blockchainFee = req.Fee
	return &SetBlockchainFeeResponse{}, nil
}
