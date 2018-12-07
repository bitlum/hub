package lnd

import (
	"context"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/hub/common/broadcast"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/bitlum/hub/registry"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"time"
)

// updateNodeInfo updates information about last synced version of the
// lightning node, if version has changed or best block has updated,
// we need to know that.
//
// NOTE: Should run as goroutine.
func (client *Client) updateNodeInfo() {
	defer func() {
		log.Info("Stopped lightning node info updates goroutine")
		client.wg.Done()
	}()

	log.Info("Started lightning node info updates goroutine")

	for {
		select {
		case <-time.After(time.Second * 25):
		case <-client.quit:
			return
		}

		m := crypto.NewMetric(client.cfg.Asset, "UpdateNodeInfo", client.cfg.MetricsBackend)
		nodeInfo, err := fetchNodeInfo(client.client)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("unable to fetch node info: %v", err)
			continue
		}
		m.Finish()

		if err := client.cfg.Storage.UpdateInfo(&lightning.Info{
			Version:     nodeInfo.Version,
			Network:     client.cfg.Net,
			BlockHeight: nodeInfo.BlockHeight,
			BlockHash:   nodeInfo.BlockHash,
			NodeInfo: &lightning.NodeInfo{
				Alias:          nodeInfo.Alias,
				Host:           client.cfg.PeerHost,
				Port:           client.cfg.PeerPort,
				IdentityPubKey: nodeInfo.IdentityPubkey,
			},
			NeutrinoInfo: &lightning.NeutrinoInfo{
				Host: client.cfg.NeutrinoHost,
				Port: client.cfg.NeutrinoPort,
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
func (client *Client) updateChannelStates() {
	defer func() {
		log.Info("Stopped local topology updates goroutine")
		client.wg.Done()
	}()

	log.Info("Started local topology updates goroutine")

	for {
		select {
		case <-time.After(time.Second * 5):
		case <-client.quit:
			return
		}

		m := crypto.NewMetric(client.cfg.Asset, "UpdateChannelStates", client.cfg.MetricsBackend)

		// TODO(andrew.shvv) track waiting closing channels
		openChannels, pendingOpenChannels, pendingClosingChannels,
		pendingForceClosingChannels, _, err := fetchChannels(client.client)
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

		nodeChannels, err := client.Channels()
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(topology updates) unable to fetch old channel state: %v"+
				"", err)
			continue
		}

		if err := syncChannelStates(openChannels, pendingOpenChannels,
			pendingClosingChannels, pendingForceClosingChannels, client.broadcaster,
			client.cfg.Storage, client.cfg.Storage, nodeChannels); err != nil {
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
	channelStorage lightning.ChannelStorage,
	indexStorage IndexesStorage,
	nodeState []*lightning.Channel) error {

	nodeChannelMap := make(map[lightning.ChannelID]*lightning.Channel)
	for _, c := range nodeState {
		nodeChannelMap[c.ChannelID] = c
	}

	// Keep new channel ids to detect channel closes at later point.
	newChannelIDs := make(map[lightning.ChannelID]struct{}, 0)

	for _, newChannel := range pendingOpenChannels {
		chanID := lightning.ChannelID(newChannel.Channel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		state := lightning.ChannelStateName("not exist")
		channel, ok := nodeChannelMap[lightning.ChannelID(newChannel.Channel.ChannelPoint)]
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case lightning.ChannelOpening:
			// Nothing has changed
		case lightning.ChannelOpened:
			// Previous channel state was opened, and now it is again
			// opening, something wrong has happened.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, lightning.ChannelOpened,
				lightning.ChannelOpening)

		case lightning.ChannelClosing:
			// Previous channel state was closing, and now it is
			// opening, we couldn't  re-open, closing channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, lightning.ChannelClosing,
				lightning.ChannelOpening)

		case lightning.ChannelClosed:
			// Previous channel state was closing, and now it is
			// opening, we couldn't  re-open, closing channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, lightning.ChannelClosed,
				lightning.ChannelOpening)

		case "not exist":
			cfg := &lightning.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := lightning.NewChannel(
				lightning.ChannelID(newChannel.Channel.ChannelPoint),
				lightning.UserID(newChannel.Channel.RemoteNodePub),
				lightning.BalanceUnit(newChannel.Channel.Capacity),
				lightning.BalanceUnit(newChannel.Channel.RemoteBalance),
				lightning.BalanceUnit(newChannel.Channel.LocalBalance),
				lightning.BalanceUnit(newChannel.CommitFee),
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
		chanID := lightning.ChannelID(newChannel.Channel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		channel, ok := nodeChannelMap[lightning.ChannelID(newChannel.Channel.ChannelPoint)]
		state := lightning.ChannelStateName("not exist")
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case lightning.ChannelOpening:
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

		case lightning.ChannelOpened:
			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set closing state"+
					"for channel(%v): %v", chanID, err)
			}
		case lightning.ChannelClosing:
			// Nothing has changed
		case lightning.ChannelClosed:
			// Previous channel state was closed, and now it is
			// closing, we couldn't  re-close, closed channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, lightning.ChannelClosed,
				lightning.ChannelClosing)

		case "not exist":
			// Previously channel not existed, it seems that because of the
			// delayed scrape we missed some of state changes.
			cfg := &lightning.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := lightning.NewChannel(
				lightning.ChannelID(newChannel.Channel.ChannelPoint),
				lightning.UserID(newChannel.Channel.RemoteNodePub),
				lightning.BalanceUnit(newChannel.Channel.Capacity),
				lightning.BalanceUnit(newChannel.Channel.RemoteBalance),
				lightning.BalanceUnit(newChannel.Channel.LocalBalance),
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
		chanID := lightning.ChannelID(newChannel.Channel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		state := lightning.ChannelStateName("not exist")
		channel, ok := nodeChannelMap[lightning.ChannelID(newChannel.Channel.ChannelPoint)]
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case lightning.ChannelOpening:
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

		case lightning.ChannelOpened:
			if err := channel.SetClosingState(); err != nil {
				return errors.Errorf("unable to set closing state"+
					"for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case lightning.ChannelClosing:
			// Nothing has changed
		case lightning.ChannelClosed:
			// Previous channel state was closed, and now it is
			// closing, we couldn't  re-close, closed channel.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, lightning.ChannelClosed,
				lightning.ChannelClosing)

		case "not exist":
			// Previously channel not existed, it seems that because of the
			// delayed scrape we missed some of state changes.
			cfg := &lightning.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := lightning.NewChannel(
				lightning.ChannelID(newChannel.Channel.ChannelPoint),
				lightning.UserID(newChannel.Channel.RemoteNodePub),
				lightning.BalanceUnit(newChannel.Channel.Capacity),
				lightning.BalanceUnit(newChannel.Channel.RemoteBalance),
				lightning.BalanceUnit(newChannel.Channel.LocalBalance),
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
		chanID := lightning.ChannelID(newChannel.ChannelPoint)
		newChannelIDs[chanID] = struct{}{}

		// Save user id <> short channel id index in order to retrieve it
		// later on payment notifications without query lightning network
		// daemon.
		userID := lightning.UserID(newChannel.RemotePubkey)
		if err := indexStorage.AddUserIDToShortChanIDIndex(userID,
			newChannel.ChanId); err != nil {
			return errors.Errorf("unable to add user_id("+
				"%v) <> short_channel_id(%v) index: %v", userID,
				newChannel.ChanId, err)
		}

		// Save channel point <> short channel id index in order to retrieve it
		// later on payment notifications without query lightning network
		// daemon.
		if err := indexStorage.AddChannelPointToShortChanIDIndex(chanID,
			newChannel.ChanId); err != nil {
			return errors.Errorf("unable to add channel_point("+
				"%v) <> short_channel_id(%v) index: %v", chanID,
				newChannel.ChanId, err)
		}

		channel, ok := nodeChannelMap[lightning.ChannelID(newChannel.ChannelPoint)]
		state := lightning.ChannelStateName("not exist")
		if ok {
			state = channel.CurrentState().Name
		}

		switch state {
		case lightning.ChannelOpening:
			if err := channel.SetOpenedState(); err != nil {
				return errors.Errorf("unable set channel state"+
					" opened for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case lightning.ChannelOpened:
			// Nothing has changed
		case lightning.ChannelClosing:
			// Previous channel state was closing, but now it is
			// opened, close couldn't be canceled.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, lightning.ChannelClosing,
				lightning.ChannelOpened)

		case lightning.ChannelClosed:
			// Previous channel state was closed, but now it is
			// opened, close couldn't be canceled.
			return errors.Errorf("impossible channel("+
				"%v) change state %v => %v", chanID, lightning.ChannelClosed,
				lightning.ChannelOpened)

		case "not exist":
			cfg := &lightning.ChannelConfig{
				Broadcaster: broadcaster,
				Storage:     channelStorage,
			}

			channel, err := lightning.NewChannel(
				lightning.ChannelID(newChannel.ChannelPoint),
				lightning.UserID(newChannel.RemotePubkey),
				lightning.BalanceUnit(newChannel.Capacity),
				lightning.BalanceUnit(newChannel.RemoteBalance),
				lightning.BalanceUnit(newChannel.LocalBalance),
				lightning.BalanceUnit(newChannel.CommitFee),
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

	for chanID, channel := range nodeChannelMap {
		_, ok := newChannelIDs[chanID]
		if ok {
			// It was already handled previously
			continue
		}

		state := channel.CurrentState().Name
		switch state {
		case lightning.ChannelOpening:
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

		case lightning.ChannelOpened:
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

		case lightning.ChannelClosing:
			if err := channel.SetClosedState(); err != nil {
				return errors.Errorf("unable set closed channel"+
					" state for channel(%v): %v", chanID, err)
			}

			log.Infof("Set state(%v) for channel(%v)",
				channel.CurrentState(), channel.ChannelID)

		case lightning.ChannelClosed:
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
func (client *Client) listenForwardingPayments() {
	defer func() {
		log.Info("Stopped sync forwarding payments goroutine")
		client.wg.Done()
	}()

	log.Info("Started sync forwarding payments goroutine")

	for {
		select {
		case <-time.After(time.Second * 5):
		case <-client.quit:
			return
		}

		client.syncForwardingUpdate(client.cfg.Storage)
	}
}

// syncForwardingUpdate is used as wrapper for fetching and syncing forwarding
// updates. Main purpose of creating distinct method was usage of defer, which
// is needed mainly for metric gathering usage.
func (client *Client) syncForwardingUpdate(storage IndexesStorage) {
	defer panicRecovering()

	var lastIndex uint32
	var err error

	m := crypto.NewMetric(client.cfg.Asset, "ListenForwardingPayments",
		client.cfg.MetricsBackend)
	defer m.Finish()

	// Try to fetch last index, and if fails than try after yet again after
	// some time.
	for {
		lastIndex, err = client.cfg.Storage.LastForwardingIndex()
		if err != nil {
			m.AddError(metrics.HighSeverity)
			log.Errorf("(forwarding updates) unable to get last"+
				" forwarding index: %v", err)
			select {
			case <-client.quit:
				return
			case <-time.After(time.Second * 5):
				continue
			}
		}

		log.Debugf("(forwarding updates) Fetched last forwarding index(%v)",
			lastIndex)
		break
	}

	events, err := fetchForwardingPayments(client.client, lastIndex)
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
		sender, err := storage.GetUserIDByShortChanID(event.ChanIdIn)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(forwarding updates) error during event"+
				" broadcasting: %v", err)
			continue
		}

		receiver, err := storage.GetUserIDByShortChanID(event.ChanIdOut)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(forwarding updates) error during event"+
				" broadcasting: %v", err)
			continue
		}

		// TODO(andrew.shvv) Add fail attempts
		status := lightning.Successful

		if err := client.cfg.Storage.StorePayment(&lightning.Payment{
			FromUser:  sender,
			ToUser:    receiver,
			FromAlias: registry.GetAlias(sender),
			ToAlias:   registry.GetAlias(receiver),
			Amount:    lightning.BalanceUnit(event.AmtOut),
			Type:      lightning.Forward,
			Status:    status,
			Time:      time.Now().Unix(),
		}); err != nil {
			log.Errorf("unable to save the payment: %v", err)
			continue
		}

		update := &lightning.UpdatePayment{
			Type:     lightning.Forward,
			Status:   status,
			Sender:   sender,
			Receiver: receiver,
			Amount:   lightning.BalanceUnit(event.AmtOut),
			Earned:   lightning.BalanceUnit(event.Fee),
		}

		// Send update notification to all listeners.
		log.Debugf("(forwarding updates) Send forwarding update: %v"+
			"", spew.Sdump(update))
		client.broadcaster.Write(update)
	}

	lastIndex += uint32(len(events))

	// Try to update the last forwarding index until this operation will
	// be successful.
	for {
		log.Debugf("(forwarding updates) Save last forwarding index"+
			" %v...", lastIndex)
		if err := client.cfg.Storage.PutLastForwardingIndex(lastIndex); err != nil {
			m.AddError(metrics.HighSeverity)
			log.Errorf("(forwarding updates) unable to get last"+
				" forwarding index: %v", err)

			select {
			case <-client.quit:
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
func (client *Client) listenIncomingPayments() {
	defer func() {
		panicRecovering()
		log.Info("Stopped sync incoming payments goroutine")
		client.wg.Done()
	}()

	log.Info("Started sync incoming payments goroutine")

	m := crypto.NewMetric(client.cfg.Asset, "ListenIncomingPayments",
		client.cfg.MetricsBackend)

	var invoiceSubsc lnrpc.Lightning_SubscribeInvoicesClient
	var err error

	for {
		select {
		case <-client.quit:
			return
		default:
		}

		// Initialise invoice subscription client, and if it was null-ed
		// than we should resubscribe on the lnd payment updates,
		// this might be because of connection with lnd was lost.
		if invoiceSubsc == nil {
			log.Info("(payments updates) Trying to subscribe on payment" +
				" updates...")

			invoiceSubsc, err = client.client.SubscribeInvoices(context.Background(),
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

		userID := lightning.UserID("unknown")

		amount := btcutil.Amount(invoiceUpdate.Value)
		if err := client.cfg.Storage.StorePayment(&lightning.Payment{
			FromUser:  userID,
			ToUser:    client.lightningNodeUserID,
			FromAlias: registry.GetAlias(userID),
			ToAlias:   registry.GetAlias(client.lightningNodeUserID),
			Amount:    lightning.BalanceUnit(amount),
			Type:      lightning.Incoming,
			Status:    lightning.Successful,
			Time:      time.Now().Unix(),
		}); err != nil {
			log.Errorf("unable to save the payment: %v", err)
			continue
		}

		// Send update notification to all lightning client updates listeners.
		client.broadcaster.Write(&lightning.UpdatePayment{
			Type:     lightning.Incoming,
			Status:   lightning.Successful,
			Sender:   userID,
			Receiver: client.lightningNodeUserID,
			Amount:   lightning.BalanceUnit(amount),
			Earned:   0,
		})
	}
}

func (client *Client) updatePeers() {
	defer func() {
		log.Info("Stopped peer updates goroutine")
		client.wg.Done()
	}()

	log.Info("Started peer updates goroutine")

	for {
		select {
		case <-time.After(time.Second * 5):
		case <-client.quit:
			return
		}

		m := crypto.NewMetric(client.cfg.Asset, "UpdatePeers", client.cfg.MetricsBackend)

		// Fetch users who are connected to us with tcp/ip connection
		connectedPeers, err := fetchUsers(client.client)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch lnd users: %v", err)
			continue
		}

		// Fetch channels and filters only those peers who are connected to us
		// with payment channels, also specify is the peer active or not.
		openChannels, pendingOpenChannels, pendingClosingChannels,
		pendingForceClosingChannels, pendingWaitingCloseChannels, err := fetchChannels(client.client)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch channels: %v", err)
			continue
		}

		hubUsers, err := client.Users()
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch lightning node "+
				"users: %v", err)
			continue
		}

		hubChannels, err := client.Channels()
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			m.Finish()

			log.Errorf("(peer updates) unable to fetch lightning node"+
				" channels: %v", err)
			continue
		}

		if err := syncPeers(client.cfg.Storage, hubUsers, hubChannels,
			connectedPeers, openChannels, pendingOpenChannels, pendingClosingChannels,
			pendingForceClosingChannels, pendingWaitingCloseChannels,
			client.broadcaster);
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
	userID           lightning.UserID
	isConnected      bool
	lockedWithRemote lightning.BalanceUnit
	lockedLocally    lightning.BalanceUnit
}

// syncPeers sync information about is the user who is connected to us with
// payment channel now active i.e has tcp/ip connection with us.
// With this we could track what percentage of overall time users are
// connected to our hub.
func syncPeers(userStorage lightning.UserStorage,
	hubUsers []*lightning.User,
	hubChannels []*lightning.Channel,
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
	connectedPeersMap := make(map[lightning.UserID]*lnrpc.Peer)
	for _, peer := range connectedPeers {
		connectedPeersMap[lightning.UserID(peer.PubKey)] = peer
	}

	hubChannelView := make(map[lightning.ChannelID]*lightning.Channel)
	for _, channel := range hubChannels {
		hubChannelView[channel.ChannelID] = channel
	}

	// This is map of users who are connected to us with payment channels.
	lndPeersView := make(map[lightning.UserID]*lndUser)

	for _, c := range openChannels {
		userID := lightning.UserID(c.RemotePubkey)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[lightning.ChannelID(c.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := lightning.BalanceUnit(c.RemoteBalance)
		lockedByHub := lightning.BalanceUnit(c.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedWithRemote += lockedByUser
			user.lockedLocally += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:           userID,
				isConnected:      isConnected,
				lockedLocally:    lockedByHub,
				lockedWithRemote: lockedByUser,
			}
		}
	}

	for _, c := range pendingOpenChannels {
		userID := lightning.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[lightning.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := lightning.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := lightning.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedWithRemote += lockedByUser
			user.lockedLocally += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:           userID,
				isConnected:      isConnected,
				lockedLocally:    lockedByHub,
				lockedWithRemote: lockedByUser,
			}
		}
	}

	for _, c := range pendingClosingChannels {
		userID := lightning.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[lightning.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := lightning.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := lightning.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedWithRemote += lockedByUser
			user.lockedLocally += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:           userID,
				isConnected:      isConnected,
				lockedLocally:    lockedByHub,
				lockedWithRemote: lockedByUser,
			}
		}
	}

	for _, c := range pendingForceClosingChannels {
		userID := lightning.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[lightning.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := lightning.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := lightning.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedWithRemote += lockedByUser
			user.lockedLocally += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:           userID,
				isConnected:      isConnected,
				lockedLocally:    lockedByHub,
				lockedWithRemote: lockedByUser,
			}
		}
	}

	for _, c := range pendingWaitingCloseChannels {
		userID := lightning.UserID(c.Channel.RemoteNodePub)
		_, isConnected := connectedPeersMap[userID]

		// If channel already exist from hub point of view,
		// than we need mark.
		// TODO(andrew.shvv) remove when channel would contain users
		hubChannel, ok := hubChannelView[lightning.ChannelID(c.Channel.ChannelPoint)]
		if ok {
			hubChannel.SetUserConnected(isConnected)
		}

		lockedByUser := lightning.BalanceUnit(c.Channel.RemoteBalance)
		lockedByHub := lightning.BalanceUnit(c.Channel.LocalBalance)

		if user, ok := lndPeersView[userID]; ok {
			user.lockedWithRemote += lockedByUser
			user.lockedLocally += lockedByHub
		} else {
			lndPeersView[userID] = &lndUser{
				userID:           userID,
				isConnected:      isConnected,
				lockedLocally:    lockedByHub,
				lockedWithRemote: lockedByUser,
			}
		}
	}

	hubPeersView := make(map[lightning.UserID]*lightning.User)
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
			hubUser.LockedByUser = lndUser.lockedWithRemote
			hubUser.LockedByHub = lndUser.lockedLocally

			if hubUser.IsConnected != lndUser.isConnected {
				hubUser.IsConnected = lndUser.isConnected
				broadcaster.Write(&lightning.UpdateUserConnected{
					User:        userID,
					IsConnected: hubUser.IsConnected,
				})
			}

			if err := hubUser.Save(); err != nil {
				return errors.Errorf("unable update user status: %v", err)
			}
		} else {
			newUser, err := lightning.NewUser(userID, lndUser.isConnected,
				registry.GetAlias(userID), lndUser.lockedLocally,
				lndUser.lockedWithRemote, &lightning.UserConfig{
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

			broadcaster.Write(&lightning.UpdateUserConnected{
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
			broadcaster.Write(&lightning.UpdateUserConnected{
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
func getInitiator(localBalance int64) lightning.ChannelInitiator {
	if localBalance == 0 {
		return lightning.RemoteInitiator
	} else {
		return lightning.LocalInitiator
	}
}

// listenOutgoingPayments listens for outgoing payment from out lightning
// network node.
//
// NOTE: Should run as goroutine.
func (client *Client) listenOutgoingPayments() {
	defer func() {
		log.Info("Stopped sync outgoing payments goroutine")
		client.wg.Done()
	}()

	log.Info("Started sync outgoing payments goroutine")

	for {
		select {
		case <-time.After(time.Second * 15):
		case <-client.quit:
			return
		}

		m := crypto.NewMetric(client.cfg.Asset, "SyncOutgoingPayments", client.cfg.MetricsBackend)

		payments, err := fetchOutgoingPayments(client.client)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			m.Finish()

			log.Errorf("(outgoing payments sync) unable to get lnd outgoing"+
				" payments: %v", err)
			continue
		}

		if err := syncOutgoing(client.lightningNodeUserID, payments, client.cfg.Storage,
			client.cfg.Storage, client.broadcaster); err != nil {
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
func syncOutgoing(localNodeID lightning.UserID, lndPayments []*lnrpc.Payment,
	syncStorage SyncStorage, paymentStorage lightning.PaymentStorage,
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

		sender := localNodeID
		receiver := lightning.UserID(payment.Path[0])
		amount := lightning.BalanceUnit(payment.Value)
		fee := lightning.BalanceUnit(payment.Fee)

		log.Infof("(outgoing payments sync) Process new outgoing payment"+
			" from(%v), to(%v), amount(%v), fee(%v), time(%v)", sender,
			receiver, amount, fee, payment.CreationDate)

		broadcaster.Write(&lightning.UpdatePayment{
			Type:     lightning.Outgoing,
			Status:   lightning.Successful,
			Sender:   sender,
			Receiver: receiver,
			Amount:   amount,
			Earned:   -fee,
		})

		if err := paymentStorage.StorePayment(&lightning.Payment{
			FromUser:  sender,
			ToUser:    receiver,
			FromAlias: registry.GetAlias(sender),
			ToAlias:   registry.GetAlias(receiver),
			Amount:    amount,
			Type:      lightning.Outgoing,
			Status:    lightning.Successful,
			Time:      payment.CreationDate,
		}); err != nil {
			log.Errorf("unable to save the payment: %v", err)
			continue
		}
	}

	return nil
}
