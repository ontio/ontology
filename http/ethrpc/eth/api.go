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
package eth

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
	"github.com/laizy/bigint"
	oComm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/states"
	common2 "github.com/ontio/ontology/core/store/common"
	sCom "github.com/ontio/ontology/core/store/common"
	otypes "github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	bactor "github.com/ontio/ontology/http/base/actor"
	hComm "github.com/ontio/ontology/http/base/common"
	types2 "github.com/ontio/ontology/http/ethrpc/types"
	utils2 "github.com/ontio/ontology/http/ethrpc/utils"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/evm"
	types3 "github.com/ontio/ontology/smartcontract/service/evm/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	errors2 "github.com/ontio/ontology/vm/evm/errors"
	"github.com/ontio/ontology/vm/evm/params"
)

const (
	eth65           = 65
	ProtocolVersion = eth65
	RPCGasCap       = config.DEFAULT_ETH_TX_MAX_GAS_LIMIT
)

type TxPoolService interface {
	Nonce(addr oComm.Address) uint64
	PendingEIPTransactions() []*types.Transaction
	PendingTransactionsByHash(target common.Hash) *types.Transaction
	GetGasPrice() uint64
}

type EthereumAPI struct {
	txpool TxPoolService
}

func NewEthereumAPI(txpool TxPoolService) *EthereumAPI {
	return &EthereumAPI{txpool: txpool}
}

func (api *EthereumAPI) ChainId() hexutil.Uint64 {
	log.Debug("eth_chainId")
	return hexutil.Uint64(utils2.GetChainId())
}

func (api *EthereumAPI) BlockNumber() (hexutil.Uint64, error) {
	log.Debug("eth_blockNumber")
	height := bactor.GetCurrentBlockHeight()
	return hexutil.Uint64(height), nil
}

func (api *EthereumAPI) GetBalance(address common.Address, _ rpc.BlockNumberOrHash) (*hexutil.Big, error) {
	log.Debugf("eth_getBalance address %v", address.Hex())
	balance, err := getOngBalance(address)
	if err != nil {
		return (*hexutil.Big)(big.NewInt(0)), err
	}
	return (*hexutil.Big)(balance.ToBigInt()), err
}

func getOngBalance(address common.Address) (states.NativeTokenBalance, error) {
	balances, _, err := hComm.GetNativeTokenBalance(0, []oComm.Address{utils.OngContractAddress}, oComm.Address(address), true)
	if err != nil {
		return states.NativeTokenBalance{}, fmt.Errorf("get ong balance error:%s", err)
	}
	return balances[0], nil
}

func (api *EthereumAPI) ProtocolVersion() hexutil.Uint {
	log.Debug("eth_protocolVersion")
	return hexutil.Uint(ProtocolVersion)
}

func (api *EthereumAPI) Syncing() (interface{}, error) {
	log.Debug("eth_syncing")
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
	log.Debug("eth_gasPrice")
	gasPrice, _, err := hComm.GetGasPrice()
	if err != nil {
		return nil
	}
	return (*hexutil.Big)(bigint.New(gasPrice).Mul(constants.GWei).BigInt())
}

func (api *EthereumAPI) Accounts() ([]common.Address, error) {
	return nil, fmt.Errorf("eth_accounts is not supported")
}

func (api *EthereumAPI) GetStorageAt(address common.Address, key string, blockNum types2.BlockNumber) (hexutil.Bytes, error) {
	log.Debugf("eth_getStorageAt address %v, key %s, blockNum %v", address.Hex(), key, blockNum)
	return bactor.GetEthStorage(address, common.HexToHash(key))
}

func (api *EthereumAPI) GetTransactionCount(address common.Address, blockNum types2.BlockNumber) (*hexutil.Uint64, error) {
	log.Debugf("eth_getTransactionCount address %v, blockNum %v", address.Hex(), blockNum)
	addr := utils2.EthToOntAddr(address)
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
	log.Debugf("eth_getBlockTransactionCountByHash hash %v", hash.Hex())
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
	log.Debugf("eth_getBlockTransactionCountByNumber number %d", number)
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
	log.Debugf("eth_getCode address %s, blockNumber %v", address.Hex(), blockNumber)
	account, err := bactor.GetEthAccount(address)
	if err != nil {
		return nil, err
	}
	if account.IsEmptyContract() {
		return nil, nil
	}
	code, err := bactor.GetEthCode(account.CodeHash)
	if err != nil {
		return nil, err
	}
	return hexutil.Bytes(code), nil
}

func (api *EthereumAPI) GetTransactionLogs(txHash common.Hash) ([]*types.Log, error) {
	log.Debugf("eth_getTransactionLogs txHash %v", txHash.Hex())
	notify, err := bactor.GetEventNotifyByTxHash(utils2.EthToOntHash(txHash))
	if err != nil {
		return nil, err
	}
	if notify == nil {
		return nil, fmt.Errorf("tx %v not found", txHash.String())
	}
	logs, _, _, _, err := generateLog(notify)
	return logs, err
}

func generateLog(rawNotify *event.ExecuteNotify) ([]*types.Log, *common.Hash, *otypes.Transaction, uint32, error) {
	var res []*types.Log
	txHash := rawNotify.TxHash
	height, tx, err := bactor.GetTxnWithHeightByTxHash(txHash)
	if err != nil {
		return nil, nil, nil, 0, err
	}
	hash := bactor.GetBlockHashFromStore(height)
	ethHash := utils2.OntToEthHash(hash)
	for idx, n := range rawNotify.Notify {
		storageLog, err := event.NotifyEventInfoToEvmLog(n)
		if err != nil {
			return nil, nil, nil, 0, err
		}
		res = append(res,
			&types.Log{
				Address:     storageLog.Address,
				Topics:      storageLog.Topics,
				Data:        storageLog.Data,
				BlockNumber: uint64(height),
				TxHash:      utils2.OntToEthHash(txHash),
				TxIndex:     uint(rawNotify.TxIndex),
				BlockHash:   ethHash,
				Index:       uint(idx),
				Removed:     false,
			})
	}
	return res, &ethHash, tx, height, nil
}

func (api *EthereumAPI) Sign(address common.Address, data hexutil.Bytes) (hexutil.Bytes, error) {
	return nil, fmt.Errorf("eth_sign is not supported")
}

func (api *EthereumAPI) SendTransaction(args types2.SendTxArgs) (common.Hash, error) {
	return [32]byte{}, fmt.Errorf("eth_sendTransaction is not supported")
}

func (api *EthereumAPI) SendRawTransaction(data hexutil.Bytes) (common.Hash, error) {
	log.Debugf("eth_sendRawTransaction data %v", data.String())
	tx := new(types.Transaction)
	if err := rlp.DecodeBytes(data, tx); err != nil {
		return common.Hash{}, err
	}

	eip155tx, err := otypes.TransactionFromEIP155(tx)
	if err != nil {
		return common.Hash{}, err
	}
	txhash := eip155tx.Hash()
	if errCode, desc := hComm.SendTxToPool(eip155tx); errCode != ontErrors.ErrNoError {
		if errCode == ontErrors.ErrDuplicatedTx {
			return common.Hash(txhash), nil
		}
		return common.Hash{}, fmt.Errorf("SendRawTransaction verified %s error: %s", txhash.ToHexString(), desc)
	}
	return common.Hash(txhash), nil
}

func (api *EthereumAPI) Call(args types2.CallArgs, blockNumber types2.BlockNumber, _ *map[common.Address]types2.Account) (hexutil.Bytes, error) {
	log.Debugf("eth_call block number %v ", blockNumber)
	msg := args.AsMessage(RPCGasCap)
	res, err := bactor.PreExecuteEip155Tx(msg)
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

func (api *EthereumAPI) EstimateGas(args types2.CallArgs, _ *rpc.BlockNumberOrHash) (hexutil.Uint64, error) {
	var (
		lo  uint64 = params.TxGas
		hi  uint64
		cap uint64
	)
	if args.From == nil {
		args.From = new(common.Address)
	}

	if args.Gas != nil && uint64(*args.Gas) >= params.TxGas && uint64(*args.Gas) <= RPCGasCap {
		hi = uint64(*args.Gas)
	} else {
		hi = RPCGasCap
	}

	if args.GasPrice != nil && args.GasPrice.ToInt().BitLen() != 0 {
		balance, _ := getOngBalance(*args.From)
		available := balance.ToBigInt()
		if args.Value != nil {
			if args.Value.ToInt().Cmp(available) >= 0 {
				return 0, errors.New("insufficient funds for transfer")
			}
			available.Sub(available, args.Value.ToInt())
		}
		allowance := new(big.Int).Div(available, args.GasPrice.ToInt())
		if allowance.IsUint64() && hi > allowance.Uint64() {
			transfer := args.Value
			if transfer == nil {
				transfer = new(hexutil.Big)
			}
			log.Warn("Gas estimation capped by limited funds", "original", hi, "balance", balance,
				"sent", transfer.ToInt(), "gasprice", args.GasPrice.ToInt(), "fundable", allowance)
			hi = allowance.Uint64()
		}
	}
	cap = hi
	executable := func(gas uint64) (bool, *types3.ExecutionResult, error) {
		args.Gas = (*hexutil.Uint64)(&gas)
		tx := args.AsMessage(RPCGasCap)
		result, err := bactor.PreExecuteEip155Tx(tx)
		if err != nil {
			if errors.Is(err, evm.ErrIntrinsicGas) {
				return true, nil, nil
			}
			return true, nil, err
		}
		return result.Failed(), result, nil
	}

	for lo+1 < hi {
		mid := (hi + lo) / 2
		failed, _, err := executable(mid)

		if err != nil {
			return 0, err
		}
		if failed {
			lo = mid
		} else {
			hi = mid
		}
	}

	if hi == cap {
		failed, result, err := executable(hi)
		if err != nil {
			return 0, err
		}
		if failed {
			if result != nil && result.Err != errors2.ErrOutOfGas {
				if len(result.Revert()) > 0 {
					return 0, newRevertError(result)
				}
				return 0, result.Err
			}
			return 0, fmt.Errorf("gas required exceeds allowance (%d)", cap)
		}
	}
	return hexutil.Uint64(hi), nil
}

func (api *EthereumAPI) GetBlockByHash(hash common.Hash, fullTx bool) (interface{}, error) {
	log.Debugf("eth_getBlockByHash hash %v, fullTx %v", hash, fullTx)
	block, err := bactor.GetBlockFromStore(oComm.Uint256(hash))
	if err != nil {
		if err == sCom.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	bloom, err := bactor.GetBloomData(block.Header.Height)
	if err != nil {
		return nil, err
	}
	return utils2.EthBlockFromOntology(block, fullTx, bloom), nil
}

func (api *EthereumAPI) GetBlockByNumber(blockNum types2.BlockNumber, fullTx bool) (interface{}, error) {
	log.Debugf("eth_getBlockByNumber blockNum %v, fullTx %v", blockNum, fullTx)
	height := uint32(blockNum)
	if blockNum.IsLatest() || blockNum.IsPending() {
		height = bactor.GetCurrentBlockHeight()
	}
	block, err := bactor.GetBlockByHeight(height)
	if err != nil {
		if err == sCom.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	bloom, err := bactor.GetBloomData(block.Header.Height)
	if err != nil {
		return nil, err
	}
	return utils2.EthBlockFromOntology(block, fullTx, bloom), nil
}

func (api *EthereumAPI) GetTransactionByHash(hash common.Hash) (*types2.Transaction, error) {
	log.Debugf("eth_getTransactionByHash hash %v", hash.Hex())
	height, tx, err := bactor.GetTxnWithHeightByTxHash(oComm.Uint256(hash))
	if err != nil {
		if err == common2.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if tx == nil {
		return nil, nil
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
	return utils2.OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx))
}

func (api *EthereumAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, idx hexutil.Uint) (*types2.Transaction, error) {
	log.Debugf("eth_getTransactionByBlockHashAndIndex hash %v, idx %v", hash.Hex(), idx.String())
	block, err := bactor.GetBlockFromStore(oComm.Uint256(hash))
	if err != nil {
		if err == sCom.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	header := block.Header
	blockHash := header.Hash()
	txs := block.Transactions
	if len(txs) >= int(idx) {
		return nil, nil
	}
	tx := txs[idx]
	return utils2.OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx))
}

func (api *EthereumAPI) GetTransactionByBlockNumberAndIndex(blockNum types2.BlockNumber, idx hexutil.Uint) (*types2.Transaction, error) {
	log.Debugf("eth_getTransactionByBlockNumberAndIndex hash %d, idx %v", blockNum, idx.String())
	height := uint32(blockNum)
	if blockNum.IsLatest() || blockNum.IsPending() {
		height = bactor.GetCurrentBlockHeight()
	}
	block, err := bactor.GetBlockByHeight(height)
	if err != nil {
		if err == sCom.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	header := block.Header
	blockHash := header.Hash()
	txs := block.Transactions
	if len(txs) >= int(idx) {
		return nil, nil
	}
	tx := txs[idx]
	return utils2.OntTxToEthTx(*tx, common.Hash(blockHash), uint64(header.Height), uint64(idx))
}

func (api *EthereumAPI) GetTransactionReceipt(hash common.Hash) (interface{}, error) {
	log.Debugf("eth_getTransactionReceipt hash %d", hash.Hex())
	notify, err := bactor.GetEventNotifyByTxHash(utils2.EthToOntHash(hash))
	if err != nil {
		if err == sCom.ErrNotFound {
			return nil, nil
		}
		return nil, nil
	}
	if notify == nil {
		return nil, nil
	}
	return generateRecipient(notify)
}

func generateRecipient(notify *event.ExecuteNotify) (interface{}, error) {
	logs, hash, tx, height, err := generateLog(notify)
	if err != nil {
		return nil, err
	}
	eip155Tx, err := tx.GetEIP155Tx()
	if err != nil {
		return nil, err
	}
	signer := types.NewEIP155Signer(big.NewInt(int64(utils2.GetChainId())))
	from, err := signer.Sender(eip155Tx)
	if err != nil {
		return nil, err
	}
	receipt := map[string]interface{}{
		// Consensus fields: These fields are defined by the Yellow Paper
		"status":            hexutil.Uint(notify.State),
		"cumulativeGasUsed": hexutil.Uint64(notify.GasConsumed),
		"logsBloom":         types.BytesToBloom(types.LogsBloom(logs)),
		"logs":              logs,

		// Implementation fields: These fields are added by geth when processing a transaction.
		// They are stored in the chain database.
		"transactionHash": common.Hash(notify.TxHash),
		"contractAddress": nil,
		"gasUsed":         hexutil.Uint64(notify.GasStepUsed),

		// Inclusion information: These fields provide information about the inclusion of the
		// transaction corresponding to this receipt.
		"blockHash":        hash,
		"blockNumber":      hexutil.Uint64(height),
		"transactionIndex": hexutil.Uint64(notify.TxIndex),

		// sender and receiver (contract or EOA) addresses
		"from": from,
		"to":   eip155Tx.To(),
	}
	if logs == nil {
		receipt["logs"] = [][]*types.Log{}
	}
	ethContractAddr := common.Address(notify.CreatedContract)
	// If the ContractAddress is 20 0x0 bytes, assume it is not a contract creation
	if ethContractAddr != (common.Address{}) {
		receipt["contractAddress"] = ethContractAddr
	}
	return receipt, nil
}

func (api *EthereumAPI) PendingTransactions() ([]*types2.Transaction, error) {
	log.Debug("eth_pendingTransactions")
	pendingTxs := api.txpool.PendingEIPTransactions()
	var rpcTxs []*types2.Transaction
	for _, v2 := range pendingTxs {
		tx, err := utils2.NewTransaction(v2, v2.Hash(), common.Hash{}, 0, 0)
		if err != nil {
			return nil, nil
		}
		rpcTxs = append(rpcTxs, tx)
	}
	return rpcTxs, nil
}

func (api *EthereumAPI) PendingTransactionsByHash(target common.Hash) (*types2.Transaction, error) {
	log.Debugf("eth_pendingTransactionsByHash target %v", target.Hex())
	ethTx := api.txpool.PendingTransactionsByHash(target)
	if ethTx == nil {
		return nil, nil
	}
	return utils2.NewTransaction(ethTx, ethTx.Hash(), common.Hash{}, 0, 0)
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
