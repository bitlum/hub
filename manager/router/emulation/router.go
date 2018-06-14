package emulation

import (
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
	"time"
	"strconv"
	"github.com/bitlum/hub/manager/common/broadcast"
)

// RouterEmulation is an implementation of router. Router interface which
// completely detached from real lightning network daemon and emulates it
// activity.
type RouterEmulation struct {
	freeBalance    router.BalanceUnit
	pendingBalance router.BalanceUnit
	network        *emulationNetwork
	feeBase        int64
	feeProportion  int64
}

// Runtime check that RouterEmulation implements router.Router interface.
var _ router.Router = (*RouterEmulation)(nil)

// NewRouter creates new entity of emulator router and start grpc server which l
func NewRouter(freeBalance router.BalanceUnit,
	blockGeneration time.Duration) *RouterEmulation {

	n := newEmulationNetwork(blockGeneration)
	r := &RouterEmulation{
		freeBalance: freeBalance,
		network:     n,
	}
	n.router = r

	return r
}

// Done returns error if router stopped working for some reason,
// and nil if it was stopped.
//
// NOTE: Part of the router.Router interface.
func (r *RouterEmulation) Done() chan error {
	return r.network.errChan
}

// Stop...
func (r *RouterEmulation) Start(host, port string) {
	r.network.start(host, port)
}

// Stop...
func (r *RouterEmulation) Stop() {
	r.network.stop()
}

// SendPayment makes the payment on behalf of router. In the context of
// lightning network hub manager this hook might be used for future
// off-chain channel re-balancing tactics.
func (r *RouterEmulation) SendPayment(userID router.UserID,
	amount router.BalanceUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	// TODO(andrew.shvv) Implement for rebalancing
	return errors.Errorf("not implemented")
}

// OpenChannel opens the channel with the given user.
func (r *RouterEmulation) OpenChannel(userID router.UserID,
	funds router.BalanceUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	r.network.channelIndex++
	id := strconv.FormatUint(r.network.channelIndex, 10)
	chanID := router.ChannelID(id)

	if _, ok := r.network.users[userID]; ok {
		// TODO(andrew.shvv) add multiple channels support
		return errors.Errorf("multiple channels unsupported")
	}

	// Ensure that initiator has enough funds to open and close the channel.
	if funds-router.BalanceUnit(r.network.blockchainFee) <= 0 {
		return errors.Errorf("router balance is not sufficient to "+
			"open the channel, need(%v)", r.network.blockchainFee)
	} else if funds-router.BalanceUnit(2*r.network.blockchainFee) <= 0 {
		return errors.Errorf("router balance is not sufficient to "+
			"close the channel after opening, need(%v)", 2*r.network.blockchainFee)
	}

	// Take fee for opening and closing the channel, from channel initiator,
	// and save close fee so that we could use it later for paying the
	// blockchain.
	openChannelFee := router.BalanceUnit(r.network.blockchainFee)
	closeChannelFee := router.BalanceUnit(r.network.blockchainFee)
	routerBalance := funds - openChannelFee - closeChannelFee
	fundingAmount := funds

	channel := router.NewChannel(chanID, userID, fundingAmount, 0,
		routerBalance, closeChannelFee, router.UserInitiator)

	r.network.users[userID] = channel
	r.network.channels[chanID] = channel

	channel.SetOpeningState()
	r.network.broadcaster.Write(&router.UpdateChannelOpening{
		UserID:        channel.UserID,
		ChannelID:     channel.ChannelID,
		UserBalance:   channel.UserBalance,
		RouterBalance: channel.RouterBalance,
		Fee:           openChannelFee,
	})

	log.Tracef("Router opened channel(%v) with user(%v)", chanID, userID)

	// Subscribe on block notification and update channel when block is
	// generated.
	l := r.network.blockNotifier.Subscribe()

	// Channel is able to operate only after block is generated.
	// Send update that channel is opened only after it is unlocked.
	start := time.Now()
	go func() {
		defer l.Stop()
		<-l.Read()

		r.network.Lock()
		defer r.network.Unlock()

		channel.SetOpenedState()
		r.network.broadcaster.Write(&router.UpdateChannelOpened{
			UserID:        channel.UserID,
			ChannelID:     channel.ChannelID,
			UserBalance:   channel.UserBalance,
			RouterBalance: channel.RouterBalance,
			Fee:           openChannelFee,
			Duration:      time.Now().Sub(start),
		})

		log.Tracef("Channel(%v) with user(%v) unlocked", chanID, userID)
	}()

	return nil
}

// CloseChannel closes the specified channel.
func (r *RouterEmulation) CloseChannel(id router.ChannelID) error {
	r.network.Lock()
	defer r.network.Unlock()

	if channel, ok := r.network.channels[id]; !ok {
		return errors.Errorf("unable to find channel with %v id: %v", id)
	} else if channel.IsPending() {
		return errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	// TODO(andrew.shvv) add multiple channels support
	for userID, channel := range r.network.users {
		if channel.ChannelID == id {

			r.pendingBalance += channel.RouterBalance

			// Lock the channel and send the closing notification.
			// Wait for block to be generated and only after that remove it
			// from router network.
			channel.SetClosingState()
			r.network.broadcaster.Write(&router.UpdateChannelClosing{
				UserID:    userID,
				ChannelID: id,
				Fee:       channel.CloseFee,
			})

			log.Tracef("Router closed channel(%v)", id)

			// Subscribe on block notification and return funds when block is
			// generated.
			l := r.network.blockNotifier.Subscribe()

			// Update router free balance only after block is mined and increase
			// router balance on amount which we locked on our side in this channel.
			start := time.Now()
			go func() {
				defer l.Stop()
				<-l.Read()

				r.network.Lock()
				defer r.network.Unlock()

				delete(r.network.users, userID)
				delete(r.network.channels, id)

				r.pendingBalance -= channel.RouterBalance
				r.freeBalance += channel.RouterBalance

				r.network.broadcaster.Write(&router.UpdateChannelClosed{
					UserID:    userID,
					ChannelID: id,
					Fee:       channel.CloseFee,
					Duration:  time.Now().Sub(start),
				})

				log.Tracef("Router received %v money previously locked in"+
					" channel(%v)", channel.RouterBalance, id)
			}()

			break
		}
	}

	return nil
}

// UpdateChannel updates the number of locked funds in the specified
// channel.
func (r *RouterEmulation) UpdateChannel(id router.ChannelID,
	newRouterBalance router.BalanceUnit) error {
	r.network.Lock()
	defer r.network.Unlock()

	channel, ok := r.network.channels[id]
	if !ok {
		return errors.Errorf("unable to find the channel with %v id", id)
	} else if channel.IsPending() {
		return errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	if newRouterBalance < 0 {
		return errors.New("new balance is lower than zero")
	}

	diff := newRouterBalance - channel.RouterBalance
	fee := router.BalanceUnit(r.network.blockchainFee)

	if diff > 0 {
		// Number of funds we want to add from our free balance to the
		// channel on router side.
		sliceInFunds := diff

		if sliceInFunds+fee > r.freeBalance {
			return errors.Errorf("insufficient free funds")
		}

		r.freeBalance -= sliceInFunds + fee
		r.pendingBalance += sliceInFunds
	} else {
		// Number of funds we want to get from our channel to the
		// channel on free balance.
		sliceOutFunds := -diff

		// Redundant check, left here just for security if input values would
		if sliceOutFunds+fee > channel.RouterBalance {
			return errors.Errorf("insufficient funds in channel")
		}

		channel.RouterBalance -= sliceOutFunds + fee
		r.pendingBalance += sliceOutFunds
	}

	// During channel update make it locked, so that it couldn't be used by
	// both sides.
	channel.SetClosedState()
	r.network.broadcaster.Write(&router.UpdateChannelUpdating{
		UserID:        channel.UserID,
		ChannelID:     channel.ChannelID,
		UserBalance:   channel.UserBalance,
		RouterBalance: channel.RouterBalance,
		Fee:           fee,
	})

	// Subscribe on block notification and return funds when block is
	// generated.
	l := r.network.blockNotifier.Subscribe()

	// Update router free balance only after block is mined and increase
	// router balance on amount which we locked on our side in this channel.
	start := time.Now()
	go func() {
		defer l.Stop()
		<-l.Read()

		r.network.Lock()
		defer r.network.Unlock()

		if diff > 0 {
			// Number of funds we want to add from our pending balance to the
			// channel on router side.
			sliceInFunds := diff

			r.pendingBalance -= sliceInFunds
			channel.RouterBalance += sliceInFunds
		} else {
			// Number of funds we want to get from our pending channel
			// balance to the free balance.
			sliceOutFunds := -diff

			r.pendingBalance -= sliceOutFunds
			r.freeBalance += sliceOutFunds
		}

		log.Tracef("Update channel(%v) balance, old(%v) => new(%v)",
			channel.RouterBalance, newRouterBalance)

		channel.SetOpenedState()
		r.network.broadcaster.Write(&router.UpdateChannelUpdated{
			UserID:        channel.UserID,
			ChannelID:     channel.ChannelID,
			UserBalance:   channel.UserBalance,
			RouterBalance: channel.RouterBalance,
			Fee:           fee,
			Duration:      time.Now().Sub(start),
		})
	}()

	return nil

}

// RegisterOnUpdates returns updates about router local network topology
// changes, about attempts of propagating the payment through the
// router, about fee changes etc.
func (r *RouterEmulation) RegisterOnUpdates() *broadcast.Receiver {
	r.network.Lock()
	defer r.network.Unlock()

	return r.network.broadcaster.Subscribe()
}

// Network returns the information about the current local network router
// topology.
func (r *RouterEmulation) Network() ([]*router.Channel, error) {
	r.network.Lock()
	defer r.network.Unlock()

	var channels []*router.Channel
	for _, channel := range r.network.channels {
		channels = append(channels, channel)
	}

	return channels, nil
}

// FreeBalance returns the amount of funds at router disposal.
func (r *RouterEmulation) FreeBalance() (router.BalanceUnit, error) {
	r.network.Lock()
	defer r.network.Unlock()

	return r.freeBalance, nil
}

// PendingBalance returns the amount of funds which in the process of
// being accepted by blockchain.
func (r *RouterEmulation) PendingBalance() (router.BalanceUnit, error) {
	r.network.Lock()
	defer r.network.Unlock()

	return r.pendingBalance, nil
}

// AverageChangeUpdateDuration average time which is needed the change of
// state to ba updated over blockchain.
func (r *RouterEmulation) AverageChangeUpdateDuration() (time.Duration, error) {
	r.network.Lock()
	defer r.network.Unlock()

	return r.network.blockGeneration, nil
}

// SetFeeBase sets base number of milli units (i.e milli satoshis in
// Bitcoin) which will be taken for every forwarding payment.
func (r *RouterEmulation) SetFeeBase(feeBase int64) error {
	r.feeBase = feeBase
	return nil
}

// SetFeeProportional sets the number of milli units (i.e milli
// satoshis in Bitcoin) which will be taken for every killo-unit of
// forwarding payment amount as a forwarding fee.
func (r *RouterEmulation) SetFeeProportional(feeProportional int64) error {
	r.feeProportion = feeProportional
	return nil
}

func (r *RouterEmulation) Asset() string {
	return "BTC"
}
