package lnd

import (
	"github.com/lightningnetwork/lnd/lnrpc"
	"time"
	"context"
	"github.com/go-errors/errors"
)

// fetchUsers fetches all peers connected to us with tcp/ip connection.
func fetchUsers(c lnrpc.LightningClient) ([]*lnrpc.Peer, error) {
	reqInfo := &lnrpc.ListPeersRequest{}
	ctx, _ := context.WithTimeout(getContext(), time.Second*5)
	resp, err := c.ListPeers(ctx, reqInfo)
	if err != nil {
		return nil, err
	}

	return resp.Peers, nil
}

// fetchChannels fetches all channels connected to the lightning network daemon.
func fetchChannels(c lnrpc.LightningClient) (
	[]*lnrpc.Channel,
	[]*lnrpc.PendingChannelsResponse_PendingOpenChannel,
	[]*lnrpc.PendingChannelsResponse_ClosedChannel,
	[]*lnrpc.PendingChannelsResponse_ForceClosedChannel,
	[]*lnrpc.PendingChannelsResponse_WaitingCloseChannel,
	error) {

	var respListChannels *lnrpc.ListChannelsResponse
	var err error

	{
		reqInfo := &lnrpc.ListChannelsRequest{}
		ctx, _ := context.WithTimeout(getContext(), time.Second*5)
		respListChannels, err = c.ListChannels(ctx, reqInfo)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}

	var respPendingChannels *lnrpc.PendingChannelsResponse
	{
		reqInfo := &lnrpc.PendingChannelsRequest{}
		ctx, _ := context.WithTimeout(getContext(), time.Second*5)
		respPendingChannels, err = c.PendingChannels(ctx, reqInfo)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
	}

	return respListChannels.Channels,
		respPendingChannels.PendingOpenChannels,
		respPendingChannels.PendingClosingChannels,
		respPendingChannels.PendingForceClosingChannels,
		respPendingChannels.WaitingCloseChannels, nil
}

// fetchInvoicePayments fetches the information about invoices which were
// created by lightning network node, and its state.
func fetchInvoicePayments(c lnrpc.LightningClient) (
	[]*lnrpc.Invoice, error) {
	reqInfo := &lnrpc.ListInvoiceRequest{}
	ctx, _ := context.WithTimeout(getContext(), time.Second*5)

	resp, err := c.ListInvoices(ctx, reqInfo)
	if err != nil {
		return nil, err
	}

	return resp.Invoices, nil
}

// fetchNodeInfo fetched information about hub node.
func fetchNodeInfo(c lnrpc.LightningClient) (
	*lnrpc.GetInfoResponse, error) {
	reqInfo := &lnrpc.GetInfoRequest{}
	ctx, _ := context.WithTimeout(getContext(), time.Second*5)
	return c.GetInfo(ctx, reqInfo)
}

// fetchOutgoingPayments fetched the list of payments which is going from hub,
// to users.
func fetchOutgoingPayments(c lnrpc.LightningClient) ([]*lnrpc.Payment,
	error) {
	reqInfo := &lnrpc.ListPaymentsRequest{}
	ctx, _ := context.WithTimeout(getContext(), time.Second*5)

	resp, err := c.ListPayments(ctx, reqInfo)
	if err != nil {
		return nil, err
	}

	return resp.Payments, nil
}

// fetchForwardingPayments gradually fetches the forwarding events from lightning
// daemon.
func fetchForwardingPayments(c lnrpc.LightningClient, index uint32) (
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

		resp, err := c.ForwardingHistory(getContext(), req)
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

// getPubKeyByChainID returns the pubkey which identifies the user by the
// given channel id.
func getPubKeyByChainID(c lnrpc.LightningClient, chanID uint64) (string, error) {
	req := &lnrpc.ListChannelsRequest{}
	resp, err := c.ListChannels(getContext(), req)
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
