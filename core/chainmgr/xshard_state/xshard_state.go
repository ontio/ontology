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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
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
	ErrMismatchedResponse  = errors.New("nismatched response")
)

type TxState struct {
	Caller    common.Address
	State     int
	Shards    map[uint64]int
	TxPayload []byte
	NextReqID int32
	Reqs      map[int32]*shardstates.XShardTxReq
	Rsps      map[int32]*shardstates.XShardTxRsp
	Result    []byte
	ResultErr error
	WriteSet  *overlaydb.MemDB
	Notify    *event.ExecuteNotify
}

type ShardTxStateMap struct {
	TxStates map[common.Uint256]*TxState
}

var shardTxStateTable = ShardTxStateMap{
	TxStates: make(map[common.Uint256]*TxState),
}

func GetTxShards(tx common.Uint256) ([]types.ShardID, error) {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		shards := make([]types.ShardID, 0, len(state.Shards))
		for s := range state.Shards {
			id, _ := types.NewShardID(s)
			shards = append(shards, id)
		}
		return shards, nil
	}

	return nil, ErrNotFound
}

func AddTxShard(tx common.Uint256, shardID types.ShardID) error {
	txState, err := CreateTxState(tx)
	if err != nil {
		return err
	}
	id := shardID.ToUint64()
	if state, present := txState.Shards[id]; !present {
		txState.Shards[id] = TxExec
	} else if state != TxExec {
		return ErrInvalidTxState
	}

	return nil
}

func IsTxExecutionPaused(tx common.Uint256) (bool, error) {
	txState, err := GetTxState(tx)
	if err != nil {
		return false, err
	}

	return txState.State != TxExec, nil
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

func GetTxCommitState(tx common.Uint256) (map[uint64]int, error) {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		return state.Shards, nil
	}
	return nil, ErrNotFound
}

// CreateTxState
// If txState available, return it.  Otherwise, Create txState.
func CreateTxState(tx common.Uint256) (*TxState, error) {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		return state, nil
	}
	state := &TxState{
		State:  TxExec,
		Shards: make(map[uint64]int),
		Reqs:   make(map[int32]*shardstates.XShardTxReq),
		Rsps:   make(map[int32]*shardstates.XShardTxRsp),
	}
	shardTxStateTable.TxStates[tx] = state
	return state, nil
}

func GetTxState(tx common.Uint256) (*TxState, error) {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		return state, nil
	}
	return nil, ErrNotFound
}

func GetNextReqIndex(tx common.Uint256) int32 {
	txState, err := CreateTxState(tx)
	if err != nil {
		return -1
	}
	if txState.NextReqID >= MaxRemoteReqPerTx {
		return -1
	}

	return txState.NextReqID
}

func SetNextReqIndex(tx common.Uint256, nextId int32) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}

	txState.NextReqID = nextId
	return nil
}

func SetTxPrepared(tx common.Uint256) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}

	if txState.State != TxExec {
		return fmt.Errorf("invalid state to prepared: %s", txState.State)
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

	txState.Shards = make(map[uint64]int)
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

func GetTxResponse(tx common.Uint256, txReq *shardstates.XShardTxReq) (*shardstates.XShardTxRsp, error) {
	txState, err := CreateTxState(tx)
	if err != nil {
		return nil, err
	}

	if rspMsg, present := txState.Rsps[txReq.IdxInTx]; present {
		return rspMsg, nil
	}
	return nil, nil
}

func PutTxResponse(tx common.Uint256, txRsp *shardstates.XShardTxRsp) error {
	txState, err := CreateTxState(tx)
	if err != nil {
		return ErrNotFound
	}

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

func GetTxRequests(tx common.Uint256) ([]*shardstates.XShardTxReq, error) {
	txState, err := GetTxState(tx)
	if err != nil {
		return nil, err
	}
	reqs := make([]*shardstates.XShardTxReq, 0)
	for _, req := range txState.Reqs {
		reqs = append(reqs, req)
	}
	return reqs, nil
}

func ValidateTxRequest(tx common.Uint256, req *shardstates.XShardTxReq) error {
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

func PutTxRequest(tx common.Uint256, txPayload []byte, req shardstates.XShardMsg) error {
	if req.Type() != shardstates.EVENT_SHARD_TXREQ {
		return fmt.Errorf("invalid type of txReq: %d", req.Type())
	}

	txReq, ok := req.(*shardstates.XShardTxReq)
	if !ok || txReq == nil {
		return fmt.Errorf("invalid txReq")
	}

	if err := ValidateTxRequest(tx, txReq); err != nil {
		return fmt.Errorf("validate tx request idx %d: %s", txReq.IdxInTx, err)
	}

	txState, err := CreateTxState(tx)
	if err != nil {
		return err
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
	SetNextReqIndex(tx, txReq.IdxInTx+1)
	return nil
}

func UpdateTxResult(tx common.Uint256, dataDB *storage.CacheDB) error {
	txState, err := GetTxState(tx)
	if err != nil {
		return err
	}
	txState.WriteSet = dataDB.GetCache()
	return nil
}

func VerifyStates(tx common.Uint256) error {
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
