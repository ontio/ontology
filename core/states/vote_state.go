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

package states

import (
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
)

type VoteState struct {
	StateBase
	PublicKeys []keypair.PublicKey
	Count      common.Fixed64
}

func (this *VoteState) Serialization(sink *common.ZeroCopySink) {
	this.StateBase.Serialization(sink)
	sink.WriteUint32(uint32(len(this.PublicKeys)))
	for _, v := range this.PublicKeys {
		buf := keypair.SerializePublicKey(v)
		sink.WriteVarBytes(buf)
	}
	sink.WriteUint64(uint64(this.Count))
}

func (this *VoteState) Deserialization(source *common.ZeroCopySource) error {
	err := this.StateBase.Deserialization(source)
	if err != nil {
		return err
	}
	n, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	for i := 0; i < int(n); i++ {
		buf, _, irregular, eof := source.NextVarBytes()
		if irregular {
			return common.ErrIrregularData
		}
		if eof {
			return io.ErrUnexpectedEOF
		}
		pk, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		this.PublicKeys = append(this.PublicKeys, pk)
	}
	c, eof := source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.Count = common.Fixed64(int64(c))
	return nil
}
