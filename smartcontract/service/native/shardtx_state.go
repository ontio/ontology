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
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

const (
	TxExec = iota
	TxPrepared
	TxAbort
	TxCommit
)

var (
	MaxRemoteReqPerTx      = 8
	ErrNotFound            = errors.New("not found")
	ErrTooMuchRemoteReq    = errors.New("too much remote request")
	ErrMismatchedTxPayload = errors.New("Mismatched Tx Payload")
	ErrMismatchedRequest   = errors.New("Mismatched request")
	ErrMismatchedResponse  = errors.New("Mismatched response")
)

type TxStateInShard struct {
	State int
}

type TxState struct {
	Caller    common.Address
	State     int
	Shards    map[uint64]*TxStateInShard
	TxPayload []byte
	Reqs      map[int32]*shardstates.XShardTxReq
	Rsps      map[int32]*shardstates.XShardTxRsp
	Result    []byte
	ResultErr error
	DB        *storage.CacheDB
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

func AddTxShard(tx common.Uint256, shardID uint64) error {
	if state, present := shardTxStateTable.TxStates[tx]; present {
		if _, present := state.Shards[shardID]; !present {
			state.Shards[shardID] = &TxStateInShard{
				State: TxExec,
			}
		}
	}

	return ErrNotFound
}

func GetTxCommitState(tx common.Uint256) (map[uint64]*TxStateInShard, error) {
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
		Shards:     make(map[uint64]*TxStateInShard),
		Reqs:       make(map[int32]*shardstates.XShardTxReq),
		Rsps:       make(map[int32]*shardstates.XShardTxRsp),
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

	txState.Shards = make(map[uint64]*TxStateInShard)
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

func GetTxResponse(tx common.Uint256, txReq *shardstates.XShardTxReq) ([]byte, error) {
	txState, err := CreateTxState(tx)
	if err != nil {
		return nil, ErrNotFound
	}

	if rspMsg, present := txState.Rsps[txReq.IdxInTx]; present {
		return rspMsg.Result, errors.New(rspMsg.Error)
	}
	return nil, ErrNotFound
}

func PutTxResponse(tx common.Uint256, txRsp *shardstates.XShardTxRsp) error {
	txState, err := CreateTxState(tx)
	if err != nil {
		return errors.ErrNotFound
	}

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
	return nil, nil
}

func PutTxRequest(tx common.Uint256, txPayload []byte, req shardstates.XShardMsg) error {
	if req.Type() != shardstates.EVENT_SHARD_TXREQ {
		return fmt.Errorf("invalid type of txReq: %d", req.Type())
	}

	txReq, ok := req.(*shardstates.XShardTxReq)
	if !ok || txReq == nil {
		return fmt.Errorf("invalid txReq")
	}

	txState, err := CreateTxState(tx)
	if err != nil {
		return err
	}

	if reqMsg, present := txState.Reqs[txReq.IdxInTx]; present {
		if reqMsg.Contract == txReq.Contract &&
			reqMsg.Method == txReq.Method &&
			bytes.Compare(reqMsg.Args, txReq.Args) == 0 {
			return nil
		} else {
			return ErrMismatchedRequest
		}
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
	return nil
}

func GetNextReqIndex(tx common.Uint256) int32 {
	// TODO
	return 0
}

func UpdateTxState(tx common.Uint256, caller common.Address, dataDB *storage.CacheDB, result []byte) error {
	txState := shardTxStateTable.TxStates[tx]
	if txState == nil {
		txState = &TxState{
			Caller: caller,
			DB:     dataDB,
		}
	}
	shardTxStateTable.TxStates[tx] = txState
	return nil
}

func VerifyStates(ctx *NativeService, tx common.Uint256) error {
	// TODO
	return nil
}

func GetTxContracts(ctx *NativeService, tx common.Uint256) ([]common.Address, error) {
	// TODO
	return []common.Address{}, nil
}

func LockContract(ctx *NativeService, contract common.Address) error {
	// TODO
	return nil
}
