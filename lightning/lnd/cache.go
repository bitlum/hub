package lnd

import (
	"github.com/bitlum/hub/common"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/btcsuite/btcutil"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-errors/errors"
	"time"
)

// updateChannelStates fetches all lightning network node channels,
// and populate internal channel storage with additional information.
//
// NOTE: Should run as goroutine.
func (c *Client) updateChannelStates() {
	defer func() {
		log.Info("Stopped local topology updates goroutine")
		c.wg.Done()
	}()

	log.Info("Started local topology updates goroutine")

	for {
		select {
		case <-time.After(time.Second * 5):
		case <-c.quit:
			return
		}

		m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
			c.cfg.MetricsBackend)

		if err := c.syncChannels(); err != nil {
			log.Errorf("(topology updates) unable sync channel states: %v", err)
			m.AddError(metrics.HighSeverity)
			continue
		}
	}
}

// syncChannels is used to save some additional data about channel
// state, which otherwise will be lost, because lnd do not expose historical
// change of channel states. Also preserve caches of information to which are
// needed to make exposed lightning api better.
func (c *Client) syncChannels() error {
	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)

	openChannels, pendingOpenChannels, pendingClosingChannels,
	pendingForceClosingChannels, waitingCloseChannels, closedChannels,
	err := fetchChannels(c.rpc)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		return errors.Errorf("unable to fetch lightning"+
			" channels: %v", err)
	}

	// Proceed channels which are currently opening, which means that funding
	// transaction has been sent and we wait for it to be approved.
	for _, openingChannel := range pendingOpenChannels {
		chanID := lightning.ChannelID(openingChannel.Channel.ChannelPoint)
		nodeID := lightning.NodeID(openingChannel.Channel.RemoteNodePub)

		prevInfo, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil && err != ErrorChannelInfoNotFound {
			return errors.Errorf("unable to fetch channel(%v) info: %v",
				chanID, err)
		}

		prevState := lightning.ChannelStateName("not synced yet")
		if prevInfo != nil {
			prevState = prevInfo.State
		}

		switch prevState {
		case lightning.ChannelClosing:
			// Transition from closing => opening
			log.Errorf("impossible channel("+
				"%v) change state closing => opening", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case lightning.ChannelClosed:
			// Transition from closed => opening
			log.Errorf("impossible channel("+
				"%v) change state closed => opening", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case lightning.ChannelOpened:
			// Sometime weird thing is happening when channel is changing its
			// state from opened to opening,
			// until this bug change channel state.

			prevInfo.State = lightning.ChannelOpening
			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

		case lightning.ChannelOpening:
			// Transition from opening => opening
			// Nothing to do

		case "not synced yet":
			// Transition from non-existed => opening

			openTxID, _, err := splitChannelPoint(openingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable to get txid from channel "+
					"point(%v): %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			// Ignore explorer error, because sometime transaction might
			// already be in the response of lightning node but not in the
			// mem-pool.
			openTxFee, err := c.cfg.Explorer.FetchTxFee(openTxID)
			if err != nil {
				log.Warnf("unable to get open fee of channel("+
					"%v): %v", chanID, err)
				m.AddError(metrics.LowSeverity)
			}

			// Ignore explorer error, because sometime transaction might
			// already be in the response of lightning node but not in the
			// mem-pool.
			openTxTime, err := c.cfg.Explorer.FetchTxTime(openTxID)
			if err != nil {
				log.Warnf("unable to get open time of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.LowSeverity)
			}

			newInfo := &ChannelAdditionalInfo{
				ChannelID:      chanID,
				NodeID:         nodeID,
				ShortChannelID: 0, // will be known in open state

				OpeningTime:          openTxTime,
				OpeningInitiator:     getOpenInitiator(openingChannel.Channel.LocalBalance),
				OpeningCommitFees:    btcutil.Amount(openingChannel.CommitFee),
				OpeningFees:          openTxFee,
				OpeningRemoteBalance: btcutil.Amount(openingChannel.Channel.RemoteBalance),
				OpeningLocalBalance:  btcutil.Amount(openingChannel.Channel.LocalBalance),

				OpenTime:          0, // will be known in open state
				OpenCommitFees:    0, // will be known in open state
				OpenRemoteBalance: 0, // will be known in open state
				OpenLocalBalance:  0, // will be known in open state
				OpenStuckBalance:  0, // will be known in open state

				ClosingTime:          0, // will be known in closing state
				ClosingFees:          0, // will be known in closing state
				ClosingRemoteBalance: 0, // will be known in closing state
				ClosingLocalBalance:  0, // will be known in closing state
				SwipeFees:            0, // will be known in closing state

				CloseTime: 0, // will be known in close state

				State: lightning.ChannelOpening,
			}

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(newInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Good transition (not-existed => opening) channel("+
				"%v), info(%v)", chanID, spew.Sdump(newInfo))

		default:
			return errors.Errorf("unhandled state: %v", prevState)
		}
	}

	// Proceed channel which are currently open.
	for _, newChannel := range openChannels {
		chanID := lightning.ChannelID(newChannel.ChannelPoint)
		nodeID := lightning.NodeID(newChannel.RemotePubkey)

		prevInfo, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil && err != ErrorChannelInfoNotFound {
			log.Errorf("unable to fetch channel(%v) info: %v",
				chanID, err)
			m.AddError(metrics.HighSeverity)
			continue
		}

		prevState := lightning.ChannelStateName("not synced yet")
		if prevInfo != nil {
			prevState = prevInfo.State
		}

		switch prevState {
		case lightning.ChannelOpening:
			// Transition from opening => opened
			// haven't skipped state transitions

			// Info already exist which means, that we already populated data
			// in switch case above. Here we need add some additional which
			// became available.
			prevInfo.ShortChannelID = newChannel.ChanId
			prevInfo.OpenTime = time.Now().Unix()
			prevInfo.OpenCommitFees = btcutil.Amount(newChannel.CommitFee)
			prevInfo.OpenRemoteBalance = btcutil.Amount(newChannel.RemoteBalance)
			prevInfo.OpenLocalBalance = btcutil.Amount(newChannel.LocalBalance)
			prevInfo.OpenStuckBalance = getStuckBalance(newChannel.PendingHtlcs)
			prevInfo.State = lightning.ChannelOpened

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Good transition (opening => opened) channel(%v), "+
				"info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelOpened:
			// Transition from opened => opened

			// Info already exist which means, that we already populated data
			// in switch case above. Here we update channel state because
			// remote, local balances, as well as commit fee might change
			// with time.
			prevInfo.OpenCommitFees = btcutil.Amount(newChannel.CommitFee)
			prevInfo.OpenRemoteBalance = btcutil.Amount(newChannel.RemoteBalance)
			prevInfo.OpenLocalBalance = btcutil.Amount(newChannel.LocalBalance)
			prevInfo.OpenStuckBalance = getStuckBalance(newChannel.PendingHtlcs)

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

		case lightning.ChannelClosing:
			// Transition from closing => opened
			log.Errorf("impossible channel("+
				"%v) change state closing => opened", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case lightning.ChannelClosed:
			// Transition from closed => opened
			log.Errorf("impossible channel("+
				"%v) change state closed => opened", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case "not synced yet":
			// Transition from non-existed => opened
			// skipped opening

			// For some reason we have skipped opening sync switch case,
			// might be for a reason of our system being down.
			// And now we see already opened channel, which means that we have
			// skipped some data, and will be unable to restore it.

			openTxID, _, err := splitChannelPoint(newChannel.ChannelPoint)
			if err != nil {
				log.Errorf("unable to get txid from channel "+
					"point(%v): %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxFee, err := c.cfg.Explorer.FetchTxFee(openTxID)
			if err != nil {
				log.Errorf("unable to get open fee of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxTime, err := c.cfg.Explorer.FetchTxTime(openTxID)
			if err != nil {
				log.Errorf("unable to get open time of channel("+
					"%v): %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			newInfo := &ChannelAdditionalInfo{
				ChannelID:      chanID,
				NodeID:         nodeID,
				ShortChannelID: newChannel.ChanId,

				OpeningTime:          openTxTime,
				OpeningInitiator:     "", // unknown, lost data
				OpeningCommitFees:    0,  // unknown, lost data
				OpeningFees:          openTxFee,
				OpeningRemoteBalance: 0, // unknown, lost data
				OpeningLocalBalance:  0, // unknown, lost data

				OpenTime:          time.Now().Unix(),
				OpenCommitFees:    btcutil.Amount(newChannel.CommitFee),
				OpenLocalBalance:  btcutil.Amount(newChannel.LocalBalance),
				OpenRemoteBalance: btcutil.Amount(newChannel.RemoteBalance),
				OpenStuckBalance:  getStuckBalance(newChannel.PendingHtlcs),

				ClosingTime:          0, // will be known in closing state
				ClosingFees:          0, // will be known in closing state
				ClosingRemoteBalance: 0, // will be known in closing state
				ClosingLocalBalance:  0, // will be known in closing state
				SwipeFees:            0, // will be known in closing state

				CloseTime: 0, // will be known in close state

				State: lightning.ChannelOpened,
			}

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(newInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (not-existed => opened) channel("+
				"%v), info(%v)", chanID, spew.Sdump(newInfo))

		default:
			return errors.Errorf("unhandled state: %v", prevState)
		}
	}

	// Proceed channel which currently waiting for close, which means that they
	// are not active, but close transaction haven't been sent in blockchain
	// network yet.
	for _, waitingClosingChannel := range waitingCloseChannels {
		chanID := lightning.ChannelID(waitingClosingChannel.Channel.ChannelPoint)
		nodeID := lightning.NodeID(waitingClosingChannel.Channel.RemoteNodePub)

		prevInfo, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil && err != ErrorChannelInfoNotFound {
			log.Errorf("unable to fetch channel(%v) state: %v",
				chanID, err)
			m.AddError(metrics.HighSeverity)
			continue
		}

		prevState := lightning.ChannelStateName("not synced yet")
		if prevInfo != nil {
			prevState = prevInfo.State
		}

		switch prevState {
		case lightning.ChannelOpening:
			// Transition from opening => wait for closing
			// skipped opened

			// For some reason we have skipped opened sync switch case,
			// might be for a reason of our system being down.
			// And now we see already channel waiting for closing,
			// which means that we have skipped some data, and will
			// be unable to restore it.

			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				waitingClosingChannel.Channel.ChannelPoint)
			if err != nil {
				return errors.Errorf("unable get short channel id by channel"+
					" point: %v", err)
			}

			// Info already exist which, means that we have populated data
			// from of opening state.
			prevInfo.ShortChannelID = shortChannelID
			prevInfo.OpenTime = time.Now().Unix()
			prevInfo.OpenCommitFees = 0 // unknown, lost data
			prevInfo.OpenRemoteBalance = btcutil.Amount(waitingClosingChannel.Channel.RemoteBalance)
			prevInfo.OpenLocalBalance = btcutil.Amount(waitingClosingChannel.Channel.LocalBalance)
			prevInfo.OpenStuckBalance = 0 // unknown, lost data
			prevInfo.State = lightning.ChannelOpened

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (opening => wait for closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelOpened:
			// Transition from opened => wait for closing
			// haven't skipped state transitions

		case lightning.ChannelClosing:
			// Transition from closing => waiting to closing
			log.Errorf("impossible channel("+
				"%v) change state (closing) => (wait for closing)", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case lightning.ChannelClosed:
			// Transition from closed => opened
			log.Errorf("impossible channel("+
				"%v) change state (closed) => (wait for closing)", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case "not synced yet":
			// Transition from non-existed => wait for closing
			// skipped opening
			// skipped opened

			openTxID, _, err := splitChannelPoint(waitingClosingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable to get txid from channel "+
					"point(%v): %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxFee, err := c.cfg.Explorer.FetchTxFee(openTxID)
			if err != nil {
				log.Errorf("unable to get open fee of channel (%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxTime, err := c.cfg.Explorer.FetchTxTime(openTxID)
			if err != nil {
				log.Errorf("unable to get open time of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				waitingClosingChannel.Channel.ChannelPoint)
			if err != nil {
				return errors.Errorf("unable get short channel id by channel"+
					" point: %v", err)
			}

			newInfo := &ChannelAdditionalInfo{
				ChannelID:      chanID,
				NodeID:         nodeID,
				ShortChannelID: shortChannelID,

				OpeningTime:          openTxTime,
				OpeningInitiator:     "", // unknown, lost data
				OpeningCommitFees:    0,  // unknown, lost data
				OpeningFees:          openTxFee,
				OpeningRemoteBalance: 0, // unknown, lost data
				OpeningLocalBalance:  0, // unknown, lost data

				OpenTime:          time.Now().Unix(),
				OpenCommitFees:    0, // unknown, lost data
				OpenLocalBalance:  btcutil.Amount(waitingClosingChannel.Channel.LocalBalance),
				OpenRemoteBalance: btcutil.Amount(waitingClosingChannel.Channel.RemoteBalance),
				OpenStuckBalance:  0, // unknown, lost data

				ClosingTime:          0, // will be known in closing state
				ClosingFees:          0, // will be known in closing state
				ClosingRemoteBalance: 0, // will be known in closing state
				ClosingLocalBalance:  0, // will be known in closing state
				SwipeFees:            0, // will be known in closing state

				CloseTime: 0, // will be known in close state

				State: lightning.ChannelOpened,
			}

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(newInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (not-existed => wait for closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(newInfo))

		default:
			return errors.Errorf("unhandled state: %v", prevState)
		}
	}

	// Proceed channel which currently waiting to be closed,
	// this channels are the ones which were closed cooperatively, and lnd
	// waits for close tx to be confirmed.
	for _, closingChannel := range pendingClosingChannels {
		chanID := lightning.ChannelID(closingChannel.Channel.ChannelPoint)
		nodeID := lightning.NodeID(closingChannel.Channel.RemoteNodePub)

		prevInfo, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil && err != ErrorChannelInfoNotFound {
			log.Errorf("unable to fetch channel(%v) state: %v",
				chanID, err)
			m.AddError(metrics.HighSeverity)
			continue
		}

		prevState := lightning.ChannelStateName("not synced yet")
		if prevInfo != nil {
			prevState = prevInfo.State
		}

		switch prevState {
		case lightning.ChannelOpening:
			// Transition from opening => closing
			// skipped opened transition

			// For some reason we have skipped opened sync switch case,
			// might be for a reason of our system being down.
			// And now we see already closing state, which means that we have
			// skipped some data, and will be unable to restore it.

			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				closingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable get short channel id by channel"+
					" point: %v", err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			prevInfo.ShortChannelID = shortChannelID
			prevInfo.OpenTime = time.Now().Unix()
			prevInfo.OpenCommitFees = 0 // unknown, lost data
			prevInfo.OpenRemoteBalance = btcutil.Amount(closingChannel.Channel.RemoteBalance)
			prevInfo.OpenLocalBalance = btcutil.Amount(closingChannel.Channel.LocalBalance)
			prevInfo.OpenStuckBalance = 0 // unknown, lost data

			prevInfo.ClosingTime = closeTxTime
			prevInfo.ClosingFees = closeTxFee
			prevInfo.ClosingRemoteBalance = btcutil.Amount(closingChannel.Channel.RemoteBalance)
			prevInfo.ClosingLocalBalance = btcutil.Amount(closingChannel.Channel.LocalBalance)
			prevInfo.SwipeFees = 0 // cooperative close
			prevInfo.State = lightning.ChannelClosing

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (opening => closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelOpened:
			// Transition from opened => closing
			// skipped opened

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			prevInfo.ClosingTime = closeTxTime
			prevInfo.ClosingFees = closeTxFee
			prevInfo.ClosingRemoteBalance = btcutil.Amount(closingChannel.Channel.RemoteBalance)
			prevInfo.ClosingLocalBalance = btcutil.Amount(closingChannel.Channel.LocalBalance)
			prevInfo.SwipeFees = 0 // cooperate close
			prevInfo.State = lightning.ChannelClosing

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Good transition (opened => closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelClosing:
			// Transition from closing => closing
			// Nothing has changed

		case lightning.ChannelClosed:
			// Transition from closed => closing
			log.Errorf("impossible channel("+
				"%v) change state closed => closing", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case "not synced yet":
			// Transition from not existed => closing
			// skipped opening transition
			// skipped opened transition

			txid, _, err := splitChannelPoint(closingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable to get txid from channel "+
					"point(%v): %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxFee, err := c.cfg.Explorer.FetchTxFee(txid)
			if err != nil {
				log.Errorf("unable to get open fee of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxTime, err := c.cfg.Explorer.FetchTxTime(txid)
			if err != nil {
				return errors.Errorf("unable to get open time of channel(%v): %v",
					chanID, err)
			}

			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				closingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable get short channel id by channel"+
					" point: %v", err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			// Save additional info about channel initiator and open channel
			// fees which will be needed later.
			newInfo := &ChannelAdditionalInfo{
				ChannelID:      chanID,
				NodeID:         nodeID,
				ShortChannelID: shortChannelID,

				OpeningTime:          openTxTime,
				OpeningInitiator:     "", // unknown, lost data
				OpeningFees:          openTxFee,
				OpeningRemoteBalance: 0, // unknown, lost data
				OpeningLocalBalance:  0, // unknown, lost data

				OpenTime:          time.Now().Unix(),
				OpenCommitFees:    0, // unknown, lost data
				OpenRemoteBalance: btcutil.Amount(closingChannel.Channel.RemoteBalance),
				OpenLocalBalance:  btcutil.Amount(closingChannel.Channel.LocalBalance),
				OpenStuckBalance:  0, // unknown, lost data

				ClosingTime:          closeTxTime,
				ClosingFees:          closeTxFee,
				ClosingRemoteBalance: btcutil.Amount(closingChannel.Channel.RemoteBalance),
				ClosingLocalBalance:  btcutil.Amount(closingChannel.Channel.LocalBalance),
				SwipeFees:            0, // cooperative close

				State: lightning.ChannelClosing,
			}

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(newInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (not-existed => closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(newInfo))

		default:
			return errors.Errorf("unhandled state: %v", prevState)
		}
	}

	// Proceed channel which currently waiting to be closed,
	// this channel are the one which were forced closed it might be that we
	// have pending htlc which has to be swiped, for that reason they might
	// stuck in this state for a long time.
	for _, closingChannel := range pendingForceClosingChannels {
		chanID := lightning.ChannelID(closingChannel.Channel.ChannelPoint)
		nodeID := lightning.NodeID(closingChannel.Channel.RemoteNodePub)

		prevInfo, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil && err != ErrorChannelInfoNotFound {
			log.Errorf("unable to fetch channel(%v) state: %v",
				chanID, err)
			m.AddError(metrics.HighSeverity)
			continue
		}

		prevState := lightning.ChannelStateName("not synced yet")
		if prevInfo != nil {
			prevState = prevInfo.State
		}

		switch prevState {
		case lightning.ChannelOpening:
			// Transition from opening => closing

			// In this case for some reason we missed opened state, and got
			// directly to the closing state, which means that we skipped sync
			// in open case.
			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				closingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable get short channel id by channel"+
					" point: %v", err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			var swipeFees btcutil.Amount
			for _, htlc := range closingChannel.PendingHtlcs {
				if htlc.Incoming {
					continue
				}

				swipeTxID, _, err := splitChannelPoint(htlc.Outpoint)
				if err != nil {
					log.Errorf("unable split "+
						"channel point: %v", err)
					m.AddError(metrics.HighSeverity)
					continue
				}

				swipeTxFee, err := c.cfg.Explorer.FetchTxFee(swipeTxID)
				if err != nil {
					return errors.Errorf("unable to get swipe tx fee(%v): %v",
						swipeTxID, err)
				}

				swipeFees += swipeTxFee
			}

			prevInfo.ShortChannelID = shortChannelID
			prevInfo.OpenTime = time.Now().Unix()
			prevInfo.OpenCommitFees = 0 // unknown to us, synced to late
			prevInfo.OpenRemoteBalance = btcutil.Amount(closingChannel.Channel.RemoteBalance)
			prevInfo.OpenLocalBalance = btcutil.Amount(closingChannel.Channel.LocalBalance)
			prevInfo.OpenStuckBalance = 0 // unknown to us, synced to late

			prevInfo.ClosingTime = closeTxTime
			prevInfo.ClosingFees = closeTxFee
			prevInfo.ClosingRemoteBalance = btcutil.Amount(closingChannel.Channel.RemoteBalance)
			prevInfo.ClosingLocalBalance = btcutil.Amount(closingChannel.Channel.LocalBalance)
			prevInfo.SwipeFees = swipeFees
			prevInfo.State = lightning.ChannelClosing

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (opening => closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelOpened:
			// Transition from opened => closing

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			var swipeFees btcutil.Amount
			for _, htlc := range closingChannel.PendingHtlcs {
				if htlc.Incoming {
					continue
				}

				swipeTxID, _, err := splitChannelPoint(htlc.Outpoint)
				if err != nil {
					log.Errorf("unable split "+
						"channel point: %v", err)
					m.AddError(metrics.HighSeverity)
					continue
				}

				swipeTxFee, err := c.cfg.Explorer.FetchTxFee(swipeTxID)
				if err != nil {
					log.Errorf("unable to get swipe tx fee(%v): %v",
						swipeTxID, err)
					m.AddError(metrics.HighSeverity)
					continue
				}

				swipeFees += swipeTxFee
			}

			prevInfo.ClosingTime = closeTxTime
			prevInfo.ClosingFees = closeTxFee
			prevInfo.ClosingRemoteBalance = btcutil.Amount(closingChannel.Channel.RemoteBalance)
			prevInfo.ClosingLocalBalance = btcutil.Amount(closingChannel.Channel.LocalBalance)
			prevInfo.SwipeFees = swipeFees
			prevInfo.State = lightning.ChannelClosing

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Good transition (opened => closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelClosing:
			// Nothing has changed

		case lightning.ChannelClosed:
			// Previous channel state was closed, and now it is
			// closing, we couldn't  re-close, closed channel.
			log.Errorf("impossible channel("+
				"%v) change state closed => closing", chanID)
			m.AddError(metrics.HighSeverity)
			continue

		case "not synced yet":
			openTxID, _, err := splitChannelPoint(closingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable to get txid from channel "+
					"point(%v): %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxFee, err := c.cfg.Explorer.FetchTxFee(openTxID)
			if err != nil {
				log.Errorf("unable to get open fee of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxTime, err := c.cfg.Explorer.FetchTxTime(openTxID)
			if err != nil {
				log.Errorf("unable to get open time of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				closingChannel.Channel.ChannelPoint)
			if err != nil {
				log.Errorf("unable get short channel id by channel"+
					" point: %v", err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closingChannel.ClosingTxid)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closingChannel.ClosingTxid, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			var swipeFees btcutil.Amount
			for _, htlc := range closingChannel.PendingHtlcs {
				if htlc.Incoming {
					continue
				}

				swipeTxID, _, err := splitChannelPoint(htlc.Outpoint)
				if err != nil {
					log.Errorf("unable split "+
						"channel point: %v", err)
					m.AddError(metrics.HighSeverity)
					break
				}

				swipeTxFee, err := c.cfg.Explorer.FetchTxFee(swipeTxID)
				if err != nil {
					log.Errorf("unable to get swipe tx fee(%v): %v",
						swipeTxID, err)
					m.AddError(metrics.HighSeverity)
					break
				}

				swipeFees += swipeTxFee
			}

			// Save additional info about channel initiator and open channel
			// fees which will be needed later.
			newInfo := &ChannelAdditionalInfo{
				ChannelID:      chanID,
				NodeID:         nodeID,
				ShortChannelID: shortChannelID,

				OpeningTime:          openTxTime,
				OpeningInitiator:     "", // unknown to us, synced to late
				OpeningFees:          openTxFee,
				OpeningRemoteBalance: 0, // unknown to us, synced to late
				OpeningLocalBalance:  0, // unknown to us, synced to late

				OpenTime:          time.Now().Unix(),
				OpenCommitFees:    0, // unknown to us, synced to late
				OpenRemoteBalance: btcutil.Amount(closingChannel.Channel.RemoteBalance),
				OpenLocalBalance:  btcutil.Amount(closingChannel.Channel.LocalBalance),
				OpenStuckBalance:  0, // unknown to us, synced to late

				ClosingTime:          closeTxTime,
				ClosingFees:          closeTxFee,
				ClosingRemoteBalance: btcutil.Amount(closingChannel.Channel.RemoteBalance),
				ClosingLocalBalance:  btcutil.Amount(closingChannel.Channel.LocalBalance),
				SwipeFees:            swipeFees,

				State: lightning.ChannelClosing,
			}

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(newInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (not-existed => closing) channel("+
				"%v), info(%v)", chanID, spew.Sdump(newInfo))

		default:
			return errors.Errorf("unhandled state: %v", prevState)
		}
	}

	// Proceed channel which has been closed, and funds has been returned back
	// to wallet.
	for _, closedChannel := range closedChannels {
		// Skip old corrupted channels.
		emptyTxHash := "0000000000000000000000000000000000000000000000000000000000000000"
		if closedChannel.ClosingTxHash == emptyTxHash {
			continue
		}

		chanID := lightning.ChannelID(closedChannel.ChannelPoint)
		nodeID := lightning.NodeID(closedChannel.RemotePubkey)

		prevInfo, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil && err != ErrorChannelInfoNotFound {
			log.Errorf("unable to fetch channel(%v) state: %v",
				chanID, err)
			m.AddError(metrics.HighSeverity)
			continue
		}

		prevState := lightning.ChannelStateName("not synced yet")
		if prevInfo != nil {
			prevState = prevInfo.State
		}

		switch prevState {
		case lightning.ChannelOpening:
			// Transition from opening => closed
			// skipped opened
			// skipped closing

			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				closedChannel.ChannelPoint)
			if err != nil {
				log.Errorf("unable get short channel id by channel"+
					" point: %v", err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closedChannel.ClosingTxHash)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closedChannel.ClosingTxHash, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closedChannel.ClosingTxHash)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closedChannel.ClosingTxHash, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			// Save additional info about channel initiator and open channel
			// fees which will be needed later.

			prevInfo.ShortChannelID = shortChannelID
			prevInfo.OpenTime = time.Now().Unix()
			prevInfo.OpenCommitFees = 0    // unknown to us
			prevInfo.OpenRemoteBalance = 0 // unknown to us
			prevInfo.OpenLocalBalance = 0  // unknown to us
			prevInfo.OpenStuckBalance = 0  // unknown to us

			prevInfo.ClosingTime = closeTxTime
			prevInfo.ClosingFees = closeTxFee
			prevInfo.ClosingRemoteBalance = 0 // unknown to us
			prevInfo.ClosingLocalBalance = 0  // unknown to us
			prevInfo.SwipeFees = 0            // unknown to us

			prevInfo.CloseTime = time.Now().Unix()
			prevInfo.State = lightning.ChannelClosed

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (opening => closed) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelOpened:
			// Transition from opened => closed
			// skipped closing

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closedChannel.ClosingTxHash)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closedChannel.ClosingTxHash, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closedChannel.ClosingTxHash)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closedChannel.ClosingTxHash, err)
			}

			// Save additional info about channel initiator and open channel
			// fees which will be needed later.
			prevInfo.ClosingTime = closeTxTime
			prevInfo.ClosingFees = closeTxFee
			prevInfo.ClosingRemoteBalance = 0 // unknown to us
			prevInfo.ClosingLocalBalance = 0  // unknown to us
			prevInfo.SwipeFees = 0            // unknown to us

			prevInfo.CloseTime = time.Now().Unix()
			prevInfo.State = lightning.ChannelClosed

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (opened => closed) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelClosing:
			// Transition from closing => closed
			// No transitions skipped

			prevInfo.CloseTime = time.Now().Unix()
			prevInfo.State = lightning.ChannelClosed

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(prevInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Good transition (closing => closed) channel("+
				"%v), info(%v)", chanID, spew.Sdump(prevInfo))

		case lightning.ChannelClosed:
			// Transition from not closed => closed

		case "not synced yet":
			// Transition from not existed => closed
			// skipped opening
			// skipped opened
			// skipped closing

			openTxID, _, err := splitChannelPoint(closedChannel.ChannelPoint)
			if err != nil {
				log.Errorf("unable to get txid from channel "+
					"point(%v): %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxFee, err := c.cfg.Explorer.FetchTxFee(openTxID)
			if err != nil {
				log.Errorf("unable to get open time of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			openTxTime, err := c.cfg.Explorer.FetchTxTime(openTxID)
			if err != nil {
				log.Errorf("unable to get open fee of channel(%v): %v",
					chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			shortChannelID, err := getShortChannelIDByChannelPoint(c.cfg.Explorer,
				closedChannel.ChannelPoint)
			if err != nil {
				log.Errorf("unable get short channel id by channel"+
					" point: %v", err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxTime, err := c.cfg.Explorer.FetchTxTime(closedChannel.ClosingTxHash)
			if err != nil {
				log.Errorf("unable to get close time of channel("+
					"%v), tx(%v): %v", chanID, closedChannel.ClosingTxHash, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			closeTxFee, err := c.cfg.Explorer.FetchTxFee(closedChannel.ClosingTxHash)
			if err != nil {
				log.Errorf("unable to get close fee of channel("+
					"%v), tx(%v): %v", chanID, closedChannel.ClosingTxHash, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			newInfo := &ChannelAdditionalInfo{
				ChannelID:      chanID,
				NodeID:         nodeID,
				ShortChannelID: shortChannelID,

				OpeningTime:          openTxTime,
				OpeningInitiator:     "", // unknown, lost data
				OpeningFees:          openTxFee,
				OpeningRemoteBalance: 0, // unknown, lost data
				OpeningLocalBalance:  0, // unknown, lost data

				OpenTime:          time.Now().Unix(),
				OpenCommitFees:    0, // unknown, lost data
				OpenRemoteBalance: 0, // unknown, lost data
				OpenLocalBalance:  0, // unknown, lost data
				OpenStuckBalance:  0, // unknown, lost data

				ClosingTime:          closeTxTime,
				ClosingFees:          closeTxFee,
				ClosingRemoteBalance: 0, // unknown, lost data
				ClosingLocalBalance:  0, // unknown, lost data
				SwipeFees:            0, // unknown, lost data

				CloseTime: time.Now().Unix(),
				State:     lightning.ChannelClosed,
			}

			if err := c.cfg.Storage.UpdateChannelAdditionalInfo(newInfo); err != nil {
				log.Errorf("unable to save channel("+
					"%v) additional info: %v", chanID, err)
				m.AddError(metrics.HighSeverity)
				continue
			}

			log.Debugf("Bad transition (not-existed => closed) channel("+
				"%v), info(%v)", chanID, spew.Sdump(newInfo))

		default:
			return errors.Errorf("unhandled state: %v", prevState)
		}
	}

	return nil
}
