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
package types

import (
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/xshard_types"
)

type CrossShardMsgHash struct {
	ShardMsgHashs []common.Uint256
	SigData       map[uint32][]byte
}

func (this *CrossShardMsgHash) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.ShardMsgHashs)))
	for _, shardMsgHash := range this.ShardMsgHashs {
		sink.WriteBytes(shardMsgHash[:])
	}
	sink.WriteUint32(uint32(len(this.SigData)))
	IndexIds := make([]uint32, 0, len(this.SigData))
	for id := range this.SigData {
		IndexIds = append(IndexIds, id)
	}
	common.SortUint32s(IndexIds)
	for _, index := range IndexIds {
		evts := this.SigData[index]
		zcpSerializeShardSigData(sink, index, evts)
	}
}

func (this *CrossShardMsgHash) Deserialization(source *common.ZeroCopySource) error {
	m, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	for i := 0; i < int(m); i++ {
		msghash, eof := source.NextHash()
		if eof {
			return io.ErrUnexpectedEOF
		}
		this.ShardMsgHashs = append(this.ShardMsgHashs, msghash)
	}
	n, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	sigData, err := zcpDeserializeShardSigData(source, n)
	if err != nil {
		return err
	}
	this.SigData = sigData
	return nil
}

func zcpSerializeShardSigData(sink *common.ZeroCopySink, index uint32, sigData []byte) {
	if len(sigData) == 0 {
		return
	}
	sink.WriteUint32(index)
	sink.WriteVarBytes(sigData)
}

func zcpDeserializeShardSigData(source *common.ZeroCopySource, sigDataCnt uint32) (map[uint32][]byte, error) {
	sigData := make(map[uint32][]byte)
	for i := uint32(0); i < sigDataCnt; i++ {
		index, eof := source.NextUint32()
		if eof {
			return nil, io.ErrUnexpectedEOF
		}
		sig, _, irregular, eof := source.NextVarBytes()
		if eof {
			return nil, io.ErrUnexpectedEOF
		}
		if irregular {
			return nil, common.ErrIrregularData
		}
		sigData[index] = sig
	}
	return sigData, nil
}

type CrossShardMsg struct {
	CrossShardMsgInfo *CrossShardMsgInfo
	ShardMsg          []xshard_types.CommonShardMsg
}

func (this *CrossShardMsg) Serialization(sink *common.ZeroCopySink) {
	this.CrossShardMsgInfo.Serialization(sink)
	xshard_types.EncodeShardCommonMsgs(sink, this.ShardMsg)
}

func (this *CrossShardMsg) Deserialization(source *common.ZeroCopySource) error {
	if this.CrossShardMsgInfo == nil {
		this.CrossShardMsgInfo = new(CrossShardMsgInfo)
	}
	err := this.CrossShardMsgInfo.Deserialization(source)
	if err != nil {
		return err
	}
	len, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	reqs := make([]xshard_types.CommonShardMsg, 0)
	for i := uint32(0); i < len; i++ {
		req, err := xshard_types.DecodeShardCommonMsg(source)
		if err != nil {
			return fmt.Errorf("failed to unmarshal req-tx: %s", err)
		}
		reqs = append(reqs, req)
	}
	this.ShardMsg = reqs
	return nil
}

type CrossShardMsgInfo struct {
	SignMsgHeight        uint32
	PreCrossShardMsgHash common.Uint256
	Index                uint32
	ShardMsgInfo         *CrossShardMsgHash
}

func (this *CrossShardMsgInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.SignMsgHeight)
	sink.WriteBytes(this.PreCrossShardMsgHash[:])
	sink.WriteUint32(this.Index)
	this.ShardMsgInfo.Serialization(sink)

}

func (this *CrossShardMsgInfo) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.SignMsgHeight, eof = source.NextUint32()
	this.PreCrossShardMsgHash, eof = source.NextHash()
	this.Index, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	err := this.ShardMsgInfo.Deserialization(source)
	if err != nil {
		return err
	}
	return nil
}

type CrossShardTxInfos struct {
	ShardMsg *CrossShardMsgInfo `json:"shard_msg"`
	Tx       *Transaction
}

func (this *CrossShardTxInfos) Serialization(sink *common.ZeroCopySink) error {
	if this.ShardMsg != nil {
		this.ShardMsg.Serialization(sink)
	} else {
		return io.ErrUnexpectedEOF
	}
	if this.Tx != nil {
		this.Tx.Serialization(sink)
	} else {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (this *CrossShardTxInfos) Deserialization(source *common.ZeroCopySource) error {
	var err error
	if this.ShardMsg == nil {
		this.ShardMsg = new(CrossShardMsgInfo)
	}
	err = this.ShardMsg.Deserialization(source)
	if err != nil {
		return err
	}
	tx := &Transaction{}
	err = tx.Deserialization(source)
	if err != nil {
		return err
	}
	this.Tx = tx
	return nil
}
