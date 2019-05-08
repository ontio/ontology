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

package dbft

import (
	"io"

	"github.com/ontio/ontology/common"
)

type BlockSignatures struct {
	msgData    ConsensusMessageData
	Signatures []SignaturesData
}

type SignaturesData struct {
	Signature []byte
	Index     uint16
}

func (self *BlockSignatures) Serialization(sink *common.ZeroCopySink) {
	self.msgData.Serialization(sink)
	sink.WriteVarUint(uint64(len(self.Signatures)))

	for _, sign := range self.Signatures {
		sink.WriteVarBytes(sign.Signature)
		sink.WriteUint16(sign.Index)
	}
}

func (self *BlockSignatures) Deserialization(source *common.ZeroCopySource) error {
	err := self.msgData.Deserialization(source)
	if err != nil {
		return err
	}

	length, _, irregular, eof := source.NextVarUint()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	for i := uint64(0); i < length; i++ {
		sig := SignaturesData{}

		sig.Signature, _, irregular, eof = source.NextVarBytes()
		if irregular {
			return common.ErrIrregularData
		}

		sig.Index, eof = source.NextUint16()
		if eof {
			return io.ErrUnexpectedEOF
		}

		self.Signatures = append(self.Signatures, sig)
	}

	return nil
}

func (self *BlockSignatures) Type() ConsensusMessageType {
	return self.ConsensusMessageData().Type
}

func (self *BlockSignatures) ViewNumber() byte {
	return self.msgData.ViewNumber
}

func (self *BlockSignatures) ConsensusMessageData() *ConsensusMessageData {
	return &(self.msgData)
}
