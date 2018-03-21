package emulation

import (
	"testing"
	"context"
	"github.com/bitlum/hub/manager/router"
	"reflect"
	"errors"
	"time"
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

	// This subscription is used to understand when new block has been
	// generated.
	s, _ := r.network.blockNotifier.Subscribe()

	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       1,
		LockedByUser: 10,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-s.C:
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &router.UpdateChannelOpened{}
	if err := checkUpdate(t, r.network.updates, obj); err != nil {
		t.Fatal(err)
	}

	if _, err := r.network.OpenChannel(context.Background(), &OpenChannelRequest{
		UserId:       2,
		LockedByUser: 0,
	}); err != nil {
		t.Fatalf("unable to emulate user openning channel: %v", err)
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-s.C:
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	obj = &router.UpdateChannelOpened{}
	if err := checkUpdate(t, r.network.updates, obj); err != nil {
		t.Fatal(err)
	}

	// Check uninitialised sender and receiver
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   0,
		Receiver: 0,
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check unknown sender
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender: 5,
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check unknown receiver
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Receiver: 5,
	}); err == nil {
		t.Fatalf("error haven't been recevied")
	}

	// Check insufficient funds
	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   1,
		Receiver: 2,
		Amount:   5,
	}); err != nil {
		t.Fatalf("unable to make a payment")
	}

	// Router error should be sent as a update info rather than error.
	obj = &router.UpdatePayment{}
	if err := checkUpdate(t, r.network.updates, obj); err != nil {
		t.Fatal(err)
	}

	// Check that we couldn't lock more that we have
	if err := r.UpdateChannel(2, 1000); err == nil {
		t.Fatalf("error haven't been received")
	}

	if err := r.UpdateChannel(2, 10); err != nil {
		t.Fatalf("unable to update channel: %v", err)
	}

	if _, err := r.network.SendPayment(context.Background(), &SendPaymentRequest{
		Sender:   1,
		Receiver: 2,
		Amount:   5,
	}); err != nil {
		t.Fatalf("unable to make a payment: %v", err)
	}

	if r.network.channels[router.ChannelID(1)].RouterBalance != 5 {
		t.Fatalf("wrong router balance")
	}

	if r.network.channels[router.ChannelID(1)].UserBalance != 5 {
		t.Fatalf("wrong user balance")
	}

	if r.network.channels[router.ChannelID(2)].RouterBalance != 5 {
		t.Fatalf("wrong router balance")
	}

	if r.network.channels[router.ChannelID(2)].UserBalance != 5 {
		t.Fatalf("wrong user balance")
	}

	if r.freeBalance != 90 {
		t.Fatalf("wrong router free balance")
	}

	// Close channel from side of router
	if err := r.CloseChannel(2); err != nil {
		t.Fatalf("unable to close the channel")
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
		&CloseChannelRequest{ChanId: 1}); err != nil {
		t.Fatalf("unable to close the channel")
	}

	// Manually trigger block generation and wait for block notification to be
	// received.
	r.network.blockNotifier.MineBlock()
	select {
	case <-s.C:
	case <-time.After(time.Second):
		t.Fatalf("haven't received block notification")
	}

	// Wait for balance to be updated after block is generated.
	time.Sleep(100 * time.Millisecond)

	balance, err = r.FreeBalance()
	if err != nil {
		t.Fatalf("unable to get free balance: %v", err)
	}

	if balance != 100 {
		t.Fatalf("router free balance hasn't been updated")
	}
}

func checkUpdate(t *testing.T, updatesChan chan interface{}, obj interface{}) error {
	desiredType := reflect.TypeOf(obj)

	select {
	case update := <-updatesChan:
		if desiredType != reflect.TypeOf(update) {
			return errors.New("wrong update type")
		}
	case <-time.After(time.Second):
		return errors.New("haven't received update")
	}

	return nil
}
