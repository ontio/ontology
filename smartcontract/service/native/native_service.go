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

package native

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/states"
	sstates "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/smartcontract/storage"
)

type (
	Handler         func(native *NativeService) ([]byte, error)
	RegisterService func(native *NativeService)
)

var (
	Contracts = make(map[common.Address]RegisterService)
)

var BYTE_FALSE = []byte{0}
var BYTE_TRUE = []byte{1}

// Native service struct
// Invoke a native smart contract, new a native service
type NativeService struct {
	CacheDB          *storage.CacheDB
	ServiceMap       map[string]Handler
	Notifications    []*event.NotifyEventInfo
	InvokeParam      sstates.ContractInvokeParam
	Input            []byte
	Tx               *types.Transaction
	ShardID          common.ShardID
	Height           uint32
	Time             uint32
	BlockHash        common.Uint256
	MainShardTxState *xshard_state.TxState
	SubShardTxState  map[xshard_state.ShardTxID]xshard_state.ShardTxInfo
	ContextRef       context.ContextRef
}

func (this *NativeService) Register(methodName string, handler Handler) {
	this.ServiceMap[methodName] = handler
}

func (this *NativeService) Invoke() (interface{}, error) {
	contract := this.InvokeParam
	services, ok := Contracts[contract.Address]
	if !ok {
		return false, fmt.Errorf("Native contract address %x haven't been registered.", contract.Address)
	}
	services(this)
	service, ok := this.ServiceMap[contract.Method]
	if !ok {
		return false, fmt.Errorf("Native contract %x doesn't support this function %s.",
			contract.Address, contract.Method)
	}
	args := this.Input
	this.Input = contract.Args
	this.ContextRef.PushContext(&context.Context{ContractAddress: contract.Address})
	notifications := this.Notifications
	this.Notifications = []*event.NotifyEventInfo{}
	result, err := service(this)
	if err != nil {
		return result, fmt.Errorf("[Invoke] Native serivce function execute error: %v", err.Error())
	}
	this.ContextRef.PopContext()
	this.ContextRef.PushNotifications(this.Notifications)
	this.Notifications = notifications
	this.Input = args
	return result, nil
}

func (this *NativeService) NativeCall(address common.Address, method string, args []byte) (interface{}, error) {
	c := states.ContractInvokeParam{
		Address: address,
		Method:  method,
		Args:    args,
	}
	this.InvokeParam = c
	return this.Invoke()
}

// runtime api
func (ctx *NativeService) NotifyRemoteShard(target common.ShardID, cont common.Address, method string, args []byte) {
	if ctx.ContextRef.IsPreExec() {
		return
	}
	txState := ctx.MainShardTxState
	// send with minimal gas fee
	msg := &xshard_types.XShardNotify{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: ctx.ShardID,
			TargetShardID: target,
			SourceTxHash:  ctx.Tx.Hash(),
		},
		NotifyID: txState.NumNotifies,
		Contract: cont,
		Payer:    ctx.Tx.Payer,
		GasPrice: ctx.Tx.GasPrice,
		Fee:      ctx.ContextRef.GetRemainGas(), // TODO: fee should be defined by caller
		Method:   method,
		Args:     args,
	}
	txState.NumNotifies += 1
	// todo: clean shardnotifies when replay transaction
	txState.ShardNotifies = append(txState.ShardNotifies, msg)
}

// runtime api
func (ctx *NativeService) InvokeRemoteShard(target common.ShardID, cont common.Address,
	method string, args []byte) ([]byte, error) {
	if ctx.ContextRef.IsPreExec() {
		return BYTE_TRUE, nil
	}
	txState := ctx.MainShardTxState
	reqIdx := txState.NextReqID
	if reqIdx >= xshard_state.MaxRemoteReqPerTx {
		return BYTE_FALSE, xshard_state.ErrTooMuchRemoteReq
	}
	// TODO: open this to check remain gas enough
	//if ctx.ContextRef.GetRemainGas() < neovm.MIN_TRANSACTION_GAS {
	//	return BYTE_FALSE, fmt.Errorf("remote invoke gas less than min gas")
	//}
	msg := &xshard_types.XShardTxReq{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: ctx.ShardID,
			TargetShardID: target,
			SourceTxHash:  ctx.Tx.Hash(),
		},
		IdxInTx:  uint64(reqIdx),
		Payer:    ctx.Tx.Payer,
		GasPrice: ctx.Tx.GasPrice,
		Fee:      ctx.ContextRef.GetRemainGas(), // use all remain gas to invoke remote shard
		Contract: cont,
		Method:   method,
		Args:     args,
	}
	txState.NextReqID += 1

	if reqIdx < uint32(len(txState.OutReqResp)) {
		if xshard_types.IsXShardMsgEqual(msg, txState.OutReqResp[reqIdx].Req) == false {
			return BYTE_FALSE, xshard_state.ErrMismatchedRequest
		}
		rspMsg := txState.OutReqResp[reqIdx].Resp
		var resultErr error = nil
		if rspMsg.Error {
			resultErr = errors.New("remote invoke got error response")
		}
		if !ctx.ContextRef.CheckUseGas(rspMsg.FeeUsed) { // charge whole remain gas
			resultErr = errors.New("remote invoke gas not enough")
			ctx.ContextRef.CheckUseGas(ctx.ContextRef.GetRemainGas())
		}
		return rspMsg.Result, resultErr
	}

	if len(txState.TxPayload) == 0 {
		txPayload := bytes.NewBuffer(nil)
		if err := ctx.Tx.Serialize(txPayload); err != nil {
			return BYTE_FALSE, fmt.Errorf("remote invoke, failed to get tx payload: %s", err)
		}
		txState.TxPayload = txPayload.Bytes()
	}

	// no response found in tx-statedb, send request
	if err := txState.AddTxShard(target); err != nil {
		return BYTE_FALSE, fmt.Errorf("remote invoke, failed to add shard: %s", err)
	}

	txState.PendingReq = msg
	txState.ExecState = xshard_state.ExecYielded

	return BYTE_FALSE, xshard_state.ErrYield
}
