package lnd

import (
	"time"
	"github.com/lightningnetwork/lnd/lnrpc"
	"context"
	"github.com/bitlum/btcutil"
	"github.com/bitlum/hub/manager/router"
	"github.com/go-errors/errors"
	"github.com/davecgh/go-spew/spew"
	"fmt"
	"github.com/bitlum/hub/manager/metrics/crypto"
	"github.com/bitlum/hub/manager/metrics"
)

// ChannelsState is a map of chan points and current state of the channel.
type ChannelsState map[string]string

// listenLocalTopologyUpdates tracks the local channel topology state updates
// and sends notifications accordingly to the occurred transition.
func (r *Router) listenLocalTopologyUpdates() {
	r.wg.Add(1)
	go func() {
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
	}()
}

// syncTopologyUpdates is used as wrapper for fetching and syncing topology
// updates. Main purpose of creating distinct method was usage of defer, which
// is needed mainly for metric gathering usage.
func (r *Router) syncTopologyUpdates() {
	defer panicRecovering()

	m := crypto.NewMetric(r.cfg.Asset, "SyncTopologyUpdates",
		r.cfg.MetricsBackend)
	defer m.Finish()

	newChannelsState := make(ChannelsState)

	// Take all pending/closing/opened channels and form the current
	// state out of it. As far as those two operation are not atomic,
	// it might happen that some channel sleeps away.
	reqPending := &lnrpc.PendingChannelsRequest{}
	respPending, err := r.client.PendingChannels(context.Background(), reqPending)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("(topology updates) unable to fetch pending"+
			" channels: %v", err)
		return
	}

	reqOpen := &lnrpc.ListChannelsRequest{}
	respOpen, err := r.client.ListChannels(context.Background(), reqOpen)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("(topology updates) unable to fetch list channels"+
			": %v", err)
		return
	}

	for _, channel := range respPending.PendingOpenChannels {
		newChannelsState[channel.Channel.ChannelPoint] = "opening"
	}
	for _, channel := range respPending.PendingClosingChannels {
		newChannelsState[channel.Channel.ChannelPoint] = "closing"

	}
	for _, channel := range respPending.PendingForceClosingChannels {
		newChannelsState[channel.Channel.ChannelPoint] = "closing"
	}
	for _, channel := range respOpen.Channels {
		newChannelsState[channel.ChannelPoint] = "opened"
	}

	// Fetch prev pending/closing/opened channels from db which
	// corresponds to the old/previous channel state.
	oldChannelsState, err := r.cfg.DB.ChannelsState()
	log.Debugf("(topology updates) Fetching channel state from db...")
	if err != nil {
		m.AddError(metrics.HighSeverity)
		log.Errorf("(topology updates) unable to fetch old channel"+
			" state: %v", err)
		return
	}

	var updates []interface{}
	for chanPoint, oldState := range oldChannelsState {
		newState, ok := newChannelsState[chanPoint]
		if !ok {
			newState = "not_exist"
		}

		update, err := r.actOnChannelStateChange(oldState, newState)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) error during channel"+
				" states comparison: %v", err)
			continue
		}

		if update != nil {
			updates = append(updates, update)
		}
	}

	for chanPoint, newState := range newChannelsState {
		oldState, ok := oldChannelsState[chanPoint]
		if !ok {
			oldState = "not_exist"
		}

		update, err := r.actOnChannelStateChange(oldState, newState)
		if err != nil {
			m.AddError(metrics.MiddleSeverity)
			log.Errorf("(topology updates) error during channel"+
				" states comparison: %v",
				err)
			continue
		}

		if update != nil {
			updates = append(updates, update)
		}
	}

	// Avoid redundant db usage.
	if len(updates) == 0 {
		return
	}

	log.Infof("(topology updates) Broadcast %v topology updates",
		len(updates))

	// Try to save new state until it will be successful,
	// otherwise we might send the same notifications.
	for {
		select {
		case <-time.After(time.Second * 5):
		case <-r.quit:
			return
		}

		log.Debug("(topology updates) Putting channel state in db...")
		if err := r.cfg.DB.PutChannelsState(newChannelsState); err != nil {
			m.AddError(metrics.HighSeverity)
			log.Errorf("(topology updates) unable update channel"+
				" states: %v", err)
			continue
		}

		break
	}

	// Send updates only after database has changed. In the time of program
	// crushes updates might not be sent, but if we sent them before the
	// database flush, than they might be sent several times which is worse.
	for _, update := range updates {
		log.Debugf("(topology updates) Send topology update: %v",
			spew.Sdump(update))
		r.broadcaster.Write(update)
	}
}

// actOnChannelStateChange handles all possible cases of channel state
// change and send the update accordingly with the occurred transition.
func (r *Router) actOnChannelStateChange(oldState,
newState string) (interface{}, error) {
	transition := fmt.Sprintf("%v => %v", oldState, newState)

	var update interface{}

	switch transition {
	case "opening => opening":
		// Do nothing

	case "opening => opened":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelOpened{}

	case "opening => closing":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelClosing{}

	case "opening => not_exist":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelClosed{}

	case "opened => opening":
		return nil, errors.New("channel was opened but than became" +
			" opening")

	case "opened => opened":
		// Do nothing

	case "opened => closing":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelClosing{}

	case "opened => not_exist":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelClosed{}

	case "closing => opening":
		return nil, errors.New("channel was closing but than became" +
			" opening")

	case "closing => opened":
		return nil, errors.New("channel was closing but than became" +
			" opened")

	case "closing => closing":
		// Do nothing

	case "closing => not_exist":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelClosed{}

	case "not_exist => opening":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelOpening{}

	case "not_exist => opened":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelOpened{}

	case "not_exist => closing":
		// TODO(andrew.shvv) Populate
		update = &router.UpdateChannelClosing{}

	case "not_exist => not_exist":
		// Do nothing

	default:
		return nil, errors.Errorf("unable to handle case: %v", transition)
	}

	return update, nil
}

// listenForwardingUpdates listener of the forwarding lightning payments.
// Fetches the lnd forwarding log, and send new update to the broadcaster.
func (r *Router) listenForwardingUpdates() {
	r.wg.Add(1)
	go func() {
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
	}()
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
		log.Debug("(forwarding updates) Fetching last forwarding index...")
		lastIndex, err = r.cfg.DB.LastForwardingIndex()
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
		if err := r.cfg.DB.PutLastForwardingIndex(lastIndex); err != nil {
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
	resp, err := r.client.ListChannels(context.Background(), req)
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
			StartTime:    0,
			EndTime:      0,
			IndexOffset:  index,
			NumMaxEvents: limit,
		}

		resp, err := r.client.ForwardingHistory(context.Background(), req)
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
	defer panicRecovering()

	m := crypto.NewMetric(r.cfg.Asset, "ListenInvoiceUpdates",
		r.cfg.MetricsBackend)

	r.wg.Add(1)
	go func() {
		defer func() {
			panicRecovering()
			log.Info("Stopped incoming payments goroutine")
			r.wg.Done()
		}()

		log.Info("Started incoming payments goroutine")

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
	}()
}
