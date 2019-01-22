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
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

type Block struct {
	Header       *Header
	Events       map[uint64][]shardstates.ShardMgmtEvent
	Transactions []*Transaction
}

func (b *Block) Serialization(sink *common.ZeroCopySink) error {
	err := b.Header.Serialization(sink)
	if err != nil {
		return err
	}

	sink.WriteUint32(uint32(len(b.Events)))
	for shardID, evts := range b.Events {
		if err := zcpSerializeShardEvents(sink, shardID, evts); err != nil {
			return err
		}
	}

	sink.WriteUint32(uint32(len(b.Transactions)))
	for _, transaction := range b.Transactions {
		err := transaction.Serialization(sink)
		if err != nil {
			return err
		}
	}
	return nil
}

// if no error, ownership of param raw is transfered to Transaction
func BlockFromRawBytes(raw []byte) (*Block, error) {
	source := common.NewZeroCopySource(raw)
	block := &Block{}
	err := block.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (self *Block) Deserialization(source *common.ZeroCopySource) error {
	if self.Header == nil {
		self.Header = new(Header)
	}
	err := self.Header.Deserialization(source)
	if err != nil {
		return err
	}

	nEvts, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	events, err := zcpDeserializeShardEvents(source, nEvts)
	if err != nil {
		return err
	}
	self.Events = events

	length, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}

	var hashes []common.Uint256
	mask := make(map[common.Uint256]bool)
	for i := uint32(0); i < length; i++ {
		transaction := new(Transaction)
		// note currently all transaction in the block shared the same source
		err := transaction.Deserialization(source)
		if err != nil {
			return err
		}
		txhash := transaction.Hash()
		if mask[txhash] {
			return errors.New("duplicated transaction in block")
		}
		mask[txhash] = true
		hashes = append(hashes, txhash)
		self.Transactions = append(self.Transactions, transaction)
	}

	root := common.ComputeMerkleRoot(hashes)
	if self.Header.TransactionsRoot != root {
		return errors.New("mismatched transaction root")
	}

	return nil
}

func (b *Block) ToArray() []byte {
	sink := common.NewZeroCopySink(nil)
	b.Serialization(sink)
	return sink.Bytes()
}

func (b *Block) Hash() common.Uint256 {
	return b.Header.Hash()
}

func (b *Block) Type() common.InventoryType {
	return common.BLOCK
}

func (b *Block) RebuildMerkleRoot() {
	txs := b.Transactions
	hashes := make([]common.Uint256, 0, len(txs))
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash())
	}
	hash := common.ComputeMerkleRoot(hashes)
	b.Header.TransactionsRoot = hash
}

func zcpSerializeShardEvents(sink *common.ZeroCopySink, shardID uint64, evts []shardstates.ShardMgmtEvent) error {
	if len(evts) == 0 {
		return nil
	}

	sink.WriteUint64(shardID)
	sink.WriteUint32(uint32(len(evts)))
	for _, evt := range evts {
		buf := new(bytes.Buffer)
		if err := evt.Serialize(buf); err != nil {
			return err
		}
		sink.WriteUint32(evt.GetType())
		sink.WriteVarBytes(buf.Bytes())
	}

	return nil
}

func zcpDeserializeShardEvents(source *common.ZeroCopySource, shardCnt uint32) (map[uint64][]shardstates.ShardMgmtEvent, error) {
	shardEvts := make(map[uint64][]shardstates.ShardMgmtEvent)
	for i := uint32(0); i < shardCnt; i++ {
		shardID, eof := source.NextUint64()
		if eof {
			return nil, io.ErrUnexpectedEOF
		}

		evtCnt, eof := source.NextUint32()
		if eof {
			return nil, io.ErrUnexpectedEOF
		}

		evts := make([]shardstates.ShardMgmtEvent, 0)
		for j := uint32(0); j < evtCnt; j++ {
			evtType, eof := source.NextUint32()
			if eof {
				return nil, io.ErrUnexpectedEOF
			}
			evtBytes, _, _, eof := source.NextVarBytes()
			if eof {
				return nil, io.ErrUnexpectedEOF
			}
			evt, err := shardstates.DecodeShardEvent(evtType, evtBytes)
			if err != nil {
				return nil, fmt.Errorf("decode event: %s", err)
			}
			evts = append(evts, evt)
		}
		shardEvts[shardID] = evts
	}

	return shardEvts, nil
}
