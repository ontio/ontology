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
	"errors"
	"io"

	ser "github.com/ontio/ontology/common/serialization"
)

type BlockSignatures struct {
	msgData    ConsensusMessageData
	Signatures []SignaturesData
}

type SignaturesData struct {
	Signature []byte
	Index     uint16
}

func (self *BlockSignatures) Serialize(w io.Writer) error {
	self.msgData.Serialize(w)
	if err := ser.WriteVarUint(w, uint64(len(self.Signatures))); err != nil {
		return errors.New("[BlockSignatures] serialization failed")
	}

	for i := 0; i < len(self.Signatures); i++ {
		if err := ser.WriteVarBytes(w, self.Signatures[i].Signature); err != nil {
			return errors.New("[BlockSignatures] serialization sig failed")
		}
		if err := ser.WriteUint16(w, self.Signatures[i].Index); err != nil {
			return errors.New("[BlockSignatures] serialization sig index failed")
		}
	}

	return nil
}

func (self *BlockSignatures) Deserialize(r io.Reader) error {
	err := self.msgData.Deserialize(r)
	if err != nil {
		return err
	}

	length, err := ser.ReadVarUint(r, 0)
	if err != nil {
		return err
	}

	for i := uint64(0); i < length; i++ {
		sig := SignaturesData{}

		sig.Signature, err = ser.ReadVarBytes(r)
		if err != nil {
			return err
		}
		sig.Index, err = ser.ReadUint16(r)
		if err != nil {
			return err
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
