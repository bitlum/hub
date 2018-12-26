package bitcoind

import (
	"fmt"
	"github.com/bitlum/go-bitcoind-rpc/rpcclient"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/go-errors/errors"
)

type Config struct {
	RPCHost  string
	RPCPort  string
	User     string
	Password string
}

type Explorer struct {
	rpc *rpcclient.Client
}

// NewExplorer creates new instance of bitcoin explorer.
func NewExplorer(cfg *Config) (*Explorer, error) {
	host := fmt.Sprintf("%v:%v", cfg.RPCHost, cfg.RPCPort)

	rpcCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         cfg.User,
		Pass:         cfg.Password,
		DisableTLS:   true,
		HTTPPostMode: true,
	}

	rpc, err := rpcclient.New(rpcCfg, nil)
	if err != nil {
		return nil, errors.Errorf("unable to create rpc client: %v", err)
	}

	return &Explorer{
		rpc: rpc,
	}, nil
}

// FetchTxHeight fetch transaction block height.
func (e *Explorer) FetchTxHeight(txID string) (uint32, error) {
	txHash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		return 0, err
	}

	txVerbose, err := e.rpc.GetRawTransactionVerbose(txHash)
	if err != nil {
		return 0, err
	}

	blockHash, err := chainhash.NewHashFromStr(txVerbose.BlockHash)
	if err != nil {
		return 0, err
	}

	block, err := e.rpc.GetBlockHeaderVerbose(blockHash)
	if err != nil {
		return 0, err
	}

	return uint32(block.Height), nil
}

// FetchTxFee fetch transaction miners fee.
func (e *Explorer) FetchTxFee(txID string) (btcutil.Amount, error) {
	txHash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		return 0, err
	}

	tx, err := e.rpc.GetRawTransaction(txHash)
	if err != nil {
		return 0, err
	}

	var overallInput btcutil.Amount
	for _, input := range tx.MsgTx().TxIn {
		outpoint := input.PreviousOutPoint

		prevTx, err := e.rpc.GetRawTransaction(&outpoint.Hash)
		if err != nil {
			return 0, err
		}

		output := prevTx.MsgTx().TxOut[outpoint.Index]
		overallInput += btcutil.Amount(output.Value)
	}

	var overallOutput btcutil.Amount
	for _, output := range tx.MsgTx().TxOut {
		overallOutput += btcutil.Amount(output.Value)
	}

	fee := overallInput - overallOutput
	return btcutil.Amount(fee), nil

}

// FetchTxIndexInBlock fetch transaction position withing block.
func (e *Explorer) FetchTxIndexInBlock(txID string) (uint32, error) {
	txHash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		return 0, err
	}

	txVerbose, err := e.rpc.GetRawTransactionVerbose(txHash)
	if err != nil {
		return 0, err
	}

	blockHash, err := chainhash.NewHashFromStr(txVerbose.BlockHash)
	if err != nil {
		return 0, err
	}

	block, err := e.rpc.GetBlockVerbose(blockHash)
	if err != nil {
		return 0, err
	}

	for index, tx := range block.Tx {
		if tx == txID {
			return uint32(index), nil
		}
	}

	return 0, errors.New("unable to find index in block")
}

// FetchTxTime fetch transaction receive time, when transaction first
// appeared in blockchain.
func (e *Explorer) FetchTxTime(txID string) (int64, error) {
	txHash, err := chainhash.NewHashFromStr(txID)
	if err != nil {
		return 0, err
	}

	txVerbose, err := e.rpc.GetRawTransactionVerbose(txHash)
	if err != nil {
		return 0, err
	}

	return txVerbose.Time, nil
}
