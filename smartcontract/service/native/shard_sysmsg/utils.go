/*
 * Copyright (C) 2019 The ontology Authors
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

package shardsysmsg

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/core/types"
	"io"

	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type ToShardsInBlock struct {
	Shards []types.ShardID `json:"shards"`
}

func (this *ToShardsInBlock) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ToShardsInBlock) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

func addToShardsInBlock(ctx *native.NativeService, toShard types.ShardID) error {
	toShards, err := getToShardsInBlock(ctx, ctx.Height)
	if err != nil {
		return err
	}
	if toShards == nil {
		toShards = []types.ShardID{toShard}
	} else {
		for _, s := range toShards {
			if s == toShard {
				// already in
				return nil
			}
		}
		toShards = append(toShards, toShard)
	}

	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := shardutil.GetUint32Bytes(ctx.Height)

	toShardsInBlk := &ToShardsInBlock{
		Shards: toShards,
	}
	buf := new(bytes.Buffer)
	if err := toShardsInBlk.Serialize(buf); err != nil {
		return fmt.Errorf("serialize to-shards in block: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_SHARDS_IN_BLOCK), blockNumBytes)
	ctx.CacheDB.Put(key, cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func getToShardsInBlock(ctx *native.NativeService, blockHeight uint32) ([]types.ShardID, error) {
	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := shardutil.GetUint32Bytes(blockHeight)

	key := utils.ConcatKey(contract, []byte(KEY_SHARDS_IN_BLOCK), blockNumBytes)
	toShardsBytes, err := ctx.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("get toShards: %s", err)
	}
	if toShardsBytes == nil {
		// not found
		return nil, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(toShardsBytes)
	if err != nil {
		return nil, fmt.Errorf("get toShards from bytes: %s", err)
	}

	req := &ToShardsInBlock{}
	if err := req.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("deserialize toShards: %s", err)
	}

	return req.Shards, nil
}

type ReqsInBlock struct {
	Reqs [][]byte `json:"reqs"`
}

func (this *ReqsInBlock) Serialize(w io.Writer) error {
	return shardutil.SerJson(w, this)
}

func (this *ReqsInBlock) Deserialize(r io.Reader) error {
	return shardutil.DesJson(r, this)
}

func addReqsInBlock(ctx *native.NativeService, req *shardstates.CommonShardMsg) error {
	reqs, err := getReqsInBlock(ctx, ctx.Height, req.GetTargetShardID())
	if err != nil {
		return err
	}
	reqBytes := new(bytes.Buffer)
	if err := req.Serialize(reqBytes); err != nil {
		return err
	}
	if reqs == nil {
		reqs = [][]byte{reqBytes.Bytes()}
	}

	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := shardutil.GetUint32Bytes(ctx.Height)
	shardIDBytes, err := shardutil.GetUint64Bytes(req.GetTargetShardID().ToUint64())
	if err != nil {
		return fmt.Errorf("serialzie toshard: %s", err)
	}

	reqInBlk := &ReqsInBlock{
		Reqs: reqs,
	}
	buf := new(bytes.Buffer)
	if err := reqInBlk.Serialize(buf); err != nil {
		return fmt.Errorf("serialize shardmgmt global state: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_REQS_IN_BLOCK), blockNumBytes, shardIDBytes)
	ctx.CacheDB.Put(key, cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func getReqsInBlock(ctx *native.NativeService, blockHeight uint32, shardID types.ShardID) ([][]byte, error) {
	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := shardutil.GetUint32Bytes(blockHeight)
	shardIDBytes, err := shardutil.GetUint64Bytes(shardID.ToUint64())
	if err != nil {
		return nil, fmt.Errorf("serialize toShard: %s", err)
	}

	key := utils.ConcatKey(contract, []byte(KEY_REQS_IN_BLOCK), blockNumBytes, shardIDBytes)
	reqBytes, err := ctx.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("get reqs in block: %s", err)
	}
	if reqBytes == nil {
		// not found
		return nil, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(reqBytes)
	if err != nil {
		return nil, fmt.Errorf("get reqs from bytes: %s", err)
	}

	req := &ReqsInBlock{}
	if err := req.Deserialize(bytes.NewBuffer(value)); err != nil {
		return nil, fmt.Errorf("deserialize reqsInBlock: %s", err)
	}

	return req.Reqs, nil
}


func txCommitReady(txState map[uint64]*native.TxStateInShard) bool {
	for _, state := range txState {
		if state.State < native.TxPrepared {
			return false
		}
	}
	return true
}

