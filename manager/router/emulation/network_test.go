package emulation

import (
	"testing"
	"context"
	"github.com/bitlum/hub/manager/router"
	"reflect"
	"time"
	"google.golang.org/grpc"
	"github.com/go-errors/errors"
	"github.com/bitlum/hub/manager/common/broadcast"
	"github.com/bitlum/hub/manager/optimisation"
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

	// Manually start the block notifier without starting the router.
	r := NewRouter(100, time.Hour)
	go r.network.blockNotifier.Start()
	defer r.network.blockNotifier.Stop()

	updates := r.RegisterOnUpdates()
	defer updates.Stop()

	// This subscription is used to understand when new block has been
	// generated.
	l := r.network.blockNotifier.Subscribe()

	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	obj = &router.UpdateChannelOpening{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &router.UpdateChannelOpened{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "2",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	obj = &router.UpdateChannelOpening{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &router.UpdateChannelOpened{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Check uninitialised sender and receiver
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "",
		Receiver: "",
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check unknown sender
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender: "5",
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check unknown receiver
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Receiver: "5",
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check insufficient funds
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   5,
	}); err != nil {
		t.Fatalf("unable to make a payment: %v", err)
	}

	obj = &router.UpdatePayment{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Check that we couldn't lock more that we have
	if err := r.UpdateChannel("2", 1000); err == nil {
		t.Fatalf("error haven't been received")
	}

	if err := r.UpdateChannel("2", 10); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}

	obj = &router.UpdateChannelUpdating{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	pendingBalance, err := r.PendingBalance()
	if err != nil {
		t.Fatalf("unable to get pending balance: %v", err)
	}

	if pendingBalance != 10 {
		t.Fatal("wrong pending balance")
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
		// Wait for balance to be updated after block is generated.
		time.Sleep(100 * time.Millisecond)
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &router.UpdateChannelUpdated{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   5,
	}); err != nil {
		t.Fatalf("unable to make a payment: %v", err)
	}

	obj = &router.UpdatePayment{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	if r.network.channels[router.ChannelID("1")].RouterBalance != 5 {
		t.Fatalf("wrong router balance")
	}

	if r.network.channels[router.ChannelID("1")].UserBalance != 5 {
		t.Fatalf("wrong user balance")
	}

	if r.network.channels[router.ChannelID("2")].RouterBalance != 5 {
		t.Fatalf("wrong router balance")
	}

	if r.network.channels[router.ChannelID("2")].UserBalance != 15 {
		t.Fatalf("wrong user balance")
	}

	if r.freeBalance != 90 {
		t.Fatalf("wrong router free balance")
	}

	// Close channel from side of router
	if err := r.CloseChannel("2"); err != nil {
		t.Fatalf("unable to close the channel")
	}

	obj = &router.UpdateChannelClosing{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
		// Wait for balance to be updated after block is generated.
		time.Sleep(100 * time.Millisecond)
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &router.UpdateChannelClosed{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	balance, err := r.FreeBalance()
	if err != nil {
		t.Fatalf("unable to get free balance: %v", err)
	}

	if balance != 95 {
		t.Fatalf("router free balance hasn't been updated")
	}

	// Close channel from side of user
	if _, err := r.network.CloseChannel(context.Background(),
		&CloseChannelRequest{ChannelId: "1"}); err != nil {
		t.Fatalf("unable to close the channel")
	}

	obj = &router.UpdateChannelClosing{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &router.UpdateChannelClosed{}
	if err := waitUpdate(updates, obj); err != nil {
		t.Fatal(err)
	}

	balance, err = r.FreeBalance()
	if err != nil {
		t.Fatalf("unable to get free balance: %v", err)
	}

	if balance != 100 {
		t.Fatalf("router free balance hasn't been updated")
	}
}

func TestIncomingOutgoingPayments(t *testing.T) {
	// Manually start the block notifier without starting the router.
	r := NewRouter(100, time.Hour)
	go r.network.blockNotifier.Start()
	defer r.network.blockNotifier.Stop()

	updates := r.RegisterOnUpdates()
	defer updates.Stop()

	// From every payment we get 2 satoshi
	r.SetFeeBase(toMilli(1))

	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	if err := waitChannelUpdate(r, updates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	if err := r.UpdateChannel("1", 10); err != nil {
		t.Fatalf("unable to udpate channel balance: %v", err)
	}

	if err := waitChannelUpdate(r, updates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	{
		if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
			Sender:   "1",
			Receiver: "0",
			Amount:   1,
		}); err != nil {
			t.Fatalf("unable to receive incoming payment: %v", err)
		}

		if r.network.channels["1"].UserBalance != 9 {
			t.Fatalf("wrong baalnce")
		}

		if r.network.channels["1"].RouterBalance != 11 {
			t.Fatalf("wrong baalnce")
		}

		update := <-updates.Read()
		payment := update.(*router.UpdatePayment)
		if payment.Amount != 1 {
			t.Fatalf("wrong amount")
		}

		// Fee exists only on forwarding payments.
		if payment.Earned != 0 {
			t.Fatalf("wrong router fee/earned")
		}
	}
	{
		if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
			Sender:   "0",
			Receiver: "1",
			Amount:   1,
		}); err != nil {
			t.Fatalf("unable to receive incoming payment: %v", err)
		}

		if r.network.channels["1"].UserBalance != 10 {
			t.Fatalf("wrong baalnce")
		}

		if r.network.channels["1"].RouterBalance != 10 {
			t.Fatalf("wrong baalnce")
		}

		update := <-updates.Read()
		payment := update.(*router.UpdatePayment)
		if payment.Amount != 1 {
			t.Fatalf("wrong amount")
		}

		if payment.Earned != 0 {
			t.Fatalf("wrong router fee/earned")
		}
	}
}

func TestSimpleStrategy(t *testing.T) {
	strategy := optimisation.NewChannelUpdateStrategy()
	r := NewRouter(100, time.Hour)

	updates := r.RegisterOnUpdates()
	defer updates.Stop()

	// Start emulation router serving on port, and connect to it over gRPC
	// client.
	r.Start("localhost", "37968")
	defer r.Stop()

	ops := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial("localhost:37968", ops...)
	if err != nil {
		t.Fatalf("unable to connect to router: %v", err)
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
	r.network.blockNotifier.MineBlock()
	<-updates.Read()

	// Change the network and lock
	currentNetwork, err := r.Channels()
	if err != nil {
		t.Fatalf("unable to get router topology: %v", err)
	}

	// Create empty network. All channels in this case has to be removed.
	var newNetwork []*router.Channel

	actions := strategy.GenerateActions(currentNetwork, newNetwork)
	for _, changeState := range actions {
		if err := changeState(r); err != nil {
			t.Fatalf("unable to apply change state function to "+
				"the router: %v", err)

		}
	}

	<-updates.Read()
	r.network.blockNotifier.MineBlock()
	<-updates.Read()

	currentNetwork, err = r.Channels()
	if err != nil {
		t.Fatalf("unable to get router topology: %v", err)
	}

	if len(currentNetwork) != 0 {
		t.Fatalf("network is not empty")
	}

}

func TestUpdateChannelFee(t *testing.T) {
	r := NewRouter(100, time.Hour)

	// Manually start te network to avoid automatic block generation.
	go r.network.blockNotifier.Start()
	defer r.network.blockNotifier.Stop()

	routerUpdates := r.RegisterOnUpdates()
	defer routerUpdates.Stop()

	// This subscription is used to understand when new block has been
	// generated in the simulation network.
	blocks := r.network.blockNotifier.Subscribe()

	blockchainFee := router.BalanceUnit(1)
	r.network.SetBlockchainFee(context.Background(), &SetBlockchainFeeRequest{
		Fee: int64(blockchainFee),
	})

	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	select {
	case update := <-routerUpdates.Read():
		u := update.(*router.UpdateChannelOpening)
		if u.UserBalance != 8 {
			t.Fatalf("wrong user balance, fee should be taken")
		}

		if u.Fee != 0 {
			t.Fatalf("wrong fee, should be zero because user is initiator")
		}
	case <-time.After(time.Second * 2):
		t.Fatalf("haven't received update")
	}

	mineBlock(t, r, blocks)
	skipUpdate(routerUpdates)

	if err := r.UpdateChannel("1", 8); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}
	skipUpdate(routerUpdates)

	mineBlock(t, r, blocks)
	skipUpdate(routerUpdates)

	// Close channel from side of router
	if err := r.CloseChannel("1"); err != nil {
		t.Fatalf("unable to close the channel")
	}

	select {
	case update := <-routerUpdates.Read():
		u := update.(*router.UpdateChannelClosing)
		if u.Fee != blockchainFee {
			t.Fatalf("wrong fee")
		}
	case <-time.After(time.Second * 2):
		t.Fatalf("haven't received update")
	}

	balance, err := r.FreeBalance()
	if err != nil {
		t.Fatalf("unable to get free balance: %v", err)
	}

	// Router should pay blockchain fee when it updated the channel.
	if balance != 92-blockchainFee {
		t.Fatalf("router free balance is wrong: %v", balance)
	}
}

func TestForwardingPaymentFee(t *testing.T) {
	r := NewRouter(100, time.Hour)

	// Manually start te network to avoid automatic block generation.
	go r.network.blockNotifier.Start()
	defer r.network.blockNotifier.Stop()

	routerUpdates := r.RegisterOnUpdates()
	defer routerUpdates.Stop()

	// Open first channel and update it update balance from router side.
	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "1",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	if err := waitChannelUpdate(r, routerUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	if err := r.UpdateChannel("1", 10); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}

	if err := waitChannelUpdate(r, routerUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	// Open second channel and update it update balance from router side.
	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       "2",
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	if err := waitChannelUpdate(r, routerUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	if err := r.UpdateChannel("2", 10); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}

	if err := waitChannelUpdate(r, routerUpdates); err != nil {
		t.Fatalf("haven't received channel update: %v", err)
	}

	// For every 10 satoshi we get 2 satoshi as proportional fee
	r.SetFeeProportional(toMilli(200))

	// From every payment we get 2 satoshi
	r.SetFeeBase(toMilli(2))

	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   2,
	}); err == nil {
		t.Fatalf("should have failed with small amount error")
	}

	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   "1",
		Receiver: "2",
		Amount:   10,
	}); err != nil {
		t.Fatalf("user is unable to forward payment")
	}

	update := <-routerUpdates.Read()
	payment := update.(*router.UpdatePayment)
	if payment.Earned != 4 {
		t.Fatalf("wrong earned / router fee")
	}

	if r.network.channels[router.ChannelID("1")].UserBalance != 0 {
		t.Fatalf("wrong user balance")
	}

	if r.network.channels[router.ChannelID("1")].RouterBalance != 20 {
		t.Fatalf("wrong router balance")
	}

	// Check that router earned fee
	if r.network.channels[router.ChannelID("2")].RouterBalance != 4 {
		t.Fatalf("wrong router balance")
	}

	if r.network.channels[router.ChannelID("2")].UserBalance != 16 {
		t.Fatalf("wrong user balance")
	}
}

// waitChannelUpdate waits for open, update or close of channel.
func waitChannelUpdate(r *RouterEmulation, updates *broadcast.Receiver) error {
	select {
	case update := <-updates.Read():
		c1 := reflect.TypeOf(&router.UpdateChannelUpdating{}) != reflect.TypeOf(update)
		c2 := reflect.TypeOf(&router.UpdateChannelOpening{}) != reflect.TypeOf(update)
		c3 := reflect.TypeOf(&router.UpdateChannelClosing{}) != reflect.TypeOf(update)

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
	l := r.network.blockNotifier.Subscribe()

	// Manually trigger block generation and wait for block notification to be
	// received, with this channel should be updated.
	r.network.blockNotifier.MineBlock()
	select {
	case <-l.Read():
	case <-time.After(time.Second):
		return errors.New("haven't received block notification")
	}

	select {
	case update := <-updates.Read():
		c1 := reflect.TypeOf(&router.UpdateChannelUpdated{}) != reflect.TypeOf(update)
		c2 := reflect.TypeOf(&router.UpdateChannelOpened{}) != reflect.TypeOf(update)
		c3 := reflect.TypeOf(&router.UpdateChannelClosed{}) != reflect.TypeOf(update)

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

func mineBlock(t *testing.T, r *RouterEmulation, blocks *broadcast.Receiver) {
	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-blocks.Read():
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}
}
