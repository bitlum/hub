package lnd

import (
	"time"
	"github.com/lightningnetwork/lnd/lnrpc"
	"context"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/bitlum/hub/manager/metrics/crypto"
	"github.com/bitlum/hub/manager/metrics"
)

// updateNodeInfo updates information about last synced version of the
// lightning node.
func (r *Router) updateNodeInfo() {
	r.wg.Add(1)
	go func() {
		defer func() {
			log.Info("Stopped lightning node info updates goroutine")
			r.wg.Done()
		}()

		log.Info("Started lightning node info updates goroutine")

		for {
			select {
			case <-time.After(time.Second * 5):
			case <-r.quit:
				return
			}

			reqInfo := &lnrpc.GetInfoRequest{}
			ctx, _ := context.WithTimeout(getContext(), time.Second*5)
			respInfo, err := r.client.GetInfo(ctx, reqInfo)
			if err != nil {
				log.Errorf("unable get lnd node info: %v", err)
				continue
			}

			if err := r.cfg.InfoStorage.UpdateInfo(&router.DbInfo{
				Version:     respInfo.Version,
				Network:     r.cfg.Net,
				BlockHeight: respInfo.BlockHeight,
				BlockHash:   respInfo.BlockHash,
				NodeInfo: &router.DbNodeInfo{
					Alias:          respInfo.Alias,
					Host:           r.cfg.PeerHost,
					Port:           r.cfg.PeerPort,
					IdentityPubKey: respInfo.IdentityPubkey,
				},
				NeutrinoInfo: &router.DbNeutrinoInfo{
					Host: r.cfg.NeutrinoHost,
					Port: r.cfg.NeutrinoPort,
				},
			}); err != nil {
				log.Errorf("unable to save lightning node info: %v", err)
			}

		}
	}()
}

// listenLocalTopologyUpdates tracks the local channel topology state updates
// and sends notifications accordingly to the occurred transition.
func (r *Router) listenLocalTopologyUpdates() {
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

		r.syncTopologyUpdates()
	}
}

// syncTopologyUpdates is used as wrapper for fetching and syncing topology
// updates. Main purpose of creating distinct method was usage of defer, which
// is needed mainly for metric gathering usage.
func (r *Router) syncTopologyUpdates() {
	defer panicRecovering()

	m := crypto.NewMetric(r.cfg.Asset, "SyncTopologyUpdates",
		r.cfg.MetricsBackend)
	defer m.Finish()

	// Take all pending/closing/opened channels and form the current
	// state out of it. As far as those two operation are not atomic,
	// it might happen that some channel sleeps away.
	reqPending := &lnrpc.PendingChannelsRequest{}
	respPending, err := r.client.PendingChannels(getContext(), reqPending)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("(topology updates) unable to fetch pending"+
			" channels: %v", err)
		return
	}

	reqOpen := &lnrpc.ListChannelsRequest{}
	respOpen, err := r.client.ListChannels(getContext(), reqOpen)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("(topology updates) unable to fetch list channels"+
			": %v", err)
		return
	}

	// Fetch prev pending/closing/opened channels from db which
	// corresponds to the old/previous channel state.
	log.Debugf("(topology updates) Fetching channel state from db...")
	channels, err := r.Network()
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("(topology updates) unable to fetch old channel state: %v"+
			"", err)
		return
	}

	routerChannelMap := make(map[router.ChannelID]*router.Channel)
	for _, c := range channels {
		routerChannelMap[c.ChannelID] = c
	}

	// TODO(andrew.shvv) it will not work after dual funding, remove it
	getInitiator := func(routerBalance, userBalance int64) router.ChannelInitiator {
		if routerBalance == 0 {
			return router.UserInitiator
		} else {
			return router.RouterInitiator
		}
	}

	// Keep new channel ids to detect channel closes at later point.
	newChannelIDs := make(map[router.ChannelID]struct{}, 0)

	for _, newChannel := range respPending.PendingOpenChannels {
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
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) impossible channel change state "+
				"%v => %v", router.ChannelOpened, router.ChannelOpening)
			continue
		case router.ChannelClosing:
			// Previous channel state was closing, and now it is
			// opening, we couldn't  re-open, closing channel.
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) impossible channel change state "+
				"%v => %v", router.ChannelClosing, router.ChannelOpening)
			continue
		case router.ChannelClosed:
			// Previous channel state was closing, and now it is
			// opening, we couldn't  re-open, closing channel.
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) impossible channel change state "+
				"%v => %v", router.ChannelClosed, router.ChannelOpening)
			continue
		case "not exist":
			cfg := &router.ChannelConfig{
				Broadcaster: r.broadcaster,
				Storage:     r.cfg.SyncStorage,
			}

			channel, err = router.NewChannel(
				router.ChannelID(newChannel.Channel.ChannelPoint),
				router.UserID(newChannel.Channel.RemoteNodePub),
				router.BalanceUnit(newChannel.Channel.Capacity),
				router.BalanceUnit(newChannel.Channel.RemoteBalance),
				router.BalanceUnit(newChannel.Channel.LocalBalance),
				router.BalanceUnit(newChannel.CommitFee),
				getInitiator(newChannel.Channel.LocalBalance,
					newChannel.Channel.RemoteBalance),
				cfg,
			)
			if err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to create new channel"+
					": %v", err)
				continue
			}


			if err := channel.Save(); err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to save new channel"+
					": %v", err)
				continue
			}

			if err := channel.SetOpeningState(); err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to set channel openning"+
					": %v", err)
				continue
			}
		}
	}

	for _, newChannel := range respPending.PendingClosingChannels {
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

			if err := channel.SetClosingState(); err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable to set closing state"+
					": %v", err)
				continue
			}

		case router.ChannelOpened:
			if err := channel.SetClosingState(); err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable to set closing state"+
					": %v", err)
				continue
			}
		case router.ChannelClosing:
			// Nothing has changed
		case router.ChannelClosed:
			// Previous channel state was closed, and now it is
			// closing, we couldn't  re-close, closed channel.
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) impossible channel change state "+
				"%v => %v", router.ChannelClosed, router.ChannelClosing)
			continue
		case "not exist":
			// Previously channel not existed, it seems that because of the
			// delayed scrape we missed some of state changes.
			cfg := &router.ChannelConfig{
				Broadcaster: r.broadcaster,
				Storage:     r.cfg.SyncStorage,
			}

			channel, err = router.NewChannel(
				router.ChannelID(newChannel.Channel.ChannelPoint),
				router.UserID(newChannel.Channel.RemoteNodePub),
				router.BalanceUnit(newChannel.Channel.Capacity),
				router.BalanceUnit(newChannel.Channel.RemoteBalance),
				router.BalanceUnit(newChannel.Channel.LocalBalance),
				0,
				getInitiator(newChannel.Channel.LocalBalance,
					newChannel.Channel.RemoteBalance),
				cfg,
			)
			if err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to create new channel"+
					": %v", err)
				continue
			}

			if err := channel.Save(); err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to save new channel"+
					": %v", err)
				continue
			}

			if err := channel.SetClosingState(); err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to set channel closing"+
					": %v", err)
				continue
			}
		}
	}

	for _, newChannel := range respPending.PendingForceClosingChannels {
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
			if err := channel.SetClosingState(); err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable to set closing state"+
					": %v", err)
				continue
			}

		case router.ChannelOpened:
			if err := channel.SetClosingState(); err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable to set closing state"+
					": %v", err)
				continue
			}
		case router.ChannelClosing:
			// Nothing has changed
		case router.ChannelClosed:
			// Previous channel state was closed, and now it is
			// closing, we couldn't  re-close, closed channel.
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) impossible channel change state "+
				"%v => %v", router.ChannelClosed, router.ChannelClosing)
			continue
		case "not exist":
			// Previously channel not existed, it seems that because of the
			// delayed scrape we missed some of state changes.
			cfg := &router.ChannelConfig{
				Broadcaster: r.broadcaster,
				Storage:     r.cfg.SyncStorage,
			}

			channel, err = router.NewChannel(
				router.ChannelID(newChannel.Channel.ChannelPoint),
				router.UserID(newChannel.Channel.RemoteNodePub),
				router.BalanceUnit(newChannel.Channel.Capacity),
				router.BalanceUnit(newChannel.Channel.RemoteBalance),
				router.BalanceUnit(newChannel.Channel.LocalBalance),
				0,
				getInitiator(newChannel.Channel.LocalBalance,
					newChannel.Channel.RemoteBalance),
				cfg,
			)
			if err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to create new channel"+
					": %v", err)
				continue
			}

			if err := channel.Save(); err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to save new channel"+
					": %v", err)
				continue
			}

			if err := channel.SetClosingState(); err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable to set closing state"+
					": %v", err)
				continue
			}
		}
	}

	for _, newChannel := range respOpen.Channels {
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
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable set channel state: %v", err)
				continue
			}

		case router.ChannelOpened:
			// Nothing has changed
		case router.ChannelClosing:
			// Previous channel state was closing, but now it is
			// opened, close couldn't be canceled.
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) impossible channel change state "+
				"%v => %v", router.ChannelClosing, router.ChannelOpened)
			continue
		case router.ChannelClosed:
			// Previous channel state was closed, but now it is
			// opened, close couldn't be canceled.
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) impossible channel change state "+
				"%v => %v", router.ChannelClosed, router.ChannelOpened)
			continue
		case "not exist":
			cfg := &router.ChannelConfig{
				Broadcaster: r.broadcaster,
				Storage:     r.cfg.SyncStorage,
			}

			channel, err = router.NewChannel(
				router.ChannelID(newChannel.ChannelPoint),
				router.UserID(newChannel.RemotePubkey),
				router.BalanceUnit(newChannel.Capacity),
				router.BalanceUnit(newChannel.RemoteBalance),
				router.BalanceUnit(newChannel.LocalBalance),
				router.BalanceUnit(newChannel.CommitFee),
				getInitiator(newChannel.LocalBalance,
					newChannel.RemoteBalance),
				cfg,
			)
			if err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to create new channel"+
					": %v", err)
				continue
			}

			if err := channel.Save(); err != nil {
				m.AddError(metrics.MiddleSeverity)
				log.Errorf("(topology updates) unable to save new channel"+
					": %v", err)
				continue
			}

			if err := channel.SetOpenedState(); err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable save opened channel"+
					" state: %v", err)
				continue
			}
		}
	}

	for chanID, channel := range routerChannelMap {
		_, ok := newChannelIDs[chanID]
		if ok {
			// It was already handled previously
			continue
		}

		switch channel.CurrentState().Name {
		case router.ChannelOpening,
			router.ChannelOpened,
			router.ChannelClosing,
			router.ChannelClosed:

			if err := channel.SetClosedState(); err != nil {
				m.AddError(metrics.HighSeverity)
				log.Errorf("(topology updates) unable set closed channel"+
					" state: %v", err)
				continue
			}
		}
	}
}

// listenForwardingUpdates listener of the forwarding lightning payments.
// Fetches the lnd forwarding log, and send new update to the broadcaster.
func (r *Router) listenForwardingUpdates() {
	defer func() {
		log.Info("Stopped forwarding payments goroutine")
		r.wg.Done()
	}()

	log.Info("Started forwarding payments goroutine")

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

	m := crypto.NewMetric(r.cfg.Asset, "SyncForwardingUpdates",
		r.cfg.MetricsBackend)
	defer m.Finish()

	// Try to fetch last index, and if fails than try after yet again after
	// some time.
	for {
		lastIndex, err = r.cfg.SyncStorage.LastForwardingIndex()
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

	events, err := r.getNewForwardingEvents(lastIndex)
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
		sender, err := r.getPubKeyByChainID(event.ChanIdIn)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(forwarding updates) error during event"+
				" broadcasting: %v", err)
			continue
		}

		receiver, err := r.getPubKeyByChainID(event.ChanIdOut)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(forwarding updates) error during event"+
				" broadcasting: %v", err)
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
		if err := r.cfg.SyncStorage.PutLastForwardingIndex(lastIndex); err != nil {
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

// getPubKeyByChainID returns the pubkey which identifies the user by the
// given channel id.
func (r *Router) getPubKeyByChainID(chanID uint64) (string, error) {
	req := &lnrpc.ListChannelsRequest{}
	resp, err := r.client.ListChannels(getContext(), req)
	if err != nil {
		return "", err
	}

	var pubKey string
	for _, channel := range resp.Channels {
		if channel.ChanId == chanID {
			pubKey = channel.RemotePubkey
			break
		}
	}

	if pubKey == "" {
		return "", errors.Errorf("unable to find node by chan id(%v)",
			chanID)
	}

	return pubKey, nil
}

// getNewForwardingEvents gradually fetches the forwarding events from lightning
// daemon.
func (r *Router) getNewForwardingEvents(index uint32) (
	[]*lnrpc.ForwardingEvent, error) {

	var events []*lnrpc.ForwardingEvent
	var limit uint32 = 1000

	// Fetch updates by chunks, in order to avoid message
	// overflow errors, lnd error response is restricted to ~50k updates.
	for {
		req := &lnrpc.ForwardingHistoryRequest{
			StartTime:    1,
			EndTime:      uint64(time.Now().Unix()),
			IndexOffset:  index,
			NumMaxEvents: limit,
		}

		resp, err := r.client.ForwardingHistory(getContext(), req)
		if err != nil {
			return nil, err
		}

		for _, event := range resp.ForwardingEvents {
			events = append(events, event)
		}

		length := uint32(len(resp.ForwardingEvents))
		index += length

		// If daemon returned less than a limit it means that we reached the
		// end of the forwarding list.
		if length < limit {
			break
		}
	}

	return events, nil
}

// listenInvoiceUpdates listener of the incoming lightning payments. If lnd was
// disconnected listener will be re-subscribe on updates when lnd goes live
// and continue listening.
func (r *Router) listenInvoiceUpdates() {
	defer func() {
		panicRecovering()
		log.Info("Stopped incoming payments goroutine")
		r.wg.Done()
	}()

	log.Info("Started incoming payments goroutine")

	m := crypto.NewMetric(r.cfg.Asset, "ListenInvoiceUpdates",
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
		update := &router.UpdatePayment{
			Type:   router.Incoming,
			Status: router.Successful,

			// TODO(andrew.shvv) Need to add sender chan id in lnd,
			// in this case we could understand from which user we have
			// received the payment.
			Sender:   "unknown",
			Receiver: router.UserID(r.nodeAddr),

			Amount: router.BalanceUnit(amount),
			Earned: 0,
		}

		// Send update notification to all listeners.
		r.broadcaster.Write(update)
	}
}
