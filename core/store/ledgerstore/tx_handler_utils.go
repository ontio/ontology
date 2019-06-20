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
package ledgerstore

import (
	"bytes"
	"fmt"
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
	cutils "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
)

func lockKey(lockedKeys map[string]struct{}, key []byte) {
	if shouldLock(key) {
		lockedKeys[string(key)] = struct{}{}
	}
}

func shouldLock(key []byte) bool {
	keyLen := len(key)
	// key contains storage prefix
	if keyLen < common.ADDR_LEN+1 {
		return false
	}
	cmpKey := key[1 : common.ADDR_LEN+1]
	if bytes.Equal(cmpKey, utils.ShardAssetAddress[:]) || bytes.Equal(cmpKey, utils.OngContractAddress[:]) ||
		bytes.Equal(cmpKey, utils.ShardMgmtContractAddress[:]) {
		return true
	}
	return false
}

func calcGasByCodeLen(codeLen int, codeGas uint64) uint64 {
	return uint64(codeLen/neovm.PER_UNIT_CODE_LEN) * codeGas
}

func buildTx(originalPayer, contract common.Address, method string, args []interface{}, shardId, gasLimit uint64,
	nonce uint32) (*types.Transaction, error) {
	invokeCode := []byte{}
	var err error = nil
	if _, ok := native.Contracts[contract]; ok {
		invokeCode, err = cutils.BuildNativeInvokeCode(contract, 0, method, args)
	} else {
		invokeCode, err = cutils.BuildNeoVMInvokeCode(contract, []interface{}{method, args})
	}
	if err != nil {
		return nil, fmt.Errorf("buildTx: build invoke failed, err: %s", err)
	}
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	mutable := &types.MutableTransaction{
		Version:  common.CURR_TX_VERSION,
		GasPrice: neovm.GAS_PRICE,
		ShardID:  shardId,
		GasLimit: gasLimit,
		TxType:   types.Invoke,
		Nonce:    nonce,
		Payer:    originalPayer,
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return nil, fmt.Errorf("buildTx: build tx failed, err: %s", err)
	}
	tx.SignedAddr = append(tx.SignedAddr, originalPayer)
	return tx, nil
}

func genNativeTransferCode(from, to common.Address, value uint64) []byte {
	transfer := ont.Transfers{States: []ont.State{{From: from, To: to, Value: value}}}
	return common.SerializeToBytes(&transfer)
}

// check whether payer ong balance sufficient
func isBalanceSufficient(payer common.Address, cache *storage.CacheDB, config *smartcontract.Config, store store.LedgerStore, gas uint64) (uint64, error) {
	balance, err := getBalanceFromNative(config, cache, store, payer)
	if err != nil {
		return 0, err
	}
	if balance < gas {
		return 0, fmt.Errorf("payer gas insufficient, need %d , only have %d", gas, balance)
	}
	return balance, nil
}

func getBalanceFromNative(config *smartcontract.Config, cache *storage.CacheDB, store store.LedgerStore, address common.Address) (uint64, error) {
	bf := new(bytes.Buffer)
	if err := utils.WriteAddress(bf, address); err != nil {
		return 0, err
	}
	sc := smartcontract.SmartContract{
		Config:  config,
		CacheDB: cache,
		Store:   store,
		Gas:     math.MaxUint64,
	}

	service, _ := sc.NewNativeService()
	result, err := service.NativeCall(utils.OngContractAddress, ont.BALANCEOF_NAME, bf.Bytes())
	if err != nil {
		return 0, err
	}
	return common.BigIntFromNeoBytes(result.([]byte)).Uint64(), nil
}

func chargeHandleRespFee(store store.LedgerStore, cache *storage.CacheDB, header *types.Header, shardId common.ShardID,
	notify *event.TransactionNotify, subTx *types.Transaction, txState *xshard_state.TxState, gasConsumed uint64,
	isExecFailed bool) bool {
	if subTx.GasPrice == 0 {
		return true
	}
	if isExecFailed {
		// exec failed, only charge x-shard invoke fee, because:
		// 1. x-shard invoke fee is debt of dest shard
		// 2. tx failed fee is charged by original invoke tx
		return chargeWholeRespFee(txState, subTx, header, cache, store, shardId, notify)
	} else {
		// exec complete, charge all tx fee except original invoke tx failed fee
		if gasConsumed > neovm.MIN_TRANSACTION_GAS {
			config := &smartcontract.Config{
				ShardID:   shardId,
				Time:      header.Timestamp,
				Height:    header.Height,
				Tx:        subTx,
				BlockHash: header.Hash(),
			}
			fee := gasConsumed - neovm.MIN_TRANSACTION_GAS
			if notifies, err := chargeCostGas(subTx.Payer, fee, config, cache, store, shardId); err == nil {
				notify.ContractEvent.Notify = append(notify.ContractEvent.Notify, notifies...)
				return true
			}
		}
	}
	return false
}

func chargeWholeRespFee(txState *xshard_state.TxState, tx *types.Transaction, header *types.Header,
	cache *storage.CacheDB, store store.LedgerStore, shardId common.ShardID, notify *event.TransactionNotify) bool {
	config := &smartcontract.Config{
		ShardID:   shardId,
		Time:      header.Timestamp,
		Height:    header.Height,
		Tx:        tx,
		BlockHash: header.Hash(),
	}
	wholeXShardInvokeFee := uint64(0)
	for _, resp := range txState.OutReqResp {
		wholeXShardInvokeFee += resp.Resp.FeeUsed
	}
	if notifies, err := chargeCostGas(tx.Payer, wholeXShardInvokeFee, config, cache, store, shardId); err != nil {
		return false
	} else {
		notify.ContractEvent.Notify = append(notify.ContractEvent.Notify, notifies...)
		return true
	}
}
