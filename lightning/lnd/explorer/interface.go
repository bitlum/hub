package explorer

import "github.com/btcsuite/btcutil"

// Explorer is the entity which gives information about bitcoin blockchain.
type Explorer interface {
	// FetchTxHeight fetch transaction block height.
	FetchTxHeight(txID string) (uint32, error)

	// FetchTxFee fetch transaction miners fee.
	FetchTxFee(txID string) (btcutil.Amount, error)

	// FetchTxIndexInBlock fetch transaction position withing block.
	FetchTxIndexInBlock(txID string) (uint32, error)

	// FetchTxTime fetch transaction receive time, when transaction first
	// appeared in blockchain.
	FetchTxTime(txID string) (int64, error)
}
