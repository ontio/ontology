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

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/errors"
	common2 "github.com/ontio/ontology/p2pserver/common"
)

type ConsensusPayload struct {
	Version         uint32
	PrevHash        common.Uint256
	Height          uint32
	BookkeeperIndex uint16
	Timestamp       uint32
	Data            []byte
	Owner           keypair.PublicKey
	Signature       []byte
	PeerId          common2.PeerId
	hash            common.Uint256
}

//get the consensus payload hash
func (this *ConsensusPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

//Check whether header is correct
func (this *ConsensusPayload) Verify() error {
	sink := common.NewZeroCopySink(nil)
	this.SerializationUnsigned(sink)

	err := signature.Verify(this.Owner, sink.Bytes(), this.Signature)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetVerifyFail, fmt.Sprintf("signature verify error. buf:%v", sink.Bytes()))
	}
	return nil
}

//serialize the consensus payload
func (this *ConsensusPayload) ToArray() []byte {
	return common.SerializeToBytes(this)
}

func (this *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	//return sig.GetHashData(cp)
	return []byte{}
}

func (this *ConsensusPayload) Serialization(sink *common.ZeroCopySink) {
	this.SerializationUnsigned(sink)
	buf := keypair.SerializePublicKey(this.Owner)
	sink.WriteVarBytes(buf)
	sink.WriteVarBytes(this.Signature)
}

//Deserialize message payload
func (this *ConsensusPayload) Deserialization(source *common.ZeroCopySource) error {
	err := this.DeserializationUnsigned(source)
	if err != nil {
		return err
	}
	buf, _, irregular, eof := source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	this.Owner, err = keypair.DeserializePublicKey(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNetUnPackFail, "deserialize publickey error")
	}

	this.Signature, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func (this *ConsensusPayload) SerializationUnsigned(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.Version)
	sink.WriteHash(this.PrevHash)
	sink.WriteUint32(this.Height)
	sink.WriteUint16(this.BookkeeperIndex)
	sink.WriteUint32(this.Timestamp)
	sink.WriteVarBytes(this.Data)
}

func (this *ConsensusPayload) DeserializationUnsigned(source *common.ZeroCopySource) error {
	var irregular, eof bool
	this.Version, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.PrevHash, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Height, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.BookkeeperIndex, eof = source.NextUint16()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Timestamp, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Data, _, irregular, eof = source.NextVarBytes()
	if eof {
		return io.ErrUnexpectedEOF
	}
	if irregular {
		return common.ErrIrregularData
	}

	return nil
}
