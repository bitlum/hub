package lnd

import (
	"context"
	"github.com/bitlum/hub/lightning"
	"github.com/bitlum/hub/lightning/lnd/explorer"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnwire"
	"strconv"
	"strings"
	"time"
)

// panicRecovering is needed to ensure that our program not stops because of
// the panic, also this is needed to be able to properly send, alert to the
// metric server, because if metric server will be unable to scrape the metric
// than we wouldn't be able to see that on service radar.
func panicRecovering() {
	if r := recover(); r != nil {
		log.Error(r)
	}
}

func getParams(netName string) (*chaincfg.Params, error) {
	switch netName {
	case "mainnet", "main":
		return &chaincfg.MainNetParams, nil
	case "regtest", "simnet":
		return &chaincfg.RegressionNetParams, nil
	case "testnet3", "test", "testnet":
		return &chaincfg.TestNet3Params, nil
	}

	return nil, errors.Errorf("network '%s' is invalid or unsupported",
		netName)
}

func timeout(sec int) context.Context {
	btcx := context.Background()
	ctx, _ := context.WithTimeout(btcx, time.Second*time.Duration(sec))
	return ctx
}

// getOpenInitiator returns initiator of the channel creation.
// TODO(andrew.shvv) it will not work after dual funding, remove it
func getOpenInitiator(localBalance int64) lightning.ChannelInitiator {
	if localBalance == 0 {
		return lightning.RemoteInitiator
	} else {
		return lightning.LocalInitiator
	}
}

func splitChannelPoint(channelPoint string) (string, string, error) {
	parts := strings.Split(channelPoint, ":")
	if len(parts) != 2 {
		return "", "", errors.Errorf("unable split channel point(%v)",
			channelPoint)
	}

	txID := parts[0]
	txIndex := parts[1]

	return txID, txIndex, nil
}

func getStuckBalance(htlcs []*lnrpc.HTLC) btcutil.Amount {
	var overallStuck btcutil.Amount
	for _, htlc := range htlcs {
		overallStuck += btcutil.Amount(htlc.Amount)
	}

	return overallStuck
}

func getShortChannelIDByChannelPoint(e explorer.Explorer, channelPoint string) (
	uint64, error) {
	txID, txPositionStr, err := splitChannelPoint(channelPoint)
	if err != nil {
		return 0, err
	}

	txPosition, err := strconv.Atoi(txPositionStr)
	if err != nil {
		return 0, err
	}

	blockHeight, err := e.FetchTxHeight(txID)
	if err != nil {
		return 0, err
	}

	blockIndex, err := e.FetchTxIndexInBlock(txID)
	if err != nil {
		return 0, err
	}

	chanID := lnwire.ShortChannelID{
		BlockHeight: blockHeight,
		TxIndex:     blockIndex,
		TxPosition:  uint16(txPosition),
	}

	return chanID.ToUint64(), nil
}
