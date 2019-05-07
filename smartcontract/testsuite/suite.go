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
package testsuite

import (
	"crypto/rand"
	"fmt"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/storage"
	"io"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	utils2 "github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native"
	shardsysmsg "github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func RandomAddress() common.Address {
	var addr common.Address
	_, _ = rand.Read(addr[:])

	return addr
}

func InstallNativeContract(addr common.Address, actions map[string]native.Handler) {
	contract := func(native *native.NativeService) {
		for name, fun := range actions {
			native.Register(name, fun)
		}
	}
	native.Contracts[addr] = contract
}

func executeTransaction(tx *types.Transaction, cache *storage.CacheDB) (*xshard_state.TxState,
	interface{}, error) {
	config := &smartcontract.Config{
		ShardID: common.NewShardIDUnchecked(tx.ShardID),
		Time:    uint32(time.Now().Unix()),
		Tx:      tx,
	}

	txHash := tx.Hash()
	txState := xshard_state.CreateTxState(xshard_state.ShardTxID(string(txHash[:])))

	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:           config,
			Store:            nil,
			MainShardTxState: txState,
			CacheDB:          cache,
			Gas:              100000000000000,
			PreExec:          true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()

		if err != nil {
			//if err == shardsysmsg.ErrYield {
			//	return txState, err
			//}
			// todo: handle error check
			if txState.PendingReq != nil {
				return txState, nil, shardsysmsg.ErrYield
			}
			return nil, nil, err
		}

		return txState, res, nil
	}

	panic("unimplemented")
}

func resumeTx(shardTxID xshard_state.ShardTxID, rspMsg *xshard_types.XShardTxRsp) (*xshard_state.TxState, interface{}, error) {
	txState := xshard_state.CreateTxState(shardTxID).Clone()

	if txState.PendingReq == nil || txState.PendingReq.IdxInTx != rspMsg.IdxInTx {
		// todo: system error or remote shard error
		return nil, nil, fmt.Errorf("invalid response id: %d", rspMsg.IdxInTx)
	}

	txState.OutReqResp = append(txState.OutReqResp, &xshard_state.XShardTxReqResp{Req: txState.PendingReq, Resp: rspMsg})
	txState.PendingReq = nil

	txPayload := txState.TxPayload
	if txPayload == nil {
		return nil, nil, fmt.Errorf("failed to get tx payload")
	}

	// FIXME: invoke neo contract
	// re-execute tx
	txState.NextReqID = 0

	tx, err := types.TransactionFromRawBytes(txPayload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to re-init original tx: %s", err)
	}

	config := &smartcontract.Config{
		ShardID: common.NewShardIDUnchecked(tx.ShardID),
		Time:    uint32(time.Now().Unix()),
		Tx:      tx,
	}

	overlay := NewOverlayDB()
	cache := storage.NewCacheDB(overlay)
	if tx.TxType == types.Invoke {
		invoke := tx.Payload.(*payload.InvokeCode)

		sc := smartcontract.SmartContract{
			Config:           config,
			Store:            nil,
			MainShardTxState: txState,
			CacheDB:          cache,
			Gas:              100000000000000,
			PreExec:          true,
		}

		//start the smart contract executive function
		engine, _ := sc.NewExecuteEngine(invoke.Code)
		res, err := engine.Invoke()

		if err != nil {
			//if err == shardsysmsg.ErrYield {
			//	return txState, err
			//}
			// todo: handle error check
			if txState.PendingReq != nil {
				return txState, nil, shardsysmsg.ErrYield
			}
			return nil, nil, err
		}

		// xshard transaction has completed
		txState.WriteSet = cache.GetCache()

		return txState, res, nil
	}

	panic("unimplemented")
}

func processXShardRsp(shardTxID xshard_state.ShardTxID, rspMsg *xshard_types.XShardTxRsp) (*xshard_state.TxState, error) {
	txState, _, err := resumeTx(shardTxID, rspMsg)

	if err != nil {
		if err == shardsysmsg.ErrYield {
			return txState, err
		}
		// Txn failed, abort all transactions
		//todo abort tx
		//if _, err2 := abortTx(ctx, txState, tx); err2 != nil {
		//	return fmt.Errorf("rwset verify %s, abort tx %v, err: %s", err, tx, err2)
		//}
		//return resultErr
		return txState, err
	}

	return txState, err
}

var ShardA = common.NewShardIDUnchecked(1)
var ShardB = common.NewShardIDUnchecked(2)

func RemoteNotifyPing(native *native.NativeService) ([]byte, error) {
	sink := common.NewZeroCopySink(10)
	sink.WriteString(fmt.Sprintf("hello from shard: %d", native.ShardID.ToUint64()))

	params := shardsysmsg.NotifyReqParam{
		ToShard: ShardB,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	shardsysmsg.RemoteNotifyApi(native, &params)

	return utils.BYTE_TRUE, nil
}

func RemoteInvokeAddAndInc(native *native.NativeService) ([]byte, error) {
	sink := common.NewZeroCopySink(10)
	sink.WriteUint64(2)
	sink.WriteUint64(3)

	params := &shardsysmsg.NotifyReqParam{
		ToShard: ShardB,
		Method:  "handlePing",
		Args:    sink.Bytes(),
	}

	sum, err := shardsysmsg.RemoteInvokeApi(native, params)
	if err != nil {
		return nil, err
	}
	source := common.NewZeroCopySource(sum)
	s, eof := source.NextUint64()
	if eof {
		return nil, io.ErrUnexpectedEOF
	}

	sink.Reset()
	sink.WriteUint64(s + 1)

	pushEvent(native, sink.Bytes())

	return sink.Bytes(), err
}

func pushEvent(native *native.NativeService, s interface{}) {
	event := new(event.NotifyEventInfo)
	event.ContractAddress = native.ContextRef.CurrentContext().ContractAddress
	event.States = s
	native.Notifications = append(native.Notifications, event)
}

func HandlePing(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_TRUE, nil
}

func BuildInvokeTx(contractAddress common.Address, method string,
	args []interface{}) *types.Transaction {
	invokCode, err := utils2.BuildNativeInvokeCode(contractAddress, 0, method, args)
	if err != nil {
		return nil
	}
	invokePayload := &payload.InvokeCode{
		Code: invokCode,
	}
	tx := &types.MutableTransaction{
		Version:  0,
		GasPrice: 0,
		GasLimit: 1000000000,
		TxType:   types.Invoke,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     make([]types.Sig, 0, 0),
	}
	res, err := tx.IntoImmutable()
	if err != nil {
		return nil
	}
	return res
}
