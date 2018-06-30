package lnd

import (
	"time"
	"github.com/bitlum/lnd/lnrpc"
	"context"
	"github.com/bitlum/hub/manager/router"
	"github.com/bitlum/hub/manager/metrics/crypto"
	"github.com/bitlum/hub/manager/metrics"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/hub/manager/common/broadcast"
	"github.com/bitlum/hub/manager/router/registry"
)

// updateNodeInfo updates information about last synced version of the
// lightning node, if version has changed or best block has updated,
// we need to know that.
//
// NOTE: Should run as goroutine.
func (r *Router) updateNodeInfo() {
	defer func() {
		log.Info("Stopped lightning node info updates goroutine")
		r.wg.Done()
	}()

	log.Info("Started lightning node info updates goroutine")

	for {
		select {
		case <-time.After(time.Second * 25):
		case <-r.quit:
			return
		}

		m := crypto.NewMetric(r.cfg.Asset, "UpdateNodeInfo", r.cfg.MetricsBackend)
		nodeInfo, err := fetchNodeInfo(r.client)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("unable to fetch node info: %v", err)
			continue
		}
		m.Finish()

		if err := r.cfg.Storage.UpdateInfo(&router.Info{
			Version:     nodeInfo.Version,
			Network:     r.cfg.Net,
			BlockHeight: nodeInfo.BlockHeight,
			BlockHash:   nodeInfo.BlockHash,
			NodeInfo: &router.NodeInfo{
				Alias:          nodeInfo.Alias,
				Host:           r.cfg.PeerHost,
				Port:           r.cfg.PeerPort,
				IdentityPubKey: nodeInfo.IdentityPubkey,
			},
			NeutrinoInfo: &router.NeutrinoInfo{
				Host: r.cfg.NeutrinoHost,
				Port: r.cfg.NeutrinoPort,
			},
		}); err != nil {
			log.Errorf("unable to save lightning node info: %v", err)
		}
	}
}

// updateChannelStates tracks the local channel topology state updates
// and sends notifications accordingly when transition has happened.
//
// NOTE: Should run as goroutine.
func (r *Router) updateChannelStates() {
	defer func() {
		log.Info("Stopped local topology updates goroutine")
		r.wg.Done()
	}()

	log.Info("Started local topology updates goroutine")

	for {
		select {
		case <-time.After(time.Second * 5):
		case <-r.quit:
			return
		}

		m := crypto.NewMetric(r.cfg.Asset, "UpdateChannelStates", r.cfg.MetricsBackend)

		// TODO(andrew.shvv) track waiting closing channels
		openChannels, pendingOpenChannels, pendingClosingChannels,
		pendingForceClosingChannels, _, err := fetchChannels(r.client)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(topology updates) unable to fetch lightning"+
				" channels: %v", err)
			continue
		}

		// Fetch prev pending/closing/opened channels from db which
		// corresponds to the old/previous channel state.
		log.Debugf("(topology updates) Fetching channel state from db...")

		routerChannels, err := r.Channels()
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(topology updates) unable to fetch old channel state: %v"+
				"", err)
			continue
		}

		if err := syncChannelStates(openChannels, pendingOpenChannels,
			pendingClosingChannels, pendingForceClosingChannels, r.broadcaster,
			r.cfg.Storage, routerChannels); err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(topology updates) unable sync channel states: %v"+
				"", err)
			continue
		}
	}
}

// syncChannelStates is used as wrapper for fetching and syncing topology
// updates. Main purpose of creating distinct method was usage of defer, which
// is needed mainly for metric gathering usage.
func syncChannelStates(
	openChannels []*lnrpc.Channel,
	pendingOpenChannels []*lnrpc.PendingChannelsResponse_PendingOpenChannel,
	pendingClosingChannels []*lnrpc.PendingChannelsResponse_ClosedChannel,
	pendingForceClosingChannels []*lnrpc.PendingChannelsResponse_ForceClosedChannel,
	broadcaster *broadcast.Broadcaster,
	channelStorage router.ChannelStorage,
	routerState []*router.Channel) error {

	routerChannelMap := make(map[router.ChannelID]*router.Channel)
	for _, c := range routerState {
		routerChannelMap[c.ChannelID] = c
	}

	// Keep new channel ids to detect channel closes at later point.
	newChannelIDs := make(map[router.ChannelID]struct{}, 0)

	for _, newChannel := range pendingOpenChannels {
		chanID := router.ChannelID(newChannel.Channel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		state := router.ChannelStateName("not exist")
		channel, ok := routerChannelMap[router.ChannelID(newChannel.Channel.ChannelPoint)]
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case router.ChannelOpening:
			// Nothing has changed
		case router.ChannelOpened:
			// Previous channel state was opened, and now it is again
			// opening, something wrong has happened.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, router.ChannelOpened,
				router.ChannelOpening)

		case router.ChannelClosing:
			// Previous channel state was closing, and now it is
			// opening, we couldn't  re-open, closing channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, router.ChannelClosing,
				router.ChannelOpening)

		case router.ChannelClosed:
			// Previous channel state was closing, and now it is
			// opening, we couldn't  re-open, closing channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, router.ChannelClosed,
				router.ChannelOpening)

		case "not exist":
			cfg := &router.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := router.NewChannel(
				router.ChannelID(newChannel.Channel.ChannelPoint),
				router.UserID(newChannel.Channel.RemoteNodePub),
				router.BalanceUnit(newChannel.Channel.Capacity),
				router.BalanceUnit(newChannel.Channel.RemoteBalance),
				router.BalanceUnit(newChannel.Channel.LocalBalance),
				router.BalanceUnit(newChannel.CommitFee),
				getInitiator(newChannel.Channel.LocalBalance),
				cfg,
			)

			if err != nil {
				return errors.Errorf("unable to create new channel("+
					"%v): %v", chanID, err)
			}

			if err := channel.Save(); err != nil {
				return errors.Errorf("unable to save new channel(%v): ¬%v",
					chanID, err)
			}

			log.Infof("Saved new channel(%v)", channel.ChannelID)

			if err := channel.SetOpeningState(); err != nil {
				return errors.Errorf("unable to set opening"+
					" state for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)
		default:
			return errors.Errorf("unhandled state: %v", state)
		}
	}

	for _, newChannel := range pendingClosingChannels {
		chanID := router.ChannelID(newChannel.Channel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		channel, ok := routerChannelMap[router.ChannelID(newChannel.Channel.ChannelPoint)]
		state := router.ChannelStateName("not exist")
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case router.ChannelOpening:
			// Previously channel was opening, it seems that because of the
			// delayed scrape we missed some of state changes.

			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable to set opened state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set closing state"+
					" for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case router.ChannelOpened:
			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set closing state"+
					"for channel(%v): %v", chanID, err)
			}
		case router.ChannelClosing:
			// Nothing has changed
		case router.ChannelClosed:
			// Previous channel state was closed, and now it is
			// closing, we couldn't  re-close, closed channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, router.ChannelClosed,
				router.ChannelClosing)

		case "not exist":
			// Previously channel not existed, it seems that because of the
			// delayed scrape we missed some of state changes.
			cfg := &router.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := router.NewChannel(
				router.ChannelID(newChannel.Channel.ChannelPoint),
				router.UserID(newChannel.Channel.RemoteNodePub),
				router.BalanceUnit(newChannel.Channel.Capacity),
				router.BalanceUnit(newChannel.Channel.RemoteBalance),
				router.BalanceUnit(newChannel.Channel.LocalBalance),
				0,
				getInitiator(newChannel.Channel.LocalBalance),
				cfg,
			)
			if err != nil {
				return errors.Errorf("unable to create new channel("+
					"%v): %v", chanID, err)
			}

			if err := channel.Save(); err != nil {
				return errors.Errorf("unable to save new channel(%v)"+
					": %v", chanID, err)
			}

			log.Infof("Saved new channel(%v)", channel.ChannelID)

			if err := channel.SetOpeningState(); err != nil {
				return errors.Errorf("unable to set opening state "+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable to set opened state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set channel closing"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)
		default:
			return errors.Errorf("unhandled state: %v", state)
		}
	}

	for _, newChannel := range pendingForceClosingChannels {
		chanID := router.ChannelID(newChannel.Channel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		state := router.ChannelStateName("not exist")
		channel, ok := routerChannelMap[router.ChannelID(newChannel.Channel.ChannelPoint)]
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case router.ChannelOpening:
			// Previously channel was opening, it seems that because of the
			// delayed scrape we missed some of state changes.

			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable to set opened state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set channel closing"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set closing state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case router.ChannelOpened:
			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set closing state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case router.ChannelClosing:
			// Nothing has changed
		case router.ChannelClosed:
			// Previous channel state was closed, and now it is
			// closing, we couldn't  re-close, closed channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, router.ChannelClosed,
				router.ChannelClosing)

		case "not exist":
			// Previously channel not existed, it seems that because of the
			// delayed scrape we missed some of state changes.
			cfg := &router.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := router.NewChannel(
				router.ChannelID(newChannel.Channel.ChannelPoint),
				router.UserID(newChannel.Channel.RemoteNodePub),
				router.BalanceUnit(newChannel.Channel.Capacity),
				router.BalanceUnit(newChannel.Channel.RemoteBalance),
				router.BalanceUnit(newChannel.Channel.LocalBalance),
				0,
				getInitiator(newChannel.Channel.LocalBalance, ),
				cfg,
			)
			if err != nil {
				return errors.Errorf("unable to create new channel"+
					"for channel(%v): %v", chanID, err)
			}

			if err := channel.Save(); err != nil {
				return errors.Errorf("unable to save new channel"+
					"for channel(%v): %v", chanID, err)
			}
			log.Infof("Saved new channel(%v)", channel.ChannelID)

			if err := channel.SetOpeningState(); err != nil {
				return errors.Errorf("unable to set opening state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable to set opened state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set channel closing"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)
		default:
			return errors.Errorf("unhandled state: %v", state)
		}
	}

	for _, newChannel := range openChannels {
		chanID := router.ChannelID(newChannel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		channel, ok := routerChannelMap[router.ChannelID(newChannel.ChannelPoint)]
		state := router.ChannelStateName("not exist")
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case router.ChannelOpening:
			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable set channel state"+
					" opened for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case router.ChannelOpened:
			// Nothing has changed
		case router.ChannelClosing:
			// Previous channel state was closing, but now it is
			// opened, close couldn't be canceled.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, router.ChannelClosing,
				router.ChannelOpened)

		case router.ChannelClosed:
			// Previous channel state was closed, but now it is
			// opened, close couldn't be canceled.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, router.ChannelClosed,
				router.ChannelOpened)

		case "not exist":
			cfg := &router.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := router.NewChannel(
				router.ChannelID(newChannel.ChannelPoint),
				router.UserID(newChannel.RemotePubkey),
				router.BalanceUnit(newChannel.Capacity),
				router.BalanceUnit(newChannel.RemoteBalance),
				router.BalanceUnit(newChannel.LocalBalance),
				router.BalanceUnit(newChannel.CommitFee),
				getInitiator(newChannel.LocalBalance),
				cfg,
			)
			if err != nil {
				return errors.Errorf("unable to create new channel("+
					"%v): %v", chanID, err)
			}

			if err := channel.Save(); err != nil {
				return errors.Errorf("unable to save new channel(%v)"+
					": %v", chanID, err)
			}
			log.Infof("Saved new channel(%v)", channel.ChannelID)

			if err := channel.SetOpeningState(); err != nil {
				return errors.Errorf("unable set opening channel"+
					" state for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable set opened channel"+
					" state for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)
		default:
			return errors.Errorf("unhandled state: %v", state)
		}
	}

	for chanID, channel := range routerChannelMap {
		_, ok := newChannelIDs[chanID]
		if ok {
			// It was already handled previously
			continue
		}

		state := channel.CurrentState().Name
		switch state {
		case router.ChannelOpening:
			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable to set opened state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set channel closing"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosedState(); err != nil {
				return errors.Errorf("unable set closed channel"+
					" state for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case router.ChannelOpened:
			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("(topology updates) unable to set channel closing"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

			if err := channel.SetClosedState(); err != nil {
				return errors.Errorf("(topology updates) unable set closed channel"+
					" state for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case router.ChannelClosing:
			if err := channel.SetClosedState(); err != nil {
				return errors.Errorf("unable set closed channel"+
					" state for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case router.ChannelClosed:
			// Nothing to do
		default:
			return errors.Errorf("unhandled state: %v", state)
		}
	}

	return nil
}

// listenForwardingPayments listens for forwarding lightning payments.
// Fetches the lnd forwarding log, and send new update to the broadcaster if
// new payment has been found.
//
// NOTE: Should run as goroutine.
func (r *Router) listenForwardingPayments() {
	defer func() {
		log.Info("Stopped sync forwarding payments goroutine")
		r.wg.Done()
	}()

	log.Info("Started sync forwarding payments goroutine")

	for {
		select {
		case <-time.After(time.Second * 5):
		case <-r.quit:
			return
		}

		r.syncForwardingUpdate()
	}
}

// syncForwardingUpdate is used as wrapper for fetching and syncing forwarding
// updates. Main purpose of creating distinct method was usage of defer, which
// is needed mainly for metric gathering usage.
func (r *Router) syncForwardingUpdate() {
	defer panicRecovering()

	var lastIndex uint32
	var err error

	m := crypto.NewMetric(r.cfg.Asset, "ListenForwardingPayments",
		r.cfg.MetricsBackend)
	defer m.Finish()

	// Try to fetch last index, and if fails than try after yet again after
	// some time.
	for {
		lastIndex, err = r.cfg.Storage.LastForwardingIndex()
		if err != nil {
			m.AddError(metrics.HighSeverity)
			log.Errorf("(forwarding updates) unable to get last"+
				" forwarding index: %v", err)
			select {
			case <-r.quit:
				return
			case <-time.After(time.Second * 5):
				continue
			}
		}

		log.Debugf("(forwarding updates) Fetched last forwarding index(%v)",
			lastIndex)
		break
	}

	events, err := fetchForwardingPayments(r.client, lastIndex)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("(forwarding updates) unable to get forwarding"+
			" events, index(%v): %v",
			lastIndex, err)
		return
	}

	// Avoid db usage if possible.
	if len(events) == 0 {
		return
	}

	log.Infof("(forwarding updates) Broadcast %v forwarding"+
		" events", len(events))
	for _, event := range events {
		sender, err := getPubKeyByChainID(r.client, event.ChanIdIn)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(forwarding updates) error during event"+
				" broadcasting: %v", err)
			continue
		}

		receiver, err := getPubKeyByChainID(r.client, event.ChanIdOut)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(forwarding updates) error during event"+
				" broadcasting: %v", err)
			continue
		}

		if err := r.cfg.Storage.StorePayment(&router.Payment{
			FromUser:  router.UserID(sender),
			ToUser:    router.UserID(receiver),
			FromAlias: registry.GetAlias(router.UserID(sender)),
			ToAlias:   registry.GetAlias(router.UserID(receiver)),
			Amount:    router.BalanceUnit(event.AmtOut),
			Type:      router.Forward,
			Status:    router.Successful,
			Time:      time.Now().Unix(),
		}); err != nil {
			log.Errorf("unable to save the payment: %v", err)
			continue
		}

		update := &router.UpdatePayment{
			Type:     router.Forward,
			Status:   router.Successful,
			Sender:   router.UserID(sender),
			Receiver: router.UserID(receiver),
			Amount:   router.BalanceUnit(event.AmtOut),
			Earned:   router.BalanceUnit(event.Fee),
		}

		// Send update notification to all listeners.
		log.Debugf("(forwarding updates) Send forwarding update: %v"+
			"", spew.Sdump(update))
		r.broadcaster.Write(update)
	}

	lastIndex += uint32(len(events))

	// Try to update the last forwarding index until this operation will
	// be successful.
	for {
		log.Debugf("(forwarding updates) Save last forwarding index"+
			" %v...", lastIndex)
		if err := r.cfg.Storage.PutLastForwardingIndex(lastIndex); err != nil {
			m.AddError(metrics.HighSeverity)
			log.Errorf("(forwarding updates) unable to get last"+
				" forwarding index: %v", err)

			select {
			case <-r.quit:
				return
			case <-time.After(time.Second * 5):
				continue
			}
		}

		break
	}
}

// listenIncomingPayments listener of the incoming lightning payments.
//
// NOTE: Should run as goroutine.
func (r *Router) listenIncomingPayments() {
	defer func() {
		panicRecovering()
		log.Info("Stopped sync incoming payments goroutine")
		r.wg.Done()
	}()

	log.Info("Started sync incoming payments goroutine")

	m := crypto.NewMetric(r.cfg.Asset, "ListenIncomingPayments",
		r.cfg.MetricsBackend)

	var invoiceSubsc lnrpc.Lightning_SubscribeInvoicesClient
	var err error

	for {
		select {
		case <-r.quit:
			return
		default:
		}

		// Initialise invoice subscription client, and if it was null-ed
		// than we should resubscribe on the lnd payment updates,
		// this might be because of connection with lnd was lost.
		if invoiceSubsc == nil {
			log.Info("(payments updates) Trying to subscribe on payment" +
				" updates...")

			invoiceSubsc, err = r.client.SubscribeInvoices(context.Background(),
				&lnrpc.InvoiceSubscription{})
			if err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(payments updates) unable to re-subscribe on"+
					" invoice updates"+
					": %v", err)

				<-time.After(time.Second * 5)
				continue
			}
		}

		log.Info("(payments updates) Waiting for invoice update" +
			" notification...")
		invoiceUpdate, err := invoiceSubsc.Recv()
		if err != nil {
			m.AddError(metrics.HighSeverity)
			// Re-subscribe on lnd invoice updates by making subscription
			// equal to nil.
			log.Errorf("(payments updates) unable to receive: %v", err)
			invoiceSubsc = nil
			continue
		}

		if !invoiceUpdate.Settled {
			log.Infof("(payments updates) Received add invoice update, "+
				"invoice(%v)",
				invoiceUpdate.PaymentRequest)
			continue
		}

		amount := btcutil.Amount(invoiceUpdate.Value)

		if err := r.cfg.Storage.StorePayment(&router.Payment{
			// TODO(andrew.shvv) Need to add sender chan id in lnd,
			// in this case we could understand from which user we have
			// received the payment.
			FromUser:  router.UserID("unknown"),
			ToUser:    r.routerUserID,
			FromAlias: registry.GetAlias("unknown"),
			ToAlias:   registry.GetAlias(r.routerUserID),
			Amount:    router.BalanceUnit(amount),
			Type:      router.Incoming,
			Status:    router.Successful,
			Time:      time.Now().Unix(),
		}); err != nil {
			log.Errorf("unable to save the payment: %v", err)
			continue
		}

		update := &router.UpdatePayment{
			Type:   router.Incoming,
			Status: router.Successful,

			// TODO(andrew.shvv) Need to add sender chan id in lnd,
			// in this case we could understand from which user we have
			// received the payment.
			Sender:   "unknown",
			Receiver: r.routerUserID,

			Amount: router.BalanceUnit(amount),
			Earned: 0,
		}

		// Send update notification to all router updates listeners.
		r.broadcaster.Write(update)
	}
}

func (r *Router) updatePeers() {
	defer func() {
		log.Info("Stopped peer updates goroutine")
		r.wg.Done()
	}()

	log.Info("Started peer updates goroutine")

	for {
		select {
		case <-time.After(time.Second * 5):
		case <-r.quit:
			return
		}

		m := crypto.NewMetric(r.cfg.Asset, "UpdatePeers", r.cfg.MetricsBackend)

		// Fetch users who are connected to us with tcp/ip connection
		connectedPeers, err := fetchUsers(r.client)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch lnd users: %v", err)
			continue
		}

		// Fetch channels and filters only those peers who are connected to us
		// with payment channels, also specify is the peer active or not.
		openChannels, pendingOpenChannels, pendingClosingChannels,
		pendingForceClosingChannels, pendingWaitingCloseChannels, err := fetchChannels(r.client)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch channels: %v", err)
			continue
		}

		hubUsers, err := r.Users()
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch router users: %v", err)
			continue
		}

		hubChannels, err := r.Channels()
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch router channels: %v",
				err)
			continue
		}

		if err := syncPeers(r.cfg.Storage, hubUsers, hubChannels,
			connectedPeers, openChannels, pendingOpenChannels, pendingClosingChannels,
			pendingForceClosingChannels, pendingWaitingCloseChannels,
			r.broadcaster);
			err != nil {

			m.AddError(metrics.HighSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to sync users: %v", err)
			continue
		}

		m.Finish()
	}
}

type lndUser struct {
	userID       router.UserID
	isConnected  bool
	lockedByUser router.BalanceUnit
	lockedByHub  router.BalanceUnit
}

// syncPeers sync information about is the user who is connected to us with
// payment channel now active i.e has tcp/ip connection with us.
// With this we could track what percentage of overall time users are
// connected to our hub.
func syncPeers(userStorage router.UserStorage,
	hubUsers []*router.User,
	hubChannels []*router.Channel,
	connectedPeers []*lnrpc.Peer,
	openChannels []*lnrpc.Channel,
	pendingOpenChannels []*lnrpc.PendingChannelsResponse_PendingOpenChannel,
	pendingClosingChannels []*lnrpc.PendingChannelsResponse_ClosedChannel,
	pendingForceClosingChannels []*lnrpc.PendingChannelsResponse_ForceClosedChannel,
	pendingWaitingCloseChannels []*lnrpc.PendingChannelsResponse_WaitingCloseChannel,
	broadcaster *broadcast.Broadcaster) error {

	if len(connectedPeers) == 0 {
		return nil
	}

	// Create map of peers so that easily check presence of user with given id.
	connectedPeersMap := make(map[router.UserID]*lnrpc.Peer)
	for _, peer := range connectedPeers {
		connectedPeersMap[router.UserID(peer.PubKey)] = peer
	}

	hubChannelView := make(map[router.ChannelID]*router.Channel)
	for _, channel := range hubChannels {
		hubChannelView[channel.ChannelID] = channel
	}

	// This is map of users who are connected to us with payment channels.
	lndPeersView := make(map[router.UserID]*lndUser)

	for _, c := range openChannels {
		userID := router.UserID(c.RemotePubkey)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[router.ChannelID(c.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := router.BalanceUnit(c.RemoteBalance)
		lockedByHub := router.BalanceUnit(c.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedByUser += lockedByUser
			user.lockedByHub += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:       userID,
				isConnected:  isConnected,
				lockedByHub:  lockedByHub,
				lockedByUser: lockedByUser,
			}
		}
	}

	for _, c := range pendingOpenChannels {
		userID := router.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[router.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := router.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := router.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedByUser += lockedByUser
			user.lockedByHub += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:       userID,
				isConnected:  isConnected,
				lockedByHub:  lockedByHub,
				lockedByUser: lockedByUser,
			}
		}
	}

	for _, c := range pendingClosingChannels {
		userID := router.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[router.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := router.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := router.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedByUser += lockedByUser
			user.lockedByHub += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:       userID,
				isConnected:  isConnected,
				lockedByHub:  lockedByHub,
				lockedByUser: lockedByUser,
			}
		}
	}

	for _, c := range pendingForceClosingChannels {
		userID := router.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[router.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := router.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := router.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedByUser += lockedByUser
			user.lockedByHub += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:       userID,
				isConnected:  isConnected,
				lockedByHub:  lockedByHub,
				lockedByUser: lockedByUser,
			}
		}
	}

	for _, c := range pendingWaitingCloseChannels {
		userID := router.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[router.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := router.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := router.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedByUser += lockedByUser
			user.lockedByHub += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:       userID,
				isConnected:  isConnected,
				lockedByHub:  lockedByHub,
				lockedByUser: lockedByUser,
			}
		}
	}

	hubPeersView := make(map[router.UserID]*router.User)
	for _, user := range hubUsers {
		hubPeersView[user.UserID] = user
	}

	for userID, lndUser := range lndPeersView {
		// If we have this is in database and it exist in list peer,
		// than we should set this used as connected.
		hubUser, ok := hubPeersView[userID]
		if ok {
			// In this case we had this user before,
			// and we need update information about him,
			// and also if his connection status has changed,
			// than we need to sen update.
			hubUser.LockedByUser = lndUser.lockedByUser
			hubUser.LockedByHub = lndUser.lockedByHub

			if hubUser.IsConnected != lndUser.isConnected {
				hubUser.IsConnected = lndUser.isConnected
				broadcaster.Write(&router.UpdateUserConnected{
					User:        userID,
					IsConnected: hubUser.IsConnected,
				})
			}

			if err := hubUser.Save(); err != nil {
				return errors.Errorf("unable update user status: %v", err)
			}
		} else {
			newUser, err := router.NewUser(userID, lndUser.isConnected,
				registry.GetAlias(userID), lndUser.lockedByHub,
				lndUser.lockedByUser, &router.UserConfig{
					Storage: userStorage,
				})
			if err != nil {
				return errors.Errorf("unable to create user: %v", err)
			}

			// In this case we have new user, which wasn't seen before,
			// and we need to save it.
			log.Infof("Save new user(%v)", lndUser.userID)
			if err := newUser.Save(); err != nil {
				return errors.Errorf("unable update user status: %v", err)
			}

			broadcaster.Write(&router.UpdateUserConnected{
				User:        userID,
				IsConnected: lndUser.isConnected,
			})
		}
	}

	for userID, hubUser := range hubPeersView {
		if _, ok := lndPeersView[userID]; ok {
			// 	Already covered in previously cycle.
			continue
		}

		// If we can't see this user in lnd view,
		// so he is not connected to us with payment channel.
		// As far as we send update only about hubUsers who has channels
		// with us, we need to mark this user as disconnected.
		if hubUser.IsConnected {
			hubUser.IsConnected = false
			broadcaster.Write(&router.UpdateUserConnected{
				User:        userID,
				IsConnected: hubUser.IsConnected,
			})

			if err := hubUser.Save(); err != nil {
				return errors.Errorf("unable update user status: %v", err)
			}
		}
	}

	return nil
}

// TODO(andrew.shvv) it will not work after dual funding, remove it
func getInitiator(routerBalance int64) router.ChannelInitiator {
	if routerBalance == 0 {
		return router.UserInitiator
	} else {
		return router.RouterInitiator
	}
}

// listenOutgoingPayments listens for outgoing payment from out lightning
// network node.
//
// NOTE: Should run as goroutine.
func (r *Router) listenOutgoingPayments() {
	defer func() {
		log.Info("Stopped sync outgoing payments goroutine")
		r.wg.Done()
	}()

	log.Info("Started sync outgoing payments goroutine")

	for {
		select {
		case <-time.After(time.Second * 15):
		case <-r.quit:
			return
		}

		m := crypto.NewMetric(r.cfg.Asset, "SyncOutgoingPayments", r.cfg.MetricsBackend)

		payments, err := fetchOutgoingPayments(r.client)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			m.Finish()

			log.Errorf("(outgoing payments sync) unable to get lnd outgoing"+
				" payments: %v", err)
			continue
		}

		if err := syncOutgoing(r.routerUserID, payments, r.cfg.Storage,
			r.cfg.Storage, r.broadcaster); err != nil {
			m.AddError(metrics.HighSeverity)
			m.Finish()

			log.Errorf("(outgoing payments sync) unable to sync outgoing"+
				" payments: %v", err)
			continue
		}

		m.Finish()
	}
}

// syncOutgoing synchronise lightning network node view on outgoing payments
// with out state, and send payment updates.
func syncOutgoing(routerID router.UserID, lndPayments []*lnrpc.Payment,
	syncStorage SyncStorage, paymentStorage router.PaymentStorage,
	broadcaster *broadcast.Broadcaster) error {

	lastOutgoingPaymentTime, err := syncStorage.LastOutgoingPaymentTime()
	if err != nil {
		return err
	}

	log.Debugf("(outgoing payments sync) Fetched last outgoing payment"+
		" time(%v)", lastOutgoingPaymentTime)

	for _, payment := range lndPayments {
		// Skip entries which are already been proceeded.
		if payment.CreationDate <= lastOutgoingPaymentTime {
			continue
		}

		// Save last time before actually save and send update, because cost
		// of not having this payment in db is less that having two such
		// payments.
		lastOutgoingPaymentTime = payment.CreationDate
		err := syncStorage.PutLastOutgoingPaymentTime(lastOutgoingPaymentTime)
		if err != nil {
			return err
		}

		sender := routerID
		receiver := router.UserID(payment.Path[0])
		amount := router.BalanceUnit(payment.Value)
		fee := router.BalanceUnit(payment.Fee)

		log.Infof("(outgoing payments sync) Process new outgoing payment"+
			" from(%v), to(%v), amount(%v), fee(%v), time(%v)", sender,
			receiver, amount, fee, payment.CreationDate)

		broadcaster.Write(&router.UpdatePayment{
			Type:     router.Outgoing,
			Status:   router.Successful,
			Sender:   sender,
			Receiver: receiver,
			Amount:   amount,
			Earned:   -fee,
		})

		if err := paymentStorage.StorePayment(&router.Payment{
			FromUser:  sender,
			ToUser:    receiver,
			FromAlias: registry.GetAlias(sender),
			ToAlias:   registry.GetAlias(receiver),
			Amount:    amount,
			Type:      router.Outgoing,
			Status:    router.Successful,
			Time:      payment.CreationDate,
		}); err != nil {
			log.Errorf("unable to save the payment: %v", err)
			continue
		}
	}

	return nil
}
