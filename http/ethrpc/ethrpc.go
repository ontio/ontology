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
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	oComm "github.com/ontio/ontology/common"
	bactor "github.com/ontio/ontology/http/base/actor"
	hComm "github.com/ontio/ontology/http/base/common"
	types2 "github.com/ontio/ontology/http/ethrpc/types"
)

const (
	eth65           = 65
	ProtocolVersion = eth65
	ChainId         = 5851
)

type EthereumAPI struct {
}

func (api *EthereumAPI) ChainId() hexutil.Uint64 {
	return hexutil.Uint64(ChainId)
}

func (api *EthereumAPI) BlockNumber() (hexutil.Uint64, error) {
	// height := bactor.GetCurrentBlockHeight()
	return hexutil.Uint64(1000), nil
}

func (api *EthereumAPI) GetBalance(address common.Address, _ rpc.BlockNumberOrHash) (*hexutil.Big, error) {
	//balances, _, err := hComm.GetContractBalance(0, []oComm.Address{utils.OngContractAddress}, oComm.Address(address), true)
	//if err != nil {
	//	return nil, fmt.Errorf("get ong balance error:%s", err)
	//}
	return (*hexutil.Big)(big.NewInt(0)), nil
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
	start := bactor.GetCurrentBlockHeight()
	var gasPrice uint64 = 0
	var end uint32 = 0
	if start > hComm.MAX_SEARCH_HEIGHT {
		end = start - hComm.MAX_SEARCH_HEIGHT
	}
	for i := start; i >= end; i-- {
		head, err := bactor.GetHeaderByHeight(i)
		if err == nil && head.TransactionsRoot != oComm.UINT256_EMPTY {
			blk, err := bactor.GetBlockByHeight(i)
			if err != nil {
				return nil
			}
			for _, v := range blk.Transactions {
				gasPrice += v.GasPrice
			}
			gasPrice = gasPrice / uint64(len(blk.Transactions))
			break
		}
	}
	return (*hexutil.Big)(new(big.Int).SetUint64(gasPrice))
}

// TODO
func (api *EthereumAPI) Accounts() ([]common.Address, error) {
	return nil, nil
}

// TODO
func (api *EthereumAPI) GetStorageAt(address common.Address, key string, blockNum types2.BlockNumber) (hexutil.Bytes, error) {
	return nil, nil
}

// TODO
func (api *EthereumAPI) GetTransactionCount(address common.Address, blockNum int64) (*hexutil.Uint64, error) {
	nonce := hexutil.Uint64(12321)
	print("12321")
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
	deployCode, err := bactor.GetContractStateFromStore(oComm.Address(address))
	if err != nil {
		return nil, err
	}
	if deployCode == nil {
		return nil, fmt.Errorf("code: %v not found", address)
	}
	code := deployCode.GetRawCode()
	return code, nil

}

// TODO
func (api *EthereumAPI) GetTransactionLogs(txHash common.Hash) ([]*types.Log, error) {
	return nil, nil
}

// TODO
func (api *EthereumAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	return nil, nil
}

// TODO
func (api *EthereumAPI) SendTransaction(args types2.SendTxArgs) (common.Hash, error) {
	return [32]byte{}, nil
}

func (api *EthereumAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	return [32]byte{}, nil
}

func (api *EthereumAPI) Call(args types2.CallArgs, blockNumber int64, _ *map[common.Address]types2.Account) (hexutil.Bytes, error) {
	return nil, nil
}

func (api *EthereumAPI) EstimateGas(args types2.CallArgs) (hexutil.Uint, error) {
	return 0, nil
}

func (api *EthereumAPI) GetBlockByHash(hash common.Hash, fullTx bool) (interface{}, error) {
	block, err := bactor.GetBlockFromStore(oComm.Uint256(hash))
	if err != nil {
		return nil, err
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
	return map[string]interface{}{
		"number":           hexutil.Uint64(100000),
		"hash":             hexutil.Bytes{},
		"parentHash":       hexutil.Bytes{},
		"nonce":            types.BlockNonce{}, // PoW specific
		"sha3Uncles":       common.Hash{},      // No uncles in Tendermint
		"logsBloom":        types2.Bloom{},
		"transactionsRoot": hexutil.Bytes{},
		"stateRoot":        hexutil.Bytes{},
		"miner":            common.Address{},
		"mixHash":          common.Hash{},
		"difficulty":       hexutil.Uint64(0),
		"totalDifficulty":  hexutil.Uint64(0),
		"extraData":        hexutil.Bytes{},
		"size":             hexutil.Uint64(0),
		"gasLimit":         hexutil.Uint64(0), // TODO Static gas limit
		"gasUsed":          (*hexutil.Big)(big.NewInt(0)),
		"timestamp":        hexutil.Uint64(0),
		"uncles":           []string{},
		"receiptsRoot":     common.Hash{},
	}, nil
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
	return OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx)), nil
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
	return OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx)), nil
}

func (api *EthereumAPI) GetTransactionByBlockNumberAndIndex(blockNum types2.BlockNumber, idx hexutil.Uint) (*types2.Transaction, error) {
	return nil, nil
}

func (api *EthereumAPI) GetTransactionReceipt(hash common.Hash) (interface{}, error) {
	return nil, nil
}

func (api *EthereumAPI) PendingTransactions() ([]*types2.Transaction, error) {
	return nil, nil
}

func (api *EthereumAPI) PendingTransactionByHash(target common.Hash) (*types2.Transaction, error) {
	return nil, nil
}

func (api *EthereumAPI) GetUncleByBlockHashAndIndex(_ common.Hash, _ hexutil.Uint) map[string]interface{} {
	return nil
}

func (api *EthereumAPI) GetUncleByBlockNumberAndIndex(_ hexutil.Uint, _ hexutil.Uint) map[string]interface{} {
	return nil
}

func (api *EthereumAPI) GetProof(address common.Address, storageKeys []string, block types2.BlockNumber) (*types2.AccountResult, error) {
	return nil, nil
}

type PublicNetAPI struct {
	networkVersion uint64
}

var networkVersion = 1

// Version returns the current ethereum protocol version.
func (s *PublicNetAPI) Version() string {
	return fmt.Sprintf("%d", networkVersion)
}
