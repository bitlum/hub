package lightning

import "github.com/btcsuite/btcutil"

type Route struct {
	Nodes       []Node
	Channels    []Channel
	TotalFee    btcutil.Amount
	TotalAmount btcutil.Amount
}
