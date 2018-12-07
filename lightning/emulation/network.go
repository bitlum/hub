package emulation

import (
	"github.com/bitlum/hub/lightning"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"github.com/go-errors/errors"
	"sync"
	"time"
	"strconv"
	"github.com/bitlum/hub/common/broadcast"
)

// emulationNetwork is used to emulate activity of users in client local
// lightning network. This structure is an implementation of gRPC service, it
// was done in order to be able to emulate activity by third-party subsystems.
type emulationNetwork struct {
	sync.Mutex

	channels     map[lightning.ChannelID]*lightning.Channel
	users        map[lightning.UserID]*lightning.Channel
	broadcaster  *broadcast.Broadcaster
	channelIndex uint64
	grpcServer   *grpc.Server
	errChan      chan error
	client       *Client

	blockNotifier   *blockNotifier
	blockGeneration time.Duration
	blockchainFee   int64
}

func newEmulationNetwork(blockGeneration time.Duration) *emulationNetwork {
	grpcServer := grpc.NewServer([]grpc.ServerOption{}...)
	network := &emulationNetwork{
		channels:        make(map[lightning.ChannelID]*lightning.Channel),
		users:           make(map[lightning.UserID]*lightning.Channel),
		broadcaster:     broadcast.NewBroadcaster(),
		errChan:         make(chan error),
		grpcServer:      grpcServer,
		blockNotifier:   newBlockNotifier(blockGeneration),
		blockGeneration: blockGeneration,
	}

	RegisterEmulatorServer(grpcServer, network)

	return network
}

// Runtime check that Client implements EmulatorServer interface.
var _ EmulatorServer = (*emulationNetwork)(nil)

// start...
func (n *emulationNetwork) start(host, port string) {
	go func() {
		addr := net.JoinHostPort(host, port)
		log.Infof("Start emulator network listening on: %v", addr)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			fail(n.errChan, "gRPC server unable to listen on %s", addr)
			return
		}
		defer lis.Close()

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
// another within the local client network.
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
	var nodeFee int64
	var transferedAmount int64

	if req.Sender == "0" {
		// In this case sender is client (outgoing payment), so incoming
		// amount - amount which is received from user is zero.
		incomingAmount = 0
		outgoingAmount = req.Amount
		transferedAmount = outgoingAmount
	} else if req.Receiver == "0" {
		// In this case receiver is client (incoming payment), so outgoing
		// amount - amount which is send from client to user is zero.
		incomingAmount = req.Amount
		outgoingAmount = 0
		transferedAmount = incomingAmount
	} else {
		// In the case sender and receiver are users (forward payment).
		// Calculate client fee which it takes for making the forwarding payment.
		nodeFee = calculateForwardingFee(req.Amount, n.client.feeBase,
			n.client.feeProportion)
		incomingAmount = req.Amount
		outgoingAmount = req.Amount - nodeFee
		transferedAmount = outgoingAmount

		if outgoingAmount <= 0 {
			n.broadcaster.Write(&lightning.UpdatePayment{
				ID:       req.Id,
				Type:     lightning.Incoming,
				Status:   lightning.UserLocalFail,
				Sender:   lightning.UserID(req.Sender),
				Receiver: lightning.UserID(req.Receiver),
				Amount:   lightning.BalanceUnit(transferedAmount),
			})

			return &SendPaymentResponse{}, errors.Errorf("fee is greater than amount")
		}
	}

	if req.Sender != "0" {
		// TODO(andrew.shvv) add multiple channels support
		channel, ok := n.users[lightning.UserID(req.Sender)]
		if !ok {
			paymentFailed = true

			n.broadcaster.Write(&lightning.UpdatePayment{
				ID:       req.Id,
				Type:     lightning.Incoming,
				Status:   lightning.UserLocalFail,
				Sender:   lightning.UserID(req.Sender),
				Receiver: lightning.UserID(req.Receiver),
				Amount:   lightning.BalanceUnit(transferedAmount),
			})

			return &SendPaymentResponse{}, errors.Errorf("unable to find sender with %v id",
				req.Sender)
		} else if channel.IsPending() {
			paymentFailed = true
			n.broadcaster.Write(&lightning.UpdatePayment{
				ID:       req.Id,
				Type:     lightning.Incoming,
				Status:   lightning.UserLocalFail,
				Sender:   lightning.UserID(req.Sender),
				Receiver: lightning.UserID(req.Receiver),
				Amount:   lightning.BalanceUnit(transferedAmount),
			})

			return &SendPaymentResponse{}, nil
		}

		channel.RemoteBalance -= lightning.BalanceUnit(incomingAmount)
		channel.LocalBalance += lightning.BalanceUnit(incomingAmount)
		defer func() {
			if paymentFailed {
				channel.RemoteBalance += lightning.BalanceUnit(incomingAmount)
				channel.LocalBalance -= lightning.BalanceUnit(incomingAmount)
			}
		}()

		if channel.RemoteBalance < 0 {
			paymentFailed = true

			// In the case of real system such information wouldn't be
			// accessible to us.
			n.broadcaster.Write(&lightning.UpdatePayment{
				ID:       req.Id,
				Type:     lightning.Incoming,
				Status:   lightning.UserLocalFail,
				Sender:   lightning.UserID(req.Sender),
				Receiver: lightning.UserID(req.Receiver),
				Amount:   lightning.BalanceUnit(transferedAmount),
			})

			return &SendPaymentResponse{}, nil
		}
	}

	if req.Receiver != "0" {
		// TODO(andrew.shvv) add multiple channels support
		channel, ok := n.users[lightning.UserID(req.Receiver)]
		if !ok {
			paymentFailed = true

			n.broadcaster.Write(&lightning.UpdatePayment{
				ID:       req.Id,
				Type:     lightning.Incoming,
				Status:   lightning.UserLocalFail,
				Sender:   lightning.UserID(req.Sender),
				Receiver: lightning.UserID(req.Receiver),
				Amount:   lightning.BalanceUnit(transferedAmount),
			})

			return nil, errors.Errorf("unable to find receiver with %v id",
				req.Receiver)
		} else if channel.IsPending() {
			paymentFailed = true

			n.broadcaster.Write(&lightning.UpdatePayment{
				ID:       req.Id,
				Type:     lightning.Incoming,
				Status:   lightning.UserLocalFail,
				Sender:   lightning.UserID(req.Sender),
				Receiver: lightning.UserID(req.Receiver),
				Amount:   lightning.BalanceUnit(transferedAmount),
			})

			return nil, errors.Errorf("channel %v is locked",
				channel.ChannelID)
		}

		channel.LocalBalance -= lightning.BalanceUnit(outgoingAmount)
		channel.RemoteBalance += lightning.BalanceUnit(outgoingAmount)
		defer func() {
			if paymentFailed {
				channel.LocalBalance += lightning.BalanceUnit(outgoingAmount)
				channel.RemoteBalance -= lightning.BalanceUnit(outgoingAmount)
			}
		}()

		if channel.LocalBalance < 0 {
			paymentFailed = true

			// As far as this is emulation we shouldn't return error,
			// but instead notify another subsystem about error,
			// so that it might be written in log for example an later examined.
			n.broadcaster.Write(&lightning.UpdatePayment{
				ID:       req.Id,
				Type:     lightning.Incoming,
				Status:   lightning.InsufficientFunds,
				Sender:   lightning.UserID(req.Sender),
				Receiver: lightning.UserID(req.Receiver),
				Amount:   lightning.BalanceUnit(transferedAmount),
			})

			return &SendPaymentResponse{}, nil
		}
	}

	n.broadcaster.Write(&lightning.UpdatePayment{
		ID:       req.Id,
		Type:     lightning.Incoming,
		Status:   lightning.Successful,
		Sender:   lightning.UserID(req.Sender),
		Receiver: lightning.UserID(req.Receiver),
		Amount:   lightning.BalanceUnit(transferedAmount),
		Earned:   lightning.BalanceUnit(nodeFee),
	})

	return &SendPaymentResponse{}, nil
}

// OpenChannel is used to emulate that user has opened the channel with the
// lightning.
//
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) OpenChannel(_ context.Context, req *OpenChannelRequest) (
	*OpenChannelResponse, error) {

	n.Lock()
	defer n.Unlock()

	n.channelIndex++
	id := strconv.FormatUint(n.channelIndex, 10)
	chanID := lightning.ChannelID(id)
	userID := lightning.UserID(req.UserId)

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
	openChannelFee := lightning.BalanceUnit(n.blockchainFee)
	closeChannelFee := lightning.BalanceUnit(n.blockchainFee)
	userBalance := lightning.BalanceUnit(req.LockedByUser) - openChannelFee - closeChannelFee
	fundingAmount := lightning.BalanceUnit(req.LockedByUser)

	cfg := &lightning.ChannelConfig{
		Broadcaster: n.broadcaster,
		Storage:     &StubChannelStorage{},
	}

	channel, err := lightning.NewChannel(chanID, userID, fundingAmount,
		userBalance, 0, closeChannelFee, lightning.RemoteInitiator, cfg)
	if err != nil {
		return nil, errors.Errorf("unable create channel: %v", err)
	}
	channel.SetUserConnected(true)

	n.users[userID] = channel
	n.channels[chanID] = channel

	if err := channel.SetOpeningState(); err != nil {
		return nil, errors.Errorf("unable set opening channel state: %v", err)
	}

	log.Tracef("User(%v) opened channel(%v) with client", userID, chanID)

	// Subscribe on block notification and update channel when block is
	// generated.
	l := n.blockNotifier.Subscribe()

	// Channel is able to operate only after block is generated.
	// Send update that channel is opened only after it is unlocked.
	go func() {
		defer l.Stop()
		<-l.Read()

		n.Lock()
		defer n.Unlock()

		if err := channel.SetOpenedState(); err != nil {
			log.Errorf("unable set opened channel "+
				"state: %v", err)
			return
		}

		log.Tracef("Channel(%v) with user(%v) opened", chanID, userID)
	}()

	return &OpenChannelResponse{
		ChannelId: strconv.FormatUint(n.channelIndex, 10),
	}, nil
}

// CloseChannel is used to emulate that user has closed the channel with the
// lightning.
//
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) CloseChannel(_ context.Context, req *CloseChannelRequest) (
	*CloseChannelResponse, error) {

	n.Lock()
	defer n.Unlock()

	chanID := lightning.ChannelID(req.ChannelId)
	channel, ok := n.channels[chanID]
	if !ok {
		return nil, errors.Errorf("unable to find the channel with %v id", chanID)
	} else if channel.IsPending() {
		return nil, errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	// Increase the pending balance till block is generated.
	n.client.pendingBalance += channel.LocalBalance

	if err := channel.SetClosingState(); err != nil {
		return nil, errors.Errorf("unable set closing channel state: %v", err)
	}

	log.Tracef("User(%v) closed channel(%v)", channel.UserID, channel.ChannelID)

	// Subscribe on block notification and return funds when block is
	// generated.
	l := n.blockNotifier.Subscribe()

	// Update client free balance only after block is mined and increase
	// client balance on amount which we locked on our side in this channel.
	go func() {
		defer l.Stop()
		<-l.Read()

		n.Lock()
		defer n.Unlock()

		if err := channel.SetClosedState(); err != nil {
			log.Errorf("unable set closing channel state: %v", err)
			return
		}

		delete(n.channels, chanID)
		delete(n.users, channel.UserID)

		n.client.pendingBalance -= channel.LocalBalance
		n.client.freeBalance += channel.LocalBalance

		log.Tracef("Client received %v money previously locked in"+
			" channel(%v)", channel.LocalBalance, channel.ChannelID)
	}()

	return &CloseChannelResponse{}, nil
}

// SetBlockGenDuration is used to set the time which is needed for blokc
// to be generated time. This would impact channel creation, channel
// update and channel close.
//
// NOTE: Part of the EmulatorServer interface.
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

// SetBlockchainFee is used to set the fee which blockchain takes for
// making an computation, transaction creation, i.e. channel updates.
//
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) SetBlockchainFee(_ context.Context,
	req *SetBlockchainFeeRequest) (*SetBlockchainFeeResponse, error) {
	n.Lock()
	defer n.Unlock()

	n.blockchainFee = req.Fee
	return &SetBlockchainFeeResponse{}, nil
}

// SetUserConnected set user being offline or online, which means that all his
// opened channels either could or could't be used for receiving and
// sending payments.
//
// NOTE: Part of the EmulatorServer interface.
func (n *emulationNetwork) SetUserConnected(_ context.Context,
	req *SetUserConnectedRequest) (*SetUserConnectedResponse, error) {
	n.Lock()
	defer n.Unlock()

	// TODO(andrew.shvv) add multi channel support
	channel, ok := n.users[lightning.UserID(req.UserId)]
	if !ok {
		return nil, errors.Errorf("unable to find user %v",
			req.UserId)
	}

	if channel.IsUserConnected != req.IsOnline {
		channel.SetUserConnected(req.IsOnline)

		n.broadcaster.Write(&lightning.UpdateUserConnected{
			User:        lightning.UserID(req.UserId),
			IsConnected: req.IsOnline,
		})
	}

	return &SetUserConnectedResponse{}, nil
}
