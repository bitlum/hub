package router

import (
	"github.com/btcsuite/btcutil"
)

type nodeEconomy struct {
	// startTime and endTIme define period for which stats has to be
	// aggregated.
	startTime int64
	endTime   int64

	paymentStats

	// qualityRatio shows how much fees we spent to move 1 sat in lightning
	// network. In this case fee will include payment fee as well as channel
	// management fees. This ratio highly depends on type of wallet activity
	// as well as channel management strategy.
	//
	// NOTE: Sanity of this metric increases with payments flows. Initially it
	// is expected for it to be high.
	qualityRatio float64
}

type paymentStats struct {
	// paymentFee is fee which were paid to sent lightning network
	// transactions from our node.
	paymentFee btcutil.Amount

	// earnedForwardFee number of funds which earned by our lightning network
	// node for forwarding payments,during specified period of time.
	earnedForwardFee btcutil.Amount

	// sentFunds is the number of funds which were sent by us to someone else in
	// lightning network, during specified period of time.
	sentFunds btcutil.Amount

	// receivedFunds is the number of funds which were received by us,
	// during specified period of time.
	receivedFunds btcutil.Amount
}
