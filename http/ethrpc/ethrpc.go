/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
package ethrpc

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	oComm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	otypes "github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	bcomn "github.com/ontio/ontology/http/base/common"
	hComm "github.com/ontio/ontology/http/base/common"
	types2 "github.com/ontio/ontology/http/ethrpc/types"
	"github.com/ontio/ontology/smartcontract/event"
	types3 "github.com/ontio/ontology/smartcontract/service/evm/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const (
	eth65           = 65
	ProtocolVersion = eth65
	RPCGasCap       = 0 // TODO modify this
)

type TxPoolService interface {
	Nonce(addr oComm.Address) uint64
	PendingEIPTransactions() map[common.Address]map[uint64]*types.Transaction
}

type EthereumAPI struct {
	txpool TxPoolService
}

func NewEthereumAPI(txpool TxPoolService) *EthereumAPI {
	return &EthereumAPI{txpool: txpool}
}

func (api *EthereumAPI) ChainId() hexutil.Uint64 {
	return hexutil.Uint64(getChainId())
}

func (api *EthereumAPI) BlockNumber() (hexutil.Uint64, error) {
	height := bactor.GetCurrentBlockHeight()
	return hexutil.Uint64(height), nil
}

func (api *EthereumAPI) GetBalance(address common.Address, _ rpc.BlockNumberOrHash) (*hexutil.Big, error) {
	balances, _, err := hComm.GetContractBalance(0, []oComm.Address{utils.OngContractAddress}, oComm.Address(address), true)
	if err != nil {
		return nil, fmt.Errorf("get ong balance error:%s", err)
	}
	return (*hexutil.Big)(big.NewInt(int64(balances[0]))), nil
}

func (api *EthereumAPI) ProtocolVersion() hexutil.Uint {
	return hexutil.Uint(ProtocolVersion)
}

func (api *EthereumAPI) Syncing() (interface{}, error) {
	curBlock := bactor.GetCurrentBlockHeight()
	peerBlock := bactor.GetMaxPeerBlockHeight()
	return map[string]interface{}{
		"startingBlock": hexutil.Uint64(0),
		"currentBlock":  hexutil.Uint64(curBlock),
		"highestBlock":  hexutil.Uint64(peerBlock), // NA
		// "pulledStates":  nil, // NA
		// "knownStates":   nil, // NA
	}, nil
}

func (api *EthereumAPI) Coinbase() (common.Address, error) {
	return [20]byte{}, nil
}

func (api *EthereumAPI) Mining() bool {
	return false
}

func (api *EthereumAPI) Hashrate() hexutil.Uint64 {
	return 0
}

func (api *EthereumAPI) GasPrice() *hexutil.Big {
	gasPrice, _, err := hComm.GetGasPrice()
	if err != nil {
		return nil
	}
	return (*hexutil.Big)(new(big.Int).SetUint64(gasPrice))
}

func (api *EthereumAPI) Accounts() ([]common.Address, error) {
	return nil, fmt.Errorf("eth_accounts is not supported")
}

func (api *EthereumAPI) GetStorageAt(address common.Address, key string, blockNum types2.BlockNumber) (hexutil.Bytes, error) {
	return bactor.GetEthStorage(address, common.HexToHash(key))
}

func (api *EthereumAPI) GetTransactionCount(address common.Address, blockNum types2.BlockNumber) (*hexutil.Uint64, error) {
	addr := EthToOntAddr(address)
	if nonce := api.txpool.Nonce(addr); blockNum.IsPending() && nonce != 0 {
		n := hexutil.Uint64(nonce)
		return &n, nil
	}
	account, err := bactor.GetEthAccount(address)
	if err != nil {
		return nil, err
	}
	nonce := hexutil.Uint64(account.Nonce)
	return &nonce, nil
}

func (api *EthereumAPI) GetBlockTransactionCountByHash(hash common.Hash) *hexutil.Uint {
	block, err := bactor.GetBlockFromStore(oComm.Uint256(hash))
	if err != nil {
		return nil
	}
	if block == nil {
		return nil
	}
	txCount := hexutil.Uint(len(block.Transactions))
	return &txCount
}

func (api *EthereumAPI) GetBlockTransactionCountByNumber(number int64) *hexutil.Uint {
	block, err := bactor.GetBlockByHeight(uint32(number))
	if err != nil {
		return nil
	}
	if block == nil {
		return nil
	}
	txCount := hexutil.Uint(len(block.Transactions))
	return &txCount
}

func (api *EthereumAPI) GetUncleCountByBlockHash(hash common.Hash) hexutil.Uint {
	return 0
}

func (api *EthereumAPI) GetUncleCountByBlockNumber(number int64) hexutil.Uint {
	return 0
}

func (api *EthereumAPI) GetCode(address common.Address, blockNumber types2.BlockNumber) (hexutil.Bytes, error) {
	account, err := bactor.GetEthAccount(address)
	if err != nil {
		return nil, err
	}
	if account.IsEmpty() {
		return nil, fmt.Errorf("contract %v not found", address.String())
	}
	code, err := bactor.GetEthCode(account.CodeHash)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(code), nil
}

func (api *EthereumAPI) GetTransactionLogs(txHash common.Hash) ([]*types.Log, error) {
	notify, err := bactor.GetEventNotifyByTxHash(EthToOntHash(txHash))
	if err != nil {
		return nil, err
	}
	if notify == nil {
		return nil, fmt.Errorf("tx %v not found", txHash.String())
	}
	return generateLog(notify)
}

func generateLog(rawNotify *event.ExecuteNotify) ([]*types.Log, error) {
	var res []*types.Log
	txHash := rawNotify.TxHash
	height, _, err := bactor.GetTxnWithHeightByTxHash(txHash)
	if err != nil {
		return nil, err
	}
	hash := bactor.GetBlockHashFromStore(height)
	if err != nil {
		return nil, err
	}
	for idx, n := range rawNotify.Notify {
		if !n.IsEvm {
			return nil, fmt.Errorf("not support tx type %v", rawNotify.TxHash.ToHexString())
		}
		source := oComm.NewZeroCopySource(n.States.([]byte))
		var storageLog otypes.StorageLog
		err := storageLog.Deserialization(source)
		if err != nil {
			return nil, err
		}
		log := &types.Log{
			Address:     storageLog.Address,
			Topics:      storageLog.Topics,
			Data:        storageLog.Data,
			BlockNumber: uint64(height),
			TxHash:      OntToEthHash(txHash),
			TxIndex:     uint(rawNotify.TxIndex),
			BlockHash:   OntToEthHash(hash),
			Index:       uint(idx),
			Removed:     false,
		}
		res = append(res, log)
	}
	return res, nil
}

func (api *EthereumAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	return nil, fmt.Errorf("eth_sign is not supported")
}

func (api *EthereumAPI) SendTransaction(args types2.SendTxArgs) (common.Hash, error) {
	return [32]byte{}, fmt.Errorf("eth_sendTransaction is not supported")
}

func (api *EthereumAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(data, tx); err != nil {
		return common.Hash{}, err
	}

	eip155tx, err := otypes.TransactionFromEIP155(tx)
	if err != nil {
		return common.Hash{}, err
	}
	txhash := eip155tx.Hash()
	if errCode, desc := bcomn.SendTxToPool(eip155tx); errCode != ontErrors.ErrNoError {
		log.Warnf("SendRawTransaction verified %s error: %s", txhash.ToHexString(), desc)
		return common.Hash{}, err
	}
	return common.Hash(txhash), nil
}

func (api *EthereumAPI) Call(args types2.CallArgs, blockNumber types2.BlockNumber, _ *map[common.Address]types2.Account) (hexutil.Bytes, error) {
	tx := args.AsTransaction(RPCGasCap)
	res, err := bactor.PreExecuteEip155Tx(tx)
	if err != nil {
		return nil, err
	}
	if len(res.Revert()) > 0 {
		return nil, newRevertError(res)
	}
	return res.Return(), res.Err
}

func newRevertError(result *types3.ExecutionResult) *revertError {
	reason, errUnpack := abi.UnpackRevert(result.Revert())
	err := errors.New("execution reverted")
	if errUnpack == nil {
		err = fmt.Errorf("execution reverted: %v", reason)
	}
	return &revertError{
		error:  err,
		reason: hexutil.Encode(result.Revert()),
	}
}

type revertError struct {
	error
	reason string // revert reason hex encoded
}

func (api *EthereumAPI) EstimateGas(args types2.CallArgs) (hexutil.Uint, error) {
	tx := args.AsTransaction(RPCGasCap)
	res, err := bactor.PreExecuteEip155Tx(tx)
	if err != nil {
		return 0, err
	}
	return hexutil.Uint(res.UsedGas), nil
}

func (api *EthereumAPI) GetBlockByHash(hash common.Hash, fullTx bool) (interface{}, error) {
	block, err := bactor.GetBlockFromStore(oComm.Uint256(hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("block: %v not found", hash.String())
	}
	return EthBlockFromOntology(block, fullTx), nil
}

func (api *EthereumAPI) GetBlockByNumber(blockNum types2.BlockNumber, fullTx bool) (interface{}, error) {
	height := uint32(blockNum)
	if blockNum.IsLatest() || blockNum.IsPending() {
		height = bactor.GetCurrentBlockHeight()
	}
	block, err := bactor.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("block: %v not found", blockNum.Int64())
	}
	return EthBlockFromOntology(block, fullTx), nil
}

func (api *EthereumAPI) GetTransactionByHash(hash common.Hash) (*types2.Transaction, error) {
	height, tx, err := bactor.GetTxnWithHeightByTxHash(oComm.Uint256(hash))
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, fmt.Errorf("tx: %v not found", hash.Hex())
	}
	block, err := bactor.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("block: %v not found", height)
	}
	header := block.Header
	blockHash := header.Hash()
	txs := block.Transactions
	idx := 0
	for i, t := range txs {
		txHash := t.Hash()
		if bytes.Equal(txHash.ToArray(), hash.Bytes()) {
			idx = i
			break
		}
	}
	return OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx))
}

func (api *EthereumAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*types2.Transaction, error) {
	block, err := bactor.GetBlockFromStore(oComm.Uint256(hash))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("block: %v not found", hash.Hex())
	}
	header := block.Header
	blockHash := header.Hash()
	txs := block.Transactions
	if len(txs) >= int(idx) {
		return nil, fmt.Errorf("access block: %v overflow %v", hash.Hex(), idx)
	}
	tx := txs[idx]
	return OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx))
}

func (api *EthereumAPI) GetTransactionByBlockNumberAndIndex(blockNum types2.BlockNumber, idx hexutil.Uint) (*types2.Transaction, error) {
	height := uint32(blockNum)
	if blockNum.IsLatest() || blockNum.IsPending() {
		height = bactor.GetCurrentBlockHeight()
	}
	block, err := bactor.GetBlockByHeight(height)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("block: %v not found", height)
	}
	header := block.Header
	blockHash := header.Hash()
	txs := block.Transactions
	if len(txs) >= int(idx) {
		return nil, fmt.Errorf("access block: %v overflow %v", height, idx)
	}
	tx := txs[idx]
	return OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx))
}

func (api *EthereumAPI) GetTransactionReceipt(hash common.Hash) (interface{}, error) {
	return nil, nil
}

func (api *EthereumAPI) PendingTransactions() ([]*types2.Transaction, error) {
	pendingTxs := api.txpool.PendingEIPTransactions()
	var rpcTxs []*types2.Transaction
	for _, v1 := range pendingTxs {
		for _, v2 := range v1 {
			tx, err := NewTransaction(v2, v2.Hash(), common.Hash{}, 0, 0)
			if err != nil {
				return nil, nil
			}
			rpcTxs = append(rpcTxs, tx)
		}
	}
	return rpcTxs, nil
}

func (api *EthereumAPI) PendingTransactionsByHash(target common.Hash) (*types2.Transaction, error) {
	pendingTxs := api.txpool.PendingEIPTransactions()
	var ethTx *types.Transaction
	for _, v1 := range pendingTxs {
		for _, v2 := range v1 {
			if v2.Hash() == target {
				ethTx = v2
				break
			}
		}
	}
	if ethTx == nil {
		return nil, fmt.Errorf("tx: %v not found", target.String())
	}
	return NewTransaction(ethTx, ethTx.Hash(), common.Hash{}, 0, 0)
}

func (api *EthereumAPI) GetUncleByBlockHashAndIndex(_ common.Hash, _ hexutil.Uint) map[string]interface{} {
	return nil
}

func (api *EthereumAPI) GetUncleByBlockNumberAndIndex(_ hexutil.Uint, _ hexutil.Uint) map[string]interface{} {
	return nil
}

func (api *EthereumAPI) GetProof(address common.Address, storageKeys []string, block types2.BlockNumber) (*types2.AccountResult, error) {
	return nil, fmt.Errorf("eth_getProof is not supported")
}

type PublicNetAPI struct {
	networkVersion uint64
}

var networkVersion = 1

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", networkVersion)
}
