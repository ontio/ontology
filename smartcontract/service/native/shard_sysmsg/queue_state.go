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
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type ToShardsInBlock struct {
	Shards []common.ShardID
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
	this.Shards = make([]common.ShardID, shardNum)
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
	this.Shards = make([]common.ShardID, num)
	for i := uint64(0); i < num; i++ {
		shard, err := utils.DeserializationShardId(source)
		if err != nil {
			return err
		}
		this.Shards[i] = shard
	}
	return nil
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
