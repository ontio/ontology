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
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/chainmgr/xshard_state"
	sComm "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type ToShardsInBlock struct {
	Shards []types.ShardID
}

func (this *ToShardsInBlock) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.Shards))); err != nil {
		return fmt.Errorf("srialize: write shards len failed, err: %s", err)
	}
	for i, shard := range this.Shards {
		if err := utils.SerializeShardId(w, shard); err != nil {
			return fmt.Errorf("serialize: write shard id failed, index %d, err: %s", i, err)
		}
	}
	return nil
}

func (this *ToShardsInBlock) Deserialize(r io.Reader) error {
	var err error = nil
	shardNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read shards len failed, err: %s", err)
	}
	this.Shards = make([]types.ShardID, shardNum)
	for i := uint64(0); i < shardNum; i++ {
		shard, err := utils.DeserializeShardId(r)
		if err != nil {
			return fmt.Errorf("deserialize: read shard failed, index %d, err: %s", i, err)
		}
		this.Shards[i] = shard
	}
	return nil
}
func (this *ToShardsInBlock) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.Shards)))
	for _, shardId := range this.Shards {
		utils.SerializationShardId(sink, shardId)
	}
}

func (this *ToShardsInBlock) Deserialization(source *common.ZeroCopySource) error {
	num, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Shards = make([]types.ShardID, num)
	for i := uint64(0); i < num; i++ {
		shard, err := utils.DeserializationShardId(source)
		if err != nil {
			return err
		}
		this.Shards[i] = shard
	}
	return nil
}

func addToShardsInBlock(ctx *native.NativeService, toShard types.ShardID) error {
	toShards, err := getToShardsInBlock(ctx, ctx.Height)
	if err != nil {
		return err
	}

	for _, s := range toShards {
		if s == toShard {
			// already in
			return nil
		}
	}
	toShards = append(toShards, toShard)

	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := utils.GetUint32Bytes(ctx.Height)

	toShardsInBlk := &ToShardsInBlock{
		Shards: toShards,
	}
	sink := common.NewZeroCopySink(0)
	toShardsInBlk.Serialization(sink)

	log.Debugf("put ToShards: height: %d, shards: %v", ctx.Height, toShards)

	key := utils.ConcatKey(contract, []byte(KEY_SHARDS_IN_BLOCK), blockNumBytes)
	xshard_state.PutKV(key, sink.Bytes())
	return nil
}

func getToShardsInBlock(ctx *native.NativeService, blockHeight uint32) ([]types.ShardID, error) {
	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := utils.GetUint32Bytes(blockHeight)

	key := utils.ConcatKey(contract, []byte(KEY_SHARDS_IN_BLOCK), blockNumBytes)
	toShardsBytes, err := xshard_state.GetKVStorageItem(key)
	if err != nil && err != xshard_state.ErrNotFound {
		return nil, fmt.Errorf("get toShards: %s", err)
	}
	if toShardsBytes == nil {
		// not found
		return nil, nil
	}

	req := &ToShardsInBlock{}
	if err := req.Deserialization(common.NewZeroCopySource(toShardsBytes)); err != nil {
		return nil, fmt.Errorf("deserialize toShards: %s: %s", err, string(toShardsBytes))
	}

	return req.Shards, nil
}

type ReqsInBlock struct {
	Reqs [][]byte
}

func (this *ReqsInBlock) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.Reqs))); err != nil {
		return fmt.Errorf("srialize: write reqs len failed, err: %s", err)
	}
	for i, req := range this.Reqs {
		if err := serialization.WriteVarBytes(w, req); err != nil {
			return fmt.Errorf("serialize: write req failed, index %d, err: %s", i, err)
		}
	}
	return nil
}

func (this *ReqsInBlock) Deserialize(r io.Reader) error {
	var err error = nil
	reqNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("deserialize: read reqs len failed, err: %s", err)
	}
	this.Reqs = make([][]byte, reqNum)
	for i := uint64(0); i < reqNum; i++ {
		req, err := serialization.ReadVarBytes(r)
		if err != nil {
			return fmt.Errorf("deserialize: read req failed, index %d, err: %s", i, err)
		}
		this.Reqs[i] = req
	}
	return nil
}

func (this *ReqsInBlock) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.Reqs)))
	for _, req := range this.Reqs {
		sink.WriteVarBytes(req)
	}
}

func (this *ReqsInBlock) Deserialization(source *common.ZeroCopySource) error {
	num, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Reqs = make([][]byte, num)
	for i := uint64(0); i < num; i++ {
		data, _, irregular, eof := source.NextVarBytes()
		if irregular {
			return common.ErrIrregularData
		}
		if eof {
			return io.ErrUnexpectedEOF
		}
		this.Reqs[i] = data
	}
	return nil
}

func addReqsInBlock(ctx *native.NativeService, req *xshard_state.CommonShardMsg) error {
	reqs, err := getReqsInBlock(ctx, ctx.Height, req.GetTargetShardID())
	if err != nil && err != sComm.ErrNotFound {
		return err
	}
	buf := common.SerializeToBytes(req)
	reqs = append(reqs, buf)

	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := utils.GetUint32Bytes(ctx.Height)
	shardIDBytes := utils.GetUint64Bytes(req.GetTargetShardID().ToUint64())

	reqInBlk := &ReqsInBlock{
		Reqs: reqs,
	}
	sink := common.NewZeroCopySink(0)
	reqInBlk.Serialization(sink)

	key := utils.ConcatKey(contract, []byte(KEY_REQS_IN_BLOCK), blockNumBytes, shardIDBytes)
	xshard_state.PutKV(key, sink.Bytes())
	return nil
}

func getReqsInBlock(ctx *native.NativeService, blockHeight uint32, shardID types.ShardID) ([][]byte, error) {
	contract := ctx.ContextRef.CurrentContext().ContractAddress
	blockNumBytes := utils.GetUint32Bytes(blockHeight)
	shardIDBytes := utils.GetUint64Bytes(shardID.ToUint64())

	key := utils.ConcatKey(contract, []byte(KEY_REQS_IN_BLOCK), blockNumBytes, shardIDBytes)
	reqBytes, err := xshard_state.GetKVStorageItem(key)
	if err != nil && err != xshard_state.ErrNotFound {
		return nil, fmt.Errorf("get reqs in block: %s", err)
	}
	if reqBytes == nil {
		// not found
		return nil, nil
	}

	req := &ReqsInBlock{}
	if err := req.Deserialization(common.NewZeroCopySource(reqBytes)); err != nil {
		return nil, fmt.Errorf("deserialize reqsInBlock: %s", err)
	}

	return req.Reqs, nil
}
