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

package xshard_state

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/storage"
)

const (
	TxUnknownState = iota
	TxExec
	TxWait
	TxPrepared
	TxAbort
	TxCommit
)

var (
	MaxRemoteReqPerTx      = int32(8)
	ErrNotFound            = errors.New("not found")
	ErrInvalidTxState      = errors.New("invalid transaction state")
	ErrTooMuchRemoteReq    = errors.New("too much remote request")
	ErrInvalidRemoteRsp    = errors.New("invalid remotes response")
	ErrMismatchedTxPayload = errors.New("mismatched Tx Payload")
	ErrMismatchedRequest   = errors.New("mismatched request")
	ErrMismatchedResponse  = errors.New("mismatched response")
)

//
// TxState
// stores intermediate states of one cross-shard transaction
//
// * State: current execution state : exec -> (wait) -> prepare -> prepared -> commit
// * Shards: store shards which participant with the transaction
// * TxPayload: serialized transaction
// * NextReqID: index of next transactional remote request
// * Reqs: sent remote requests in this transaction
// * Rsps: received remote responses in this transaction
// * Result: execution result of the transaction
// * ResultErr: execution error of the transaction
// * WriteSet:
// * Notify:
//
type TxState struct {
	State     int
	Shards    map[types.ShardID]int
	TxPayload []byte
	NextReqID int32
	Reqs      map[uint64]*XShardTxReq
	Rsps      map[uint64]*XShardTxRsp
	Result    []byte
	ResultErr error
	WriteSet  *overlaydb.MemDB
	Notify    *event.ExecuteNotify
}

//
// ShardTxStateMap
// stores all intermediate states of cross-shard transactions
//
type ShardTxStateMap struct {
	TxStates map[common.Uint256]*TxState
}

var shardTxStateTable = ShardTxStateMap{
	TxStates: make(map[common.Uint256]*TxState),
}

func (self *TxState) GetTxShards() []types.ShardID {
	shards := make([]types.ShardID, 0, len(self.Shards))
	for id := range self.Shards {
		shards = append(shards, id)
	}
	return shards
}

//
// GetTxShards
// get shards which participant with the procession of transaction
//
func GetTxShards(tx common.Uint256) ([]types.ShardID, error) {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		shards := make([]types.ShardID, 0, len(state.Shards))
		for id := range state.Shards {
			shards = append(shards, id)
		}
		return shards, nil
	}

	return nil, ErrNotFound
}

func (self *TxState) AddTxShard(id types.ShardID) error {
	if state, present := self.Shards[id]; !present {
		self.Shards[id] = TxExec
	} else if state != TxExec {
		return ErrInvalidTxState
	}

	return nil
}

//
// AddTxShard
// add participated shard to txState
//
func AddTxShard(tx common.Uint256, id types.ShardID) error {
	txState := CreateTxState(tx)
	if state, present := txState.Shards[id]; !present {
		txState.Shards[id] = TxExec
	} else if state != TxExec {
		return ErrInvalidTxState
	}

	return nil
}

func (self *TxState) IsTxExecutionPaused(tx common.Uint256) bool {
	return self.State != TxExec
}

func IsTxExecutionPaused(tx common.Uint256) (bool, error) {
	txState, err := GetTxState(tx)
	if err != nil {
		return false, err
	}

	return txState.State != TxExec, nil
}

func (txState *TxState) SetTxExecutionPaused() {
	switch txState.State {
	case TxExec:
		txState.State = TxWait
	}
}

func SetTxExecutionPaused(tx common.Uint256) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}
	switch txState.State {
	case TxExec:
		txState.State = TxWait
	}
	return nil
}

func (txState *TxState) SetTxExecutionContinued() {
	switch txState.State {
	case TxWait:
		txState.State = TxExec
	}
}

func SetTxExecutionContinued(tx common.Uint256) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}
	switch txState.State {
	case TxWait:
		txState.State = TxExec
	}
	return nil
}

func (self *TxState) SetShardPrepared(shardId types.ShardID) error {
	if _, ok := self.Shards[shardId]; !ok {
		return fmt.Errorf("invalid shard ID %d, in tx commit", shardId)
	}
	self.Shards[shardId] = TxPrepared
	return nil
}

func (self *TxState) IsTxCommitReady() bool {
	if self.State != TxPrepared {
		return false
	}
	for _, state := range self.Shards {
		if state != TxPrepared {
			return false
		}
	}
	return true
}

func (self *TxState) Clone() *TxState {
	txs := &TxState{
		State:     self.State,
		Shards:    make(map[types.ShardID]int),
		TxPayload: self.TxPayload,
		NextReqID: self.NextReqID,
		Reqs:      make(map[uint64]*XShardTxReq),
		Rsps:      make(map[uint64]*XShardTxRsp),
		Result:    make([]byte, len(self.Result)),
		ResultErr: self.ResultErr,
		WriteSet:  nil,
		Notify:    self.Notify,
	}

	for k, v := range self.Shards {
		txs.Shards[k] = v
	}
	// todo: need deep clone?
	for k, v := range self.Reqs {
		txs.Reqs[k] = v
	}
	for k, v := range self.Rsps {
		txs.Rsps[k] = v
	}
	//todo: need clone?
	txs.WriteSet = self.WriteSet

	return txs
}

func GetTxCommitState(tx common.Uint256) (map[types.ShardID]int, error) {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		return state.Shards, nil
	}
	return nil, ErrNotFound
}

// CreateTxState
// If txState available, return it.  Otherwise, Create txState.
func CreateTxState(tx common.Uint256) *TxState {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		return state
	}
	state := &TxState{
		State:  TxExec,
		Shards: make(map[types.ShardID]int),
		Reqs:   make(map[uint64]*XShardTxReq),
		Rsps:   make(map[uint64]*XShardTxRsp),
	}
	shardTxStateTable.TxStates[tx] = state
	return state
}

func PutTxState(tx common.Uint256, state *TxState) {
	shardTxStateTable.TxStates[tx] = state
}

func GetTxState(tx common.Uint256) (*TxState, error) {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		return state, nil
	}
	return nil, ErrNotFound
}

func GetNextReqIndex(tx common.Uint256) int32 {
	txState := CreateTxState(tx)
	if txState.NextReqID >= MaxRemoteReqPerTx {
		return -1
	}

	return txState.NextReqID
}

func SetNextReqIndex(tx common.Uint256, nextId uint64) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}
	if nextId > math.MaxInt32 {
		return fmt.Errorf("SetNextReqIndex: next id %d is too large", nextId)
	}
	txState.NextReqID = int32(nextId)
	return nil
}

func (self *TxState) SetTxPrepared() error {
	if self.State != TxExec {
		return fmt.Errorf("invalid state to prepared: %s", self.State)
	}

	self.State = TxPrepared
	return nil
}

func SetTxPrepared(tx common.Uint256) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}

	if txState.State != TxExec {
		return fmt.Errorf("invalid state to prepared: %d", txState.State)
	}

	txState.State = TxPrepared
	return nil
}

func SetTxCommitted(tx common.Uint256, isCommit bool) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}

	if isCommit && txState.State != TxPrepared {
		return fmt.Errorf("invalid state to commit: %d", txState.State)
	}

	if isCommit {
		txState.State = TxCommit
	} else {
		txState.State = TxAbort
	}

	txState.Shards = make(map[types.ShardID]int)
	return nil
}

func SetTxResult(tx common.Uint256, result []byte, resultErr error) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}

	txState.Result = result
	txState.ResultErr = resultErr
	return nil
}

func (self *TxState) GetResponse(reqIndex uint64) *XShardTxRsp {
	return self.Rsps[reqIndex]
}

//
// GetTxResponse
// get remote response of the request, if existed.
// return nil if not existed
//
func GetTxResponse(tx common.Uint256, txReq *XShardTxReq) *XShardTxRsp {
	txState := CreateTxState(tx)

	if rspMsg, present := txState.Rsps[txReq.IdxInTx]; present {
		return rspMsg
	}
	return nil
}

func (txState *TxState) PutTxResponse(txRsp *XShardTxRsp) error {
	// check if corresponding request existed
	if _, present := txState.Reqs[txRsp.IdxInTx]; !present {
		return ErrInvalidRemoteRsp
	}
	// check if duplicated response
	if rspMsg, present := txState.Rsps[txRsp.IdxInTx]; present {
		if bytes.Compare(rspMsg.Result, txRsp.Result) == 0 &&
			rspMsg.Error == txRsp.Error {
			return nil
		}
		return ErrMismatchedResponse
	}

	txState.Rsps[txRsp.IdxInTx] = txRsp
	return nil
}

//
// PutTxResponse
// add remote response to txState.
// if not matched with previous response, return ErrMismatchResponse
//
func PutTxResponse(tx common.Uint256, txRsp *XShardTxRsp) error {
	txState := CreateTxState(tx)

	// check if corresponding request existed
	if _, present := txState.Reqs[txRsp.IdxInTx]; !present {
		return ErrInvalidRemoteRsp
	}
	// check if duplicated response
	if rspMsg, present := txState.Rsps[txRsp.IdxInTx]; present {
		if bytes.Compare(rspMsg.Result, txRsp.Result) == 0 &&
			rspMsg.Error == txRsp.Error {
			return nil
		}
		return ErrMismatchedResponse
	}

	txState.Rsps[txRsp.IdxInTx] = txRsp
	return nil
}

func GetTxPayload(tx common.Uint256) ([]byte, error) {
	txState, err := GetTxState(tx)
	if err != nil {
		return nil, err
	}

	if txState.TxPayload == nil {
		return nil, ErrNotFound
	}
	return txState.TxPayload, nil
}

func GetTxRequests(tx common.Uint256) ([]*XShardTxReq, error) {
	txState, err := GetTxState(tx)
	if err != nil {
		return nil, err
	}
	reqs := make([]*XShardTxReq, 0)
	for _, req := range txState.Reqs {
		reqs = append(reqs, req)
	}
	return reqs, nil
}

//
// ValidateTxRequest
// check if the remote request is consistent with previous request which has same Index
//
func ValidateTxRequest(tx common.Uint256, req *XShardTxReq) error {
	txState, err := GetTxState(tx)
	if err == ErrNotFound {
		return nil
	}

	if reqMsg, present := txState.Reqs[req.IdxInTx]; present {
		if reqMsg.Contract == req.Contract &&
			reqMsg.Method == req.Method &&
			bytes.Compare(reqMsg.Args, req.Args) == 0 {
			return nil
		} else {
			return ErrMismatchedRequest
		}
	}

	return nil
}

func (txState *TxState) ValidateTxRequest(req *XShardTxReq) error {
	if reqMsg, present := txState.Reqs[req.IdxInTx]; present {
		if reqMsg.Contract == req.Contract &&
			reqMsg.Method == req.Method &&
			bytes.Compare(reqMsg.Args, req.Args) == 0 {
			return nil
		} else {
			return ErrMismatchedRequest
		}
	}

	return nil
}

func (txState *TxState) PutTxRequest(txPayload []byte, txReq *XShardTxReq) error {
	if err := txState.ValidateTxRequest(txReq); err != nil {
		return fmt.Errorf("validate tx request idx %d: %s", txReq.IdxInTx, err)
	}

	if txPayload != nil {
		if txState.TxPayload != nil {
			if bytes.Compare(txState.TxPayload, txPayload) != 0 {
				return ErrMismatchedTxPayload
			}
		} else {
			txState.TxPayload = txPayload
		}
	}
	txState.Reqs[txReq.IdxInTx] = txReq
	txState.NextReqID = int32(txReq.IdxInTx + 1)
	return nil
}

//
// PutTxRequest
// add remote request to txState
//	1. check if remote request is valid
//  2. add serialized tx to txState
//  3. update next request index
//
func PutTxRequest(tx common.Uint256, txPayload []byte, req XShardMsg) error {
	if req.Type() != EVENT_SHARD_TXREQ {
		return fmt.Errorf("invalid type of txReq: %d", req.Type())
	}

	txReq, ok := req.(*XShardTxReq)
	if !ok || txReq == nil {
		return fmt.Errorf("invalid txReq")
	}

	if err := ValidateTxRequest(tx, txReq); err != nil {
		return fmt.Errorf("validate tx request idx %d: %s", txReq.IdxInTx, err)
	}

	txState := CreateTxState(tx)

	if txPayload != nil {
		if txState.TxPayload != nil {
			if bytes.Compare(txState.TxPayload, txPayload) != 0 {
				return ErrMismatchedTxPayload
			}
		} else {
			txState.TxPayload = txPayload
		}
	}
	txState.Reqs[txReq.IdxInTx] = txReq
	SetNextReqIndex(tx, txReq.IdxInTx+1)
	return nil
}

//
// UpdateTxResult
// save writeset of the transaction to txState
//
func UpdateTxResult(tx common.Uint256, dataDB *storage.CacheDB) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}
	txState.WriteSet = dataDB.GetCache()
	return nil
}

func (self *TxState) VerifyStates() error {
	// TODO
	return nil
}

func GetTxContracts(tx common.Uint256) ([]common.Address, error) {
	// TODO
	return []common.Address{}, nil
}

func LockContract(contract common.Address) error {
	// TODO: lock contract if it does not support concurrency (shard-sysmsg contract support concurrency)
	return nil
}

func UnlockContract(contract common.Address) error {
	return nil
}
