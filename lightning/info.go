package lightning

import "github.com/btcsuite/btcutil"

type NodeInfo struct {
	Alias          string
	Host           string

	Port           string
	IdentityPubKey string
}

type NeutrinoInfo struct {
	Host string
	Port string
}

type Info struct {
	MinPaymentAmount btcutil.Amount
	MaxPaymentAmount btcutil.Amount

	Version     string
	Network     string
	BlockHeight uint32
	BlockHash   string

	NodeInfo     *NodeInfo
	NeutrinoInfo *NeutrinoInfo
}
