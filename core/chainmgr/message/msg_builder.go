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

package message

import (
	"fmt"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/xshard_types"
)

type CrossShardMsgHash struct {
	ShardID      common.ShardID
	ShardMsgHash common.Uint256
	SigData      [][]byte
}

type CrossShardMsg struct {
	FromShardID       common.ShardID
	MsgHeight         uint32
	SignMsgHeight     uint32
	CrossShardMsgRoot common.Uint256
	ShardMsg          []xshard_types.CommonShardMsg
	ShardMsgHash      []*CrossShardMsgHash
}

func (this *CrossShardMsg) Serialization(sink *common.ZeroCopySink) {
	sink.WriteShardID(this.FromShardID)
	sink.WriteUint32(this.MsgHeight)
	sink.WriteUint32(this.SignMsgHeight)
	sink.WriteBytes(this.CrossShardMsgRoot[:])
	xshard_types.EncodeShardCommonMsgs(sink, this.ShardMsg)
	sink.WriteVarUint(uint64(len(this.ShardMsgHash)))
	for _, shardMsgHash := range this.ShardMsgHash {
		sink.WriteShardID(shardMsgHash.ShardID)
		sink.WriteBytes(shardMsgHash.ShardMsgHash[:])
		sink.WriteVarUint(uint64(len(shardMsgHash.SigData)))
		for _, sig := range shardMsgHash.SigData {
			sink.WriteVarBytes(sig)
		}
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
		crossShardMsgHash.ShardID, err = source.NextShardID()
		if err != nil {
			return err
		}
		crossShardMsgHash.ShardMsgHash, eof = source.NextHash()
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
			crossShardMsgHash.SigData = append(crossShardMsgHash.SigData, sig)
		}
		this.ShardMsgHash = append(this.ShardMsgHash, crossShardMsgHash)
	}
	return nil
}

//
// NewCrossShardTxMsg: create cross-shard transaction, to remote ShardSysMsg contract
//  @payload: contains N sub-txns
//
//  One block can generated multiple cross-shard sub-txns, marshaled to [][]byte.
//  NewCrossShardTXMsg creates one cross-shard forwarding Tx, which contains all sub-txns.
//
func NewCrossShardTxMsg(account *account.Account, height uint32, toShardID common.ShardID, gasPrice, gasLimit uint64,
	msgs []xshard_types.CommonShardMsg) (*types.Transaction, error) {
	// build transaction
	shardCall := &payload.ShardCall{
		Msgs: msgs,
	}
	mutable := &types.MutableTransaction{
		ShardID:  toShardID.ToUint64(),
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Payer:    account.Address,
		TxType:   types.ShardCall,
		Nonce:    height, // use height as nonce
		Payload:  shardCall,
		Version:  common.VERSION_SUPPORT_SHARD,
		Sigs:     nil,
	}

	// add signatures
	txHash := mutable.Hash()
	sig, err := signature.Sign(account.SigScheme, account.PrivateKey, txHash.ToArray(), nil)
	if err != nil {
		return nil, fmt.Errorf("sign tx: %s", err)
	}
	sigData, err := signature.Serialize(sig)
	if err != nil {
		return nil, fmt.Errorf("serialize sig: %s", err)
	}
	mutable.Sigs = []types.Sig{
		{
			PubKeys: []keypair.PublicKey{account.PubKey()},
			M:       1,
			SigData: [][]byte{sigData},
		},
	}
	return mutable.IntoImmutable()
}
