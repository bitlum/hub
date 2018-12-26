package router

import (
	"github.com/bitlum/btcutil"
	"github.com/bitlum/hub/lightning"
)

type PaymentAttempt struct {
	Amount btcutil.Amount
	Route  lightning.Route
	Error  PaymentError
}

type AttemptsStorage interface {
	SaveAttempt()
}

type PaymentCounter struct {
	SuccessPayments int64
	FailPayments    int64
}

type PaymentStatsStorage interface {
	GetNodeCounter(nodeID lightning.NodeID) (PaymentCounter, error)
	GetChannelCounter(channelID lightning.ChannelID) (PaymentCounter, error)
}
