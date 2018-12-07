package emulation

import (
	"github.com/bitlum/hub/common/broadcast"
	"github.com/bitlum/hub/lightning"
	"github.com/go-errors/errors"
	"strconv"
	"time"
)

// Client is an implementation of lightning. Client interface which
// completely detached from real lightning network daemon and emulates it
// activity.
type Client struct {
	freeBalance    lightning.BalanceUnit
	pendingBalance lightning.BalanceUnit
	network        *emulationNetwork
	feeBase        int64
	feeProportion  int64
}

// Runtime check that Client implements lightning.Client interface.
var _ lightning.Client = (*Client)(nil)

// NewClient creates new entity of emulator lightning client and start gRPC
// server to control it.
func NewClient(freeBalance lightning.BalanceUnit,
	blockGeneration time.Duration) *Client {

	n := newEmulationNetwork(blockGeneration)
	client := &Client{
		freeBalance: freeBalance,
		network:     n,
	}
	n.client = client

	return client
}

// Done returns error if client stopped working for some reason,
// and nil if it was stopped.
//
// NOTE: Part of the lightning.Client interface.
func (client *Client) Done() chan error {
	return client.network.errChan
}

// Stop...
func (client *Client) Start(host, port string) {
	client.network.start(host, port)
}

// Stop...
func (client *Client) Stop() {
	client.network.stop()
}

// SendPayment makes the payment on behalf of lightning. In the context of
// lightning network hub manager this hook might be used for future
// off-chain channel re-balancing tactics.
func (client *Client) SendPayment(userID lightning.UserID,
	amount lightning.BalanceUnit) error {
	client.network.Lock()
	defer client.network.Unlock()

	// TODO(andrew.shvv) Implement for rebalancing
	return errors.Errorf("not implemented")
}

// OpenChannel opens the channel with the given user.
func (client *Client) OpenChannel(userID lightning.UserID,
	funds lightning.BalanceUnit) error {
	client.network.Lock()
	defer client.network.Unlock()

	client.network.channelIndex++
	id := strconv.FormatUint(client.network.channelIndex, 10)
	chanID := lightning.ChannelID(id)

	if _, ok := client.network.users[userID]; ok {
		// TODO(andrew.shvv) add multiple channels support
		return errors.Errorf("multiple channels unsupported")
	}

	// Ensure that initiator has enough funds to open and close the channel.
	if funds-lightning.BalanceUnit(client.network.blockchainFee) <= 0 {
		return errors.Errorf("client balance is not sufficient to "+
			"open the channel, need(%v)", client.network.blockchainFee)
	} else if funds-lightning.BalanceUnit(2*client.network.blockchainFee) <= 0 {
		return errors.Errorf("client balance is not sufficient to "+
			"close the channel after opening, need(%v)", 2*client.network.blockchainFee)
	}

	// Take fee for opening and closing the channel, from channel initiator,
	// and save close fee so that we could use it later for paying the
	// blockchain.
	openChannelFee := lightning.BalanceUnit(client.network.blockchainFee)
	closeChannelFee := lightning.BalanceUnit(client.network.blockchainFee)
	localBalance := funds - openChannelFee - closeChannelFee
	fundingAmount := funds

	cfg := &lightning.ChannelConfig{
		Broadcaster: client.network.broadcaster,
		Storage:     &StubChannelStorage{},
	}

	channel, err := lightning.NewChannel(chanID, userID, fundingAmount, 0,
		localBalance, closeChannelFee, lightning.RemoteInitiator, cfg)
	if err != nil {
		return errors.Errorf("unable create channel: %v", err)
	}

	if err := channel.SetUserConnected(true); err != nil {
		return errors.Errorf("unable set user active: %v", err)
	}

	client.network.users[userID] = channel
	client.network.channels[chanID] = channel

	if err := channel.SetOpeningState(); err != nil {
		return errors.Errorf("unable set opening state: %v", err)
	}

	log.Tracef("Client opened channel(%v) with user(%v)", chanID, userID)

	// Subscribe on block notification and update channel when block is
	// generated.
	l := client.network.blockNotifier.Subscribe()

	// Channel is able to operate only after block is generated.
	// Send update that channel is opened only after it is unlocked.
	go func() {
		defer l.Stop()
		<-l.Read()

		client.network.Lock()
		defer client.network.Unlock()

		if err := channel.SetOpenedState(); err != nil {
			log.Errorf("unable set opened state: %v", err)
			return
		}

		log.Tracef("Channel(%v) with user(%v) unlocked", chanID, userID)
	}()

	return nil
}

// CloseChannel closes the specified channel.
func (client *Client) CloseChannel(id lightning.ChannelID) error {
	client.network.Lock()
	defer client.network.Unlock()

	if channel, ok := client.network.channels[id]; !ok {
		return errors.Errorf("unable to find channel with %v id: %v", id)
	} else if channel.IsPending() {
		return errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	// TODO(andrew.shvv) add multiple channels support
	for userID, channel := range client.network.users {
		if channel.ChannelID == id {

			client.pendingBalance += channel.LocalBalance

			// Lock the channel and send the closing notification.
			// Wait for block to be generated and only after that remove it
			// from client network.
			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set closing state: %v", err)
			}

			log.Tracef("Client closed channel(%v)", id)

			// Subscribe on block notification and return funds when block is
			// generated.
			l := client.network.blockNotifier.Subscribe()

			// Update client free balance only after block is mined and increase
			// client balance on amount which we locked on our side in this channel.
			go func() {
				defer l.Stop()
				<-l.Read()

				client.network.Lock()
				defer client.network.Unlock()

				delete(client.network.users, userID)
				delete(client.network.channels, id)

				client.pendingBalance -= channel.LocalBalance
				client.freeBalance += channel.LocalBalance

				if err := channel.SetClosedState(); err != nil {
					log.Errorf("unable to set closed state: %v", err)
					return
				}

				log.Tracef("Client received %v money previously locked in"+
					" channel(%v)", channel.LocalBalance, id)
			}()

			break
		}
	}

	return nil
}

// UpdateChannel updates the number of locked funds in the specified
// channel.
func (client *Client) UpdateChannel(id lightning.ChannelID,
	newClientBalance lightning.BalanceUnit) error {
	client.network.Lock()
	defer client.network.Unlock()

	channel, ok := client.network.channels[id]
	if !ok {
		return errors.Errorf("unable to find the channel with %v id", id)
	} else if channel.IsPending() {
		return errors.Errorf("channel %v is locked",
			channel.ChannelID)
	}

	if newClientBalance < 0 {
		return errors.New("new balance is lower than zero")
	}

	diff := newClientBalance - channel.LocalBalance
	fee := lightning.BalanceUnit(client.network.blockchainFee)

	if diff > 0 {
		// Number of funds we want to add from our free balance to the
		// channel on client side.
		sliceInFunds := diff

		if sliceInFunds+fee > client.freeBalance {
			return errors.Errorf("insufficient free funds")
		}

		client.freeBalance -= sliceInFunds + fee
		client.pendingBalance += sliceInFunds
	} else {
		// Number of funds we want to get from our channel to the
		// channel on free balance.
		sliceOutFunds := -diff

		// Redundant check, left here just for security if input values would
		if sliceOutFunds+fee > channel.LocalBalance {
			return errors.Errorf("insufficient funds in channel")
		}

		channel.LocalBalance -= sliceOutFunds + fee
		client.pendingBalance += sliceOutFunds
	}

	// During channel update make it locked, so that it couldn't be used by
	// both sides.
	if err := channel.SetUpdatingState(fee); err != nil {
		return errors.Errorf("unable to set updating state: %v", err)
	}

	// Subscribe on block notification and return funds when block is
	// generated.
	l := client.network.blockNotifier.Subscribe()

	// Update client free balance only after block is mined and increase
	// client balance on amount which we locked on our side in this channel.
	go func() {
		defer l.Stop()
		<-l.Read()

		client.network.Lock()
		defer client.network.Unlock()

		if diff > 0 {
			// Number of funds we want to add from our pending balance to the
			// channel on client side.
			sliceInFunds := diff

			client.pendingBalance -= sliceInFunds
			channel.LocalBalance += sliceInFunds
		} else {
			// Number of funds we want to get from our pending channel
			// balance to the free balance.
			sliceOutFunds := -diff

			client.pendingBalance -= sliceOutFunds
			client.freeBalance += sliceOutFunds
		}

		log.Tracef("Update channel(%v) balance, old(%v) => new(%v)",
			channel.LocalBalance, newClientBalance)

		if err := channel.SetUpdatedState(fee); err != nil {
			log.Errorf("unable to set updated/opened state: %v", err)
			return
		}
	}()

	return nil

}

// RegisterOnUpdates returns updates about client local network topology
// changes, about attempts of propagating the payment through the
// client, about fee changes etc.
func (client *Client) RegisterOnUpdates() *broadcast.Receiver {
	client.network.Lock()
	defer client.network.Unlock()

	return client.network.broadcaster.Subscribe()
}

// Channels returns all channels which are connected to lightning.
func (client *Client) Channels() ([]*lightning.Channel, error) {
	client.network.Lock()
	defer client.network.Unlock()

	var channels []*lightning.Channel
	for _, channel := range client.network.channels {
		channels = append(channels, channel)
	}

	return channels, nil
}

// Users return all users which connected or were connected to client
// with payment channel.
func (client *Client) Users() ([]*lightning.User, error) {
	client.network.Lock()
	defer client.network.Unlock()

	// TODO(andrew.shvv) Implement
	var users []*lightning.User
	return users, nil
}

// FreeBalance returns the amount of funds at client disposal.
func (client *Client) FreeBalance() (lightning.BalanceUnit, error) {
	client.network.Lock()
	defer client.network.Unlock()

	return client.freeBalance, nil
}

// PendingBalance returns the amount of funds which in the process of
// being accepted by blockchain.
func (client *Client) PendingBalance() (lightning.BalanceUnit, error) {
	client.network.Lock()
	defer client.network.Unlock()

	return client.pendingBalance, nil
}

// SetFeeBase sets base number of milli units (i.e milli satoshis in
// Bitcoin) which will be taken for every forwarding payment.
func (client *Client) SetFeeBase(feeBase int64) error {
	client.feeBase = feeBase
	return nil
}

// SetFeeProportional sets the number of milli units (i.e milli
// satoshis in Bitcoin) which will be taken for every killo-unit of
// forwarding payment amount as a forwarding fee.
func (client *Client) SetFeeProportional(feeProportional int64) error {
	client.feeProportion = feeProportional
	return nil
}

func (client *Client) Asset() string {
	return "BTC"
}
