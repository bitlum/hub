package emulation

import (
	"testing"
	"context"
	"github.com/bitlum/hub/lightning"
	"reflect"
	"time"
	"google.golang.org/grpc"
	"github.com/go-errors/errors"
	"github.com/bitlum/hub/common/broadcast"
	"github.com/bitlum/hub/optimisation"
)

func TestStartNetwork(t *testing.T) {
	n := newEmulationNetwork(10 * time.Millisecond)
	n.start("localhost", "12674")
	n.stop()
	if err := <-n.done(); err != nil {
		t.Fatalf("stoped with error: %v", err)
	}
}

func TestEmulationNetwork(t *testing.T) {
	var obj interface{}

	// Manually start the block notifier without starting the client.
	client := NewClient(100, time.Hour)
	go client.network.blockNotifier.Start()
	defer client.network.blockNotifier.Stop()

	updates := client.RegisterOnUpdates()
	defer updates.Stop()

	// This subscription is used to understand when new block has been
	// generated.
	l := client.network.blockNotifier.Subscribe()

	if _, err := client.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	obj = &lightning.UpdateChannelOpening{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	client.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &lightning.UpdateChannelOpened{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	if _, err := client.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "2",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	obj = &lightning.UpdateChannelOpening{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	client.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &lightning.UpdateChannelOpened{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Check uninitialised sender and receiver
	if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "",
		Receiver: "",
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check unknown sender
	if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender: "5",
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check unknown receiver
	if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
		Receiver: "5",
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check insufficient funds
	if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   5,
	}); err != nil {
		t.Fatalf("unable to make a payment: %v", err)
	}

	obj = &lightning.UpdatePayment{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Check that we couldn't lock more that we have
	if err := client.UpdateChannel("2", 1000); err == nil {
		t.Fatalf("error haven't been received")
	}

	if err := client.UpdateChannel("2", 10); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}

	obj = &lightning.UpdateChannelUpdating{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	pendingBalance, err := client.PendingBalance()
	if err != nil {
		t.Fatalf("unable to get pending balance: %v", err)
	}

	if pendingBalance != 10 {
		t.Fatal("wrong pending balance")
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	client.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
		// Wait for balance to be updated after block is generated.
		time.Sleep(100 * time.Millisecond)
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &lightning.UpdateChannelUpdated{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   5,
	}); err != nil {
		t.Fatalf("unable to make a payment: %v", err)
	}

	obj = &lightning.UpdatePayment{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	if client.network.channels[lightning.ChannelID("1")].LocalBalance != 5 {
		t.Fatalf("wrong client balance")
	}

	if client.network.channels[lightning.ChannelID("1")].RemoteBalance != 5 {
		t.Fatalf("wrong user balance")
	}

	if client.network.channels[lightning.ChannelID("2")].LocalBalance != 5 {
		t.Fatalf("wrong client balance")
	}

	if client.network.channels[lightning.ChannelID("2")].RemoteBalance != 15 {
		t.Fatalf("wrong user balance")
	}

	if client.freeBalance != 90 {
		t.Fatalf("wrong client free balance")
	}

	// Close channel from side of client
	if err := client.CloseChannel("2"); err != nil {
		t.Fatalf("unable to close the channel")
	}

	obj = &lightning.UpdateChannelClosing{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	client.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
		// Wait for balance to be updated after block is generated.
		time.Sleep(100 * time.Millisecond)
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &lightning.UpdateChannelClosed{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	balance, err := client.FreeBalance()
	if err != nil {
		t.Fatalf("unable to get free balance: %v", err)
	}

	if balance != 95 {
		t.Fatalf("client free balance hasn't been updated")
	}

	// Close channel from side of user
	if _, err := client.network.CloseChannel(context.Background(),
		&CloseChannelRequest{ChannelId: "1"}); err != nil {
		t.Fatalf("unable to close the channel")
	}

	obj = &lightning.UpdateChannelClosing{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	client.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &lightning.UpdateChannelClosed{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	balance, err = client.FreeBalance()
	if err != nil {
		t.Fatalf("unable to get free balance: %v", err)
	}

	if balance != 100 {
		t.Fatalf("client free balance hasn't been updated")
	}
}

func TestIncomingOutgoingPayments(t *testing.T) {
	// Manually start the block notifier without starting the client.
	client := NewClient(100, time.Hour)
	go client.network.blockNotifier.Start()
	defer client.network.blockNotifier.Stop()

	updates := client.RegisterOnUpdates()
	defer updates.Stop()

	// From every payment we get 2 satoshi
	client.SetFeeBase(toMilli(1))

	if _, err := client.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	if err := waitChannelUpdate(client, updates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	if err := client.UpdateChannel("1", 10); err != nil {
		t.Fatalf("unable to udpate channel balance: %v", err)
	}

	if err := waitChannelUpdate(client, updates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	{
		if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
			Sender:   "1",
			Receiver: "0",
			Amount:   1,
		}); err != nil {
			t.Fatalf("unable to receive incoming payment: %v", err)
		}

		if client.network.channels["1"].RemoteBalance != 9 {
			t.Fatalf("wrong baalnce")
		}

		if client.network.channels["1"].LocalBalance != 11 {
			t.Fatalf("wrong baalnce")
		}

		update := <-updates.Read()
		payment := update.(*lightning.UpdatePayment)
		if payment.Amount != 1 {
			t.Fatalf("wrong amount")
		}

		// Fee exists only on forwarding payments.
		if payment.Earned != 0 {
			t.Fatalf("wrong client fee/earned")
		}
	}
	{
		if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
			Sender:   "0",
			Receiver: "1",
			Amount:   1,
		}); err != nil {
			t.Fatalf("unable to receive incoming payment: %v", err)
		}

		if client.network.channels["1"].RemoteBalance != 10 {
			t.Fatalf("wrong baalnce")
		}

		if client.network.channels["1"].LocalBalance != 10 {
			t.Fatalf("wrong baalnce")
		}

		update := <-updates.Read()
		payment := update.(*lightning.UpdatePayment)
		if payment.Amount != 1 {
			t.Fatalf("wrong amount")
		}

		if payment.Earned != 0 {
			t.Fatalf("wrong client fee/earned")
		}
	}
}

func TestSimpleStrategy(t *testing.T) {
	strategy := optimisation.NewChannelUpdateStrategy()
	client := NewClient(100, time.Hour)

	updates := client.RegisterOnUpdates()
	defer updates.Stop()

	// Start emulation client serving on port, and connect to it over gRPC
	// client.
	client.Start("localhost", "37968")
	defer client.Stop()

	ops := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial("localhost:37968", ops...)
	if err != nil {
		t.Fatalf("unable to connect to client: %v", err)
	}
	c := NewEmulatorClient(conn)

	// Send request for opening channel from user side.
	if _, err := c.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 1000,
	}); err != nil {
		t.Fatalf("user unable to open channel: %v", err)
	}

	// As far as update requires block generating we have to emulate it and
	// wait for state update notification to be received.
	<-updates.Read()
	client.network.blockNotifier.MineBlock()
	<-updates.Read()

	// Change the network and lock
	currentNetwork, err := client.Channels()
	if err != nil {
		t.Fatalf("unable to get client topology: %v", err)
	}

	// Create empty network. All channels in this case has to be removed.
	var newNetwork []*lightning.Channel

	actions := strategy.GenerateActions(currentNetwork, newNetwork)
	for _, changeState := range actions {
		if err := changeState(client); err != nil {
			t.Fatalf("unable to apply change state function to "+
				"the client: %v", err)

		}
	}

	<-updates.Read()
	client.network.blockNotifier.MineBlock()
	<-updates.Read()

	currentNetwork, err = client.Channels()
	if err != nil {
		t.Fatalf("unable to get client topology: %v", err)
	}

	if len(currentNetwork) != 0 {
		t.Fatalf("network is not empty")
	}

}

func TestUpdateChannelFee(t *testing.T) {
	client := NewClient(100, time.Hour)

	// Manually start te network to avoid automatic block generation.
	go client.network.blockNotifier.Start()
	defer client.network.blockNotifier.Stop()

	nodeUpdates := client.RegisterOnUpdates()
	defer nodeUpdates.Stop()

	// This subscription is used to understand when new block has been
	// generated in the simulation network.
	blocks := client.network.blockNotifier.Subscribe()

	blockchainFee := lightning.BalanceUnit(1)
	client.network.SetBlockchainFee(context.Background(), &SetBlockchainFeeRequest{
		Fee: int64(blockchainFee),
	})

	if _, err := client.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	select {
	case update := <-nodeUpdates.Read():
		u := update.(*lightning.UpdateChannelOpening)
		if u.RemoteBalance != 8 {
			t.Fatalf("wrong user balance, fee should be taken")
		}

		if u.Fee != 0 {
			t.Fatalf("wrong fee, should be zero because user is initiator")
		}
	case <-time.After(time.Second * 2):
		t.Fatalf("haven't received update")
	}

	mineBlock(t, client, blocks)
	skipUpdate(nodeUpdates)

	if err := client.UpdateChannel("1", 8); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}
	skipUpdate(nodeUpdates)

	mineBlock(t, client, blocks)
	skipUpdate(nodeUpdates)

	// Close channel from side of client
	if err := client.CloseChannel("1"); err != nil {
		t.Fatalf("unable to close the channel")
	}

	select {
	case update := <-nodeUpdates.Read():
		u := update.(*lightning.UpdateChannelClosing)
		if u.Fee != blockchainFee {
			t.Fatalf("wrong fee")
		}
	case <-time.After(time.Second * 2):
		t.Fatalf("haven't received update")
	}

	balance, err := client.FreeBalance()
	if err != nil {
		t.Fatalf("unable to get free balance: %v", err)
	}

	// Client should pay blockchain fee when it updated the channel.
	if balance != 92-blockchainFee {
		t.Fatalf("client free balance is wrong: %v", balance)
	}
}

func TestForwardingPaymentFee(t *testing.T) {
	client := NewClient(100, time.Hour)

	// Manually start te network to avoid automatic block generation.
	go client.network.blockNotifier.Start()
	defer client.network.blockNotifier.Stop()

	nodeUpdates := client.RegisterOnUpdates()
	defer nodeUpdates.Stop()

	// Open first channel and update it update balance from client side.
	if _, err := client.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	if err := waitChannelUpdate(client, nodeUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	if err := client.UpdateChannel("1", 10); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}

	if err := waitChannelUpdate(client, nodeUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	// Open second channel and update it update balance from client side.
	if _, err := client.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "2",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	if err := waitChannelUpdate(client, nodeUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	if err := client.UpdateChannel("2", 10); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}

	if err := waitChannelUpdate(client, nodeUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	// For every 10 satoshi we get 2 satoshi as proportional fee
	client.SetFeeProportional(toMilli(200))

	// From every payment we get 2 satoshi
	client.SetFeeBase(toMilli(2))

	if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   2,
	}); err == nil {
		t.Fatalf("should have failed with small amount error")
	}

	obj := &lightning.UpdatePayment{}
	if err := waitUpdate(nodeUpdates, obj); err != nil {
		t.Fatal(err)
	}

	if _, err := client.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   10,
	}); err != nil {
		t.Fatalf("user is unable to forward payment")
	}

	update := <-nodeUpdates.Read()
	payment := update.(*lightning.UpdatePayment)

	if payment.Earned != 4 {
		t.Fatalf("wrong earned / client fee")
	}

	if client.network.channels[lightning.ChannelID("1")].RemoteBalance != 0 {
		t.Fatalf("wrong user balance")
	}

	if client.network.channels[lightning.ChannelID("1")].LocalBalance != 20 {
		t.Fatalf("wrong client balance")
	}

	// Check that client earned fee
	if client.network.channels[lightning.ChannelID("2")].LocalBalance != 4 {
		t.Fatalf("wrong client balance")
	}

	if client.network.channels[lightning.ChannelID("2")].RemoteBalance != 16 {
		t.Fatalf("wrong user balance")
	}
}

// waitChannelUpdate waits for open, update or close of channel.
func waitChannelUpdate(client *Client, updates *broadcast.Receiver) error {
	select {
	case update := <-updates.Read():
		c1 := reflect.TypeOf(&lightning.UpdateChannelUpdating{}) != reflect.TypeOf(update)
		c2 := reflect.TypeOf(&lightning.UpdateChannelOpening{}) != reflect.TypeOf(update)
		c3 := reflect.TypeOf(&lightning.UpdateChannelClosing{}) != reflect.TypeOf(update)

		if c1 && c2 && c3 {
			return errors.Errorf("wrong update type, "+
				"expected one of the channel updates, "+
				"received: %v", reflect.TypeOf(update))
		}
	case <-time.After(time.Second * 2):
		return errors.New("haven't received update")
	}

	// This subscription is used to understand when new block has been
	// generated.
	l := client.network.blockNotifier.Subscribe()

	// Manually trigger block generation and wait for block notification to be
	// received, with this channel should be updated.
	client.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		return errors.New("haven't received block notification")
	}

	select {
	case update := <-updates.Read():
		c1 := reflect.TypeOf(&lightning.UpdateChannelUpdated{}) != reflect.TypeOf(update)
		c2 := reflect.TypeOf(&lightning.UpdateChannelOpened{}) != reflect.TypeOf(update)
		c3 := reflect.TypeOf(&lightning.UpdateChannelClosed{}) != reflect.TypeOf(update)

		if c1 && c2 && c3 {
			return errors.Errorf("wrong update type, "+
				"expected one of the channel updates, "+
				"received: %v", reflect.TypeOf(update))
		}
	case <-time.After(time.Second * 2):
		return errors.New("haven't received update")
	}

	return nil
}

func waitUpdate(receiver *broadcast.Receiver, obj interface{}) error {
	desiredType := reflect.TypeOf(obj)

	select {
	case update := <-receiver.Read():
		if desiredType != reflect.TypeOf(update) {
			return errors.Errorf("wrong update type, expected: %v, "+
				"received: %v", desiredType, reflect.TypeOf(update))
		}
	case <-time.After(time.Second * 2):
		return errors.New("haven't received update")
	}

	return nil
}

func skipUpdate(receiver *broadcast.Receiver) error {
	select {
	case <-receiver.Read():
		return nil
	case <-time.After(time.Second * 2):
		return errors.New("haven't received update")
	}

	return nil
}

func mineBlock(t *testing.T, client *Client, blocks *broadcast.Receiver) {
	// Manually trigger block generation and wait for block notification to be
	// received.
	client.network.blockNotifier.MineBlock()
	select {
	case <-blocks.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}
}
