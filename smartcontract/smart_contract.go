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
package smartcontract

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/store"
	ctypes "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/neovm"
	"github.com/ontio/ontology/smartcontract/storage"
	vm "github.com/ontio/ontology/vm/neovm"
)

const (
	MAX_EXECUTE_ENGINE = 1024
)

// SmartContract describe smart contract execute engine
type SmartContract struct {
	Contexts      []*context.Context    // all execute smart contract context
	CacheDB       *storage.CacheDB      // state cache
	ShardTxState  *xshard_state.TxState // shardid is tx hash
	Store         store.LedgerStore     // ledger store
	Config        *Config
	Notifications []*event.NotifyEventInfo // all execute smart contract event notify info
	GasTable      map[string]uint64
	LockedAddress map[common.Address]struct{}
	Gas           uint64
	ExecStep      int
	PreExec       bool

	IsShardCall bool
	FromShard   common.ShardID
}

// Config describe smart contract need parameters configuration
type Config struct {
	ShardID      common.ShardID
	Time         uint32              // current block timestamp
	Height       uint32              // current block height
	ParentHeight uint32              // parent block height
	BlockHash    common.Uint256      // current block hash
	Tx           *ctypes.Transaction // current transaction
}

// PushContext push current context to smart contract
func (this *SmartContract) PushContext(context *context.Context) {
	this.Contexts = append(this.Contexts, context)
}

// CurrentContext return smart contract current context
func (this *SmartContract) CurrentContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[len(this.Contexts)-1]
}

// CallingContext return smart contract caller context
func (this *SmartContract) CallingContext() *context.Context {
	if len(this.Contexts) < 2 {
		return nil
	}
	return this.Contexts[len(this.Contexts)-2]
}

// EntryContext return smart contract entry entrance context
func (this *SmartContract) EntryContext() *context.Context {
	if len(this.Contexts) < 1 {
		return nil
	}
	return this.Contexts[0]
}

// PopContext pop smart contract current context
func (this *SmartContract) PopContext() {
	if len(this.Contexts) > 1 {
		this.Contexts = this.Contexts[:len(this.Contexts)-1]
	}
}

// PushNotifications push smart contract event info
func (this *SmartContract) PushNotifications(notifications []*event.NotifyEventInfo) {
	this.Notifications = append(this.Notifications, notifications...)
}

func (this *SmartContract) CheckExecStep() bool {
	if this.ExecStep >= neovm.VM_STEP_LIMIT {
		return false
	}
	this.ExecStep += 1
	return true
}

func (this *SmartContract) CheckUseGas(gas uint64) bool {
	if this.Gas < gas {
		return false
	}
	this.Gas -= gas
	return true
}

func (this *SmartContract) GetParentHeight() uint32 {
	return this.Config.ParentHeight
}

func (this *SmartContract) checkContexts() bool {
	if len(this.Contexts) > MAX_EXECUTE_ENGINE {
		return false
	}
	return true
}

// Execute is smart contract execute manager
// According different vm type to launch different service
func (this *SmartContract) NewExecuteEngine(code []byte) (context.Engine, error) {
	if !this.checkContexts() {
		return nil, fmt.Errorf("%s", "engine over max limit!")
	}
	service := &neovm.NeoVmService{
		Store:         this.Store,
		CacheDB:       this.CacheDB,
		ContextRef:    this,
		GasTable:      this.GasTable,
		LockedAddress: this.LockedAddress,
		Code:          code,
		Tx:            this.Config.Tx,
		ShardID:       this.Config.ShardID,
		ShardTxState:  this.ShardTxState,
		Time:          this.Config.Time,
		Height:        this.Config.Height,
		BlockHash:     this.Config.BlockHash,
		Engine:        vm.NewExecutionEngine(),
		PreExec:       this.PreExec,
	}
	return service, nil
}

func (this *SmartContract) NewNativeService() (*native.NativeService, error) {
	if !this.checkContexts() {
		return nil, fmt.Errorf("%s", "engine over max limit!")
	}
	service := &native.NativeService{
		CacheDB:       this.CacheDB,
		ContextRef:    this,
		ShardTxState:  this.ShardTxState,
		LockedAddress: this.LockedAddress,
		Tx:            this.Config.Tx,
		ShardID:       this.Config.ShardID,
		Time:          this.Config.Time,
		Height:        this.Config.Height,
		BlockHash:     this.Config.BlockHash,
		ServiceMap:    make(map[string]native.Handler),
	}
	return service, nil
}

// CheckWitness check whether authorization correct
// If address is wallet address, check whether in the signature addressed list
// Else check whether address is calling contract address
// Param address: wallet address or contract address
func (this *SmartContract) CheckWitness(address common.Address) bool {
	if this.checkAccountAddress(address) || this.checkContractAddress(address) {
		return true
	}
	return false
}

func (this *SmartContract) checkAccountAddress(address common.Address) bool {
	addresses, err := this.Config.Tx.GetSignatureAddresses()
	if err != nil {
		log.Errorf("get signature address error:%v", err)
		return false
	}
	for _, v := range addresses {
		if v == address {
			return true
		}
	}
	return false
}

func (this *SmartContract) IsPreExec() bool {
	return this.PreExec
}

func (this *SmartContract) GetRemainGas() uint64 {
	return this.Gas
}

func (this *SmartContract) CheckCallShard(fromShard common.ShardID) bool {
	return this.IsShardCall && this.FromShard.ToUint64() == fromShard.ToUint64()
}

func (this *SmartContract) GetMetaData(contract common.Address) (*payload.MetaDataCode, bool, error) {
	meta, err := this.CacheDB.GetMetaData(contract)
	if err != nil {
		return nil, true, fmt.Errorf("GetMetaData %s", err)
	}
	if this.Config.ShardID.IsRootShard() {
		return nil, true, nil
	}
	if meta == nil {
		meta, err = this.Store.GetParentMetaData(this.Config.ParentHeight, contract)
		if err != nil {
			return nil, false, fmt.Errorf("GetMetaData from parent, err: %s", err)
		}
		return meta, false, nil
	} else {
		return meta, true, nil
	}
}

func (this *SmartContract) checkContractAddress(address common.Address) bool {
	if this.CallingContext() != nil && this.CallingContext().ContractAddress == address {
		return true
	}
	return false
}

func (this *SmartContract) NotifyRemoteShard(target common.ShardID, cont common.Address, fee uint64, method string,
	args []byte) {
	if this.IsPreExec() {
		return
	}
	if this.Gas < fee {
		log.Errorf("NotifyRemoteShard: gas not enough")
		return
	}
	if err := this.checkInvoke(cont); err != nil {
		log.Errorf("NotifyRemoteShard: failed, err: %s", err)
		return
	}
	txState := this.ShardTxState
	msg := &xshard_types.XShardNotify{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: this.Config.ShardID,
			TargetShardID: target,
			SourceTxHash:  this.Config.Tx.Hash(),
			ShardTxID:     txState.TxID,
		},
		NotifyID: txState.NumNotifies,
		Contract: cont,
		Payer:    this.Config.Tx.Payer,
		Fee:      fee,
		Method:   method,
		Args:     args,
	}
	txState.NumNotifies += 1
	// todo: clean shardnotifies when replay transaction
	txState.ShardNotifies = append(txState.ShardNotifies, msg)
}

func (this *SmartContract) InvokeRemoteShard(target common.ShardID, cont common.Address, method string,
	args []byte) ([]byte, error) {
	if this.IsPreExec() {
		return native.BYTE_TRUE, nil
	}
	if err := this.checkInvoke(cont); err != nil {
		return native.BYTE_FALSE, fmt.Errorf("InvokeRemoteShard: failed, err: %s", err)
	}
	if this.Config.ShardID.IsRootShard() || target.IsRootShard() {
		return native.BYTE_FALSE, fmt.Errorf("InvokeRemoteShard: root cannot participate in")
	}
	txState := this.ShardTxState
	reqIdx := txState.NextReqID
	if reqIdx >= xshard_state.MaxRemoteReqPerTx {
		return native.BYTE_FALSE, xshard_state.ErrTooMuchRemoteReq
	}
	if this.Gas < neovm.MIN_TRANSACTION_GAS {
		return native.BYTE_FALSE, fmt.Errorf("InvokeRemoteShard: gas less than min gas")
	}
	msg := &xshard_types.XShardTxReq{
		ShardMsgHeader: xshard_types.ShardMsgHeader{
			SourceShardID: this.Config.ShardID,
			TargetShardID: target,
			SourceTxHash:  this.Config.Tx.Hash(),
			ShardTxID:     txState.TxID,
		},
		IdxInTx:  uint64(reqIdx),
		Payer:    this.Config.Tx.Payer,
		Fee:      this.Gas, // use all remain gas to invoke remote shard
		GasPrice: this.Config.Tx.GasPrice,
		Contract: cont,
		Method:   method,
		Args:     args,
	}
	txState.NextReqID += 1

	if reqIdx < uint32(len(txState.OutReqResp)) {
		if xshard_types.IsXShardMsgEqual(msg, txState.OutReqResp[reqIdx].Req) == false {
			return native.BYTE_FALSE, xshard_state.ErrMismatchedRequest
		}
		rspMsg := txState.OutReqResp[reqIdx].Resp
		var resultErr error = nil
		if rspMsg.Error {
			resultErr = fmt.Errorf("InvokeRemoteShard: got error response")
		}
		if !this.CheckUseGas(rspMsg.FeeUsed) { // charge whole remain gas
			resultErr = fmt.Errorf("InvokeRemoteShard: gas not enough")
			this.CheckUseGas(this.Gas)
		}
		return rspMsg.Result, resultErr
	}

	if len(txState.TxPayload) == 0 {
		txPayload := bytes.NewBuffer(nil)
		if err := this.Config.Tx.Serialize(txPayload); err != nil {
			return native.BYTE_FALSE, fmt.Errorf("InvokeRemoteShard: failed to get tx payload: %s", err)
		}
		txState.TxPayload = txPayload.Bytes()
	}

	// no response found in tx-statedb, send request
	if err := txState.AddTxShard(target); err != nil {
		return native.BYTE_FALSE, fmt.Errorf("InvokeRemoteShard: failed to add shard: %s", err)
	}

	txState.PendingOutReq = msg
	txState.ExecState = xshard_state.ExecYielded

	return native.BYTE_FALSE, xshard_state.ErrYield
}

func (this *SmartContract) checkInvoke(destContract common.Address) error {
	// if caller is native contract, no need to check destContract
	if _, ok := native.Contracts[this.CurrentContext().ContractAddress]; ok {
		return nil
	} else if _, ok := native.Contracts[destContract]; ok {
		return fmt.Errorf("checkInvoke: neovm contract cannot x-shard call native contract")
	}
	caller := this.CurrentContext().ContractAddress
	meta, _, err := this.GetMetaData(caller)
	if err != nil {
		return fmt.Errorf("checkInvoke: cannot get %s meta", caller.ToHexString())

	}
	if meta == nil {
		return fmt.Errorf("checkInvoke: self %s doesn't initialized meta", caller.ToHexString())
	}
	canInvoke := false
	for _, addr := range meta.InvokedContract {
		if addr == destContract {
			canInvoke = true
			break
		}
	}
	if !canInvoke {
		return fmt.Errorf("checkInvoke: contract %s unregister to self %s meta",
			destContract.ToHexString(), caller.ToHexString())
	}
	return nil
}
