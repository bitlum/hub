package lnd

import (
	"github.com/bitlum/hub/common"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/metrics"
	"github.com/bitlum/hub/metrics/crypto"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"strconv"
	"strings"
)

// Runtime check to ensure that Client implements lightning.
// TopologyClient interface.
var _ lightning.TopologyClient = (*Client)(nil)

// Network returns the information about the current local topology of
// our lightning node.
//
// NOTE: Part of the lightning.TopologyClient interface.
func (c *Client) Channels() ([]*lightning.Channel, error) {
	select {
	case <-c.startedTrigger:
	}

	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	openChannels, pendingOpenChannels, pendingClosingChannels,
	pendingForceClosingChannels, waitingCloseChannels, closedChannels,
	err := fetchChannels(c.rpc)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable to sync channels: %v", err)
		log.Error(err)
		return nil, err
	}

	// Update additional info about channels
	if err := c.syncChannels(); err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable to sync channels: %v", err)
		log.Error(err)
	}

	var channels []*lightning.Channel

	for _, pendingOpenInfo := range pendingOpenChannels {
		chanID := lightning.ChannelID(pendingOpenInfo.Channel.ChannelPoint)
		nodeID := lightning.NodeID(pendingOpenInfo.Channel.RemoteNodePub)

		infoFromDB, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			err := errors.Errorf("unable to get info of channel(%v): %v",
				chanID, err)
			log.Error(err)

			// For some new channels, we might not see transaction yet,
			// or it might not be broadcasted in blockchain by another party,
			// just skip such channels.
			continue
		}

		stateMap := make(map[lightning.ChannelStateName]interface{})

		stateMap[lightning.ChannelOpening] = &lightning.ChannelStateOpening{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpeningTime,
			CommitFee:     btcutil.Amount(pendingOpenInfo.CommitFee),
			OpenFee:       infoFromDB.OpeningFees,
			RemoteBalance: btcutil.Amount(pendingOpenInfo.Channel.RemoteBalance),
			LocalBalance:  btcutil.Amount(pendingOpenInfo.Channel.LocalBalance),
			Initiator:     infoFromDB.OpeningInitiator,
		}

		channels = append(channels, &lightning.Channel{
			NodeID:    nodeID,
			ChannelID: chanID,
			State:     lightning.ChannelOpening,
			States:    stateMap,
		})
	}

	for _, openedChannel := range openChannels {
		chanID := lightning.ChannelID(openedChannel.ChannelPoint)
		nodeID := lightning.NodeID(openedChannel.RemotePubkey)

		infoFromDB, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			err := errors.Errorf("unable to get info of channel(%v): %v",
				chanID, err)
			log.Error(err)
			return nil, err
		}

		stateMap := make(map[lightning.ChannelStateName]interface{})

		stateMap[lightning.ChannelOpening] = &lightning.ChannelStateOpening{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpeningTime,
			CommitFee:     btcutil.Amount(openedChannel.CommitFee),
			OpenFee:       infoFromDB.OpeningFees,
			RemoteBalance: btcutil.Amount(openedChannel.RemoteBalance),
			LocalBalance:  btcutil.Amount(openedChannel.LocalBalance),
			Initiator:     infoFromDB.OpeningInitiator,
		}

		stateMap[lightning.ChannelOpened] = &lightning.ChannelStateOpened{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpenTime,
			CommitFee:     btcutil.Amount(openedChannel.CommitFee),
			RemoteBalance: btcutil.Amount(openedChannel.RemoteBalance),
			LocalBalance:  btcutil.Amount(openedChannel.LocalBalance),
			IsActive:      openedChannel.Active,
			StuckBalance:  getStuckBalance(openedChannel.PendingHtlcs),
		}

		channels = append(channels, &lightning.Channel{
			NodeID:    nodeID,
			ChannelID: chanID,
			State:     lightning.ChannelOpened,
			States:    stateMap,
		})
	}

	for _, waitingCloseInfo := range waitingCloseChannels {
		chanID := lightning.ChannelID(waitingCloseInfo.Channel.ChannelPoint)
		nodeID := lightning.NodeID(waitingCloseInfo.Channel.RemoteNodePub)

		infoFromDB, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			err := errors.Errorf("unable to get info of channel(%v): %v",
				chanID, err)
			log.Error(err)
			return nil, err
		}

		stateMap := make(map[lightning.ChannelStateName]interface{})

		stateMap[lightning.ChannelOpening] = &lightning.ChannelStateOpening{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpeningTime,
			OpenFee:       infoFromDB.OpeningFees,
			CommitFee:     infoFromDB.OpeningCommitFees,
			RemoteBalance: infoFromDB.OpeningRemoteBalance,
			LocalBalance:  infoFromDB.OpeningLocalBalance,
			Initiator:     infoFromDB.OpeningInitiator,
		}

		stateMap[lightning.ChannelOpened] = &lightning.ChannelStateOpened{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpenTime,
			CommitFee:     infoFromDB.OpenCommitFees,
			RemoteBalance: btcutil.Amount(waitingCloseInfo.Channel.RemoteBalance),
			LocalBalance:  btcutil.Amount(waitingCloseInfo.Channel.LocalBalance),
			IsActive:      false,
			StuckBalance:  infoFromDB.OpenStuckBalance,
		}

		channels = append(channels, &lightning.Channel{
			NodeID:    nodeID,
			ChannelID: chanID,
			State:     lightning.ChannelOpened,
			States:    stateMap,
		})
	}

	for _, pendingClosingInfo := range pendingClosingChannels {
		chanID := lightning.ChannelID(pendingClosingInfo.Channel.ChannelPoint)
		nodeID := lightning.NodeID(pendingClosingInfo.Channel.RemoteNodePub)

		infoFromDB, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			err := errors.Errorf("unable to get info of channel(%v): %v",
				chanID, err)
			log.Error(err)
			return nil, err
		}

		stateMap := make(map[lightning.ChannelStateName]interface{})

		stateMap[lightning.ChannelOpening] = &lightning.ChannelStateOpening{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpeningTime,
			OpenFee:       infoFromDB.OpeningFees,
			CommitFee:     infoFromDB.OpeningCommitFees,
			RemoteBalance: infoFromDB.OpeningRemoteBalance,
			LocalBalance:  infoFromDB.OpeningLocalBalance,
			Initiator:     infoFromDB.OpeningInitiator,
		}

		stateMap[lightning.ChannelOpened] = &lightning.ChannelStateOpened{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpenTime,
			CommitFee:     infoFromDB.OpenCommitFees,
			RemoteBalance: infoFromDB.OpenRemoteBalance,
			LocalBalance:  infoFromDB.OpenLocalBalance,
			IsActive:      false,
			StuckBalance:  infoFromDB.OpenStuckBalance,
		}

		stateMap[lightning.ChannelClosing] = &lightning.ChannelStateClosing{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.ClosingTime,
			CloseFee:      infoFromDB.ClosingFees,
			RemoteBalance: btcutil.Amount(pendingClosingInfo.Channel.RemoteBalance),
			LocalBalance:  btcutil.Amount(pendingClosingInfo.Channel.LocalBalance),
			LockedBalance: 0,
		}

		channels = append(channels, &lightning.Channel{
			NodeID:    nodeID,
			ChannelID: chanID,
			State:     lightning.ChannelClosing,
			States:    stateMap,
		})
	}

	for _, pendingClosingInfo := range pendingForceClosingChannels {
		chanID := lightning.ChannelID(pendingClosingInfo.Channel.ChannelPoint)
		nodeID := lightning.NodeID(pendingClosingInfo.Channel.RemoteNodePub)

		infoFromDB, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			err := errors.Errorf("unable to get info of channel(%v): %v",
				chanID, err)
			log.Error(err)
			return nil, err
		}

		stateMap := make(map[lightning.ChannelStateName]interface{})

		stateMap[lightning.ChannelOpening] = &lightning.ChannelStateOpening{
			CreationTime:  infoFromDB.OpeningTime,
			OpenFee:       infoFromDB.OpeningFees,
			CommitFee:     infoFromDB.OpeningCommitFees,
			RemoteBalance: infoFromDB.OpeningRemoteBalance,
			LocalBalance:  infoFromDB.OpeningLocalBalance,
			Initiator:     infoFromDB.OpeningInitiator,
		}

		stateMap[lightning.ChannelOpened] = &lightning.ChannelStateOpened{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpenTime,
			CommitFee:     infoFromDB.OpenCommitFees,
			RemoteBalance: infoFromDB.OpenRemoteBalance,
			LocalBalance:  infoFromDB.OpenLocalBalance,
			IsActive:      false,
			StuckBalance:  infoFromDB.OpenStuckBalance,
		}

		stateMap[lightning.ChannelClosing] = &lightning.ChannelStateClosing{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.ClosingTime,
			CloseFee:      infoFromDB.ClosingFees,
			RemoteBalance: btcutil.Amount(pendingClosingInfo.Channel.RemoteBalance),
			LocalBalance:  btcutil.Amount(pendingClosingInfo.Channel.LocalBalance),
			LockedBalance: btcutil.Amount(pendingClosingInfo.LimboBalance),
		}

		channels = append(channels, &lightning.Channel{
			NodeID:    nodeID,
			ChannelID: chanID,
			State:     lightning.ChannelClosing,
			States:    stateMap,
		})
	}

	for _, closeSummary := range closedChannels {
		// Skip old corrupted channels.
		emptyTxHash := "0000000000000000000000000000000000000000000000000000000000000000"
		if closeSummary.ClosingTxHash == emptyTxHash {
			continue
		}

		chanID := lightning.ChannelID(closeSummary.ChannelPoint)
		nodeID := lightning.NodeID(closeSummary.RemotePubkey)

		infoFromDB, err := c.cfg.Storage.GetChannelAdditionalInfoByID(chanID)
		if err != nil {
			m.AddError(metrics.HighSeverity)
			err := errors.Errorf("unable to get info of channel(%v): %v",
				chanID, err)
			log.Error(err)
			return nil, err
		}

		stateMap := make(map[lightning.ChannelStateName]interface{})

		stateMap[lightning.ChannelOpening] = &lightning.ChannelStateOpening{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpeningTime,
			OpenFee:       infoFromDB.OpeningFees,
			CommitFee:     infoFromDB.OpeningCommitFees,
			RemoteBalance: infoFromDB.OpeningRemoteBalance,
			LocalBalance:  infoFromDB.OpeningLocalBalance,
			Initiator:     infoFromDB.OpeningInitiator,
		}

		stateMap[lightning.ChannelOpened] = &lightning.ChannelStateOpened{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.OpenTime,
			CommitFee:     infoFromDB.OpenCommitFees,
			RemoteBalance: infoFromDB.OpenRemoteBalance,
			LocalBalance:  infoFromDB.OpenLocalBalance,
			IsActive:      false,
			StuckBalance:  infoFromDB.OpenStuckBalance,
		}

		stateMap[lightning.ChannelClosing] = &lightning.ChannelStateClosing{
			ChannelID:     chanID,
			CreationTime:  infoFromDB.ClosingTime,
			CloseFee:      infoFromDB.ClosingFees,
			RemoteBalance: infoFromDB.ClosingRemoteBalance,
			LocalBalance:  infoFromDB.ClosingLocalBalance,
			LockedBalance: 0,
		}

		stateMap[lightning.ChannelClosed] = &lightning.ChannelStateClosed{
			ChannelID:    chanID,
			CreationTime: infoFromDB.CloseTime,
			CloseFee:     infoFromDB.ClosingFees,
			LocalBalance: btcutil.Amount(closeSummary.TimeLockedBalance),
		}

		channels = append(channels, &lightning.Channel{
			NodeID:    nodeID,
			ChannelID: chanID,
			State:     lightning.ChannelClosed,
			States:    stateMap,
		})
	}

	return channels, nil
}

// OpenChannel opens the lightning network channel with the given node.
//
// NOTE: Part of the lightning.TopologyClient interface.
func (c *Client) OpenChannel(nodeID lightning.NodeID, funds btcutil.Amount) error {
	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	req := &lnrpc.OpenChannelRequest{
		NodePubkeyString:   string(nodeID),
		LocalFundingAmount: int64(funds),
		PushSat:            0,
		TargetConf:         1,
		Private:            false,
		MinHtlcMsat:        1000,
	}

	if _, err := c.rpc.OpenChannelSync(timeout(200), req); err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable to send open channel request: %v", err)
		log.Error(err)
		return err
	}

	return nil
}

// CloseChannel closes the specified lightning network channel.
//
// NOTE: Part of the lightning.TopologyClient interface.
func (c *Client) CloseChannel(id lightning.ChannelID) error {
	m := crypto.NewMetric(c.cfg.Asset, common.GetFunctionName(),
		c.cfg.MetricsBackend)
	defer m.Finish()

	parts := strings.Split(string(id), ":")
	if len(parts) != 2 {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable to split chan point(%v) on "+
			"funding tx id and output index", id)
		log.Error(err)
		return err
	}

	index, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable decode tx index: %v", err)
		log.Error(err)
		return err
	}

	fundingTx := &lnrpc.ChannelPoint_FundingTxidStr{
		FundingTxidStr: parts[0],
	}

	req := &lnrpc.CloseChannelRequest{
		ChannelPoint: &lnrpc.ChannelPoint{
			FundingTxid: fundingTx,
			OutputIndex: uint32(index),
		},
		Force: false,
	}

	if _, err = c.rpc.CloseChannel(timeout(30), req); err != nil {
		m.AddError(metrics.HighSeverity)
		err := errors.Errorf("unable close the channel: %v", err)
		log.Error(err)
		return err
	}

	return nil
}

// ConnectToNode connects to node with tcp / ip connection.
//
// NOTE: Part of the lightning.TopologyClient interface.
func (c *Client) ConnectToNode(nodeID lightning.NodeID) error {
	pubKey := string(nodeID)

	// Probably we already connected, check it.
	connected, err := c.checkNodeConn(nodeID)
	if err != nil {
		return errors.Errorf("unable check connection: %v", err)
	}

	// Exit if already connected
	if connected {
		return nil
	}

	// Fetch addressed of node we would like connect to.
	var addresses []*lnrpc.NodeAddress
	{
		req := &lnrpc.NodeInfoRequest{PubKey: pubKey}
		resp, err := c.rpc.GetNodeInfo(timeout(30), req)
		if err != nil {
			return errors.Errorf("unable to get node info: %v", err)
		}

		addresses = resp.Node.Addresses
	}

	// For every advertised by node address we have to try to connect to it.
	for _, address := range addresses {
		{
			addr := &lnrpc.LightningAddress{
				Pubkey: pubKey,
				Host:   address.Addr,
			}

			req := &lnrpc.ConnectPeerRequest{
				Addr: addr,
				Perm: false,
			}

			if _, err := c.rpc.ConnectPeer(timeout(30), req); err != nil {
				if !strings.Contains(err.Error(), "already connected to peer") {
					log.Warnf("unable to connect(%v@%v): %v", addr.Pubkey,
						addr.Host, err)
					continue
				}
			}
		}
	}

	// At the end lets check that we are connected to be 100% sure.
	connected, err = c.checkNodeConn(nodeID)
	if err != nil {
		return errors.Errorf("unable check connection: %v", err)
	}

	if connected {
		return nil
	} else {
		return errors.New("unable to connect")
	}
}

// checkNodeConn check that we actually were connected to node, because
// sometimes even if response was without error, remote peer might close
// connect with us for some reason.
func (c *Client) checkNodeConn(nodeID lightning.NodeID) (bool, error) {
	req := &lnrpc.ListPeersRequest{}
	resp, err := c.rpc.ListPeers(timeout(30), req)
	if err != nil {
		return false, errors.Errorf("unable to list peers: %v", err)
	}

	pubKey := string(nodeID)
	for _, node := range resp.Peers {
		if node.PubKey == pubKey {
			return true, nil
		}
	}

	return false, nil
}
