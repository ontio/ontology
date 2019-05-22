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
	ShardID common.ShardID
	MsgHash common.Uint256
	SigData [][]byte
}

func (this *CrossShardMsgHash) Serialization(sink *common.ZeroCopySink) {
	sink.WriteShardID(this.ShardID)
	sink.WriteBytes(this.MsgHash[:])
	sink.WriteVarUint(uint64(len(this.SigData)))
	for _, sig := range this.SigData {
		sink.WriteVarBytes(sig)
	}
}

func (this *CrossShardMsgHash) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	var err error
	this.ShardID, err = source.NextShardID()
	if err != nil {
		return err
	}
	this.MsgHash, eof = source.NextHash()
	n, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	for j := 0; j < int(n); j++ {
		sig, _, irregular, eof := source.NextVarBytes()
		if eof {
			return io.ErrUnexpectedEOF
		}
		if irregular {
			return common.ErrIrregularData
		}
		this.SigData = append(this.SigData, sig)
	}
	return nil
}

type CrossShardMsg struct {
	FromShardID          common.ShardID
	MsgHeight            uint32
	SignMsgHeight        uint32
	PreCrossShardMsgHash common.Uint256
	CrossShardMsgRoot    common.Uint256
	ShardMsg             []xshard_types.CommonShardMsg
	ShardMsgHashs        []*CrossShardMsgHash
}

func (this *CrossShardMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteShardID(this.FromShardID)
	sink.WriteUint32(this.MsgHeight)
	sink.WriteUint32(this.SignMsgHeight)
	sink.WriteBytes(this.PreCrossShardMsgHash[:])
	sink.WriteBytes(this.CrossShardMsgRoot[:])
	xshard_types.EncodeShardCommonMsgs(sink, this.ShardMsg)
	sink.WriteVarUint(uint64(len(this.ShardMsgHashs)))
	for _, shardMsgHash := range this.ShardMsgHashs {
		shardMsgHash.Serialization(sink)
	}
}

func (this *CrossShardMsg) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	var err error
	this.FromShardID, err = source.NextShardID()
	if err != nil {
		return err
	}
	this.MsgHeight, eof = source.NextUint32()
	this.SignMsgHeight, eof = source.NextUint32()
	this.PreCrossShardMsgHash, eof = source.NextHash()
	this.CrossShardMsgRoot, eof = source.NextHash()

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
	if eof {
		return io.ErrUnexpectedEOF
	}
	m, _, irregular, eof := source.NextVarUint()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}
	for i := 0; i < int(m); i++ {
		crossShardMsgHash := &CrossShardMsgHash{}
		err = crossShardMsgHash.Deserialization(source)
		if err != nil {
			return err
		}
		this.ShardMsgHashs = append(this.ShardMsgHashs, crossShardMsgHash)
	}
	return nil
}

type CrossShardTxInfos struct {
	ShardMsg *CrossShardMsg `json:"shard_msg"`
	Tx       *Transaction
}

func (this *CrossShardTxInfos) Serialization(sink *common.ZeroCopySink) error {
	this.ShardMsg.Serialization(sink)
	this.Tx.Serialization(sink)
	return nil
}

func (this *CrossShardTxInfos) Deserialization(source *common.ZeroCopySource) error {
	var err error
	shardMsg := &CrossShardMsg{}
	err = shardMsg.Deserialization(source)
	if err != nil {
		return err
	}
	this.ShardMsg = shardMsg
	transaction := new(Transaction)
	err = transaction.Deserialization(source)
	if err != nil {
		return err
	}
	this.Tx = transaction
	return nil
}
