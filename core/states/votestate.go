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
	"github.com/Ontology/crypto"
	"github.com/Ontology/common"
	"io"
	"github.com/Ontology/common/serialization"
)

type VoteState struct {
	StateBase
	PublicKeys []*crypto.PubKey
	Count      common.Fixed64
}

func (this *VoteState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.PublicKeys)))
	for _, v := range this.PublicKeys {
		err := v.Serialize(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *VoteState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(VoteState)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return err
	}
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		pk := new(crypto.PubKey)
		if err := pk.DeSerialize(r); err != nil {
			return err
		}
		this.PublicKeys = append(this.PublicKeys, pk)
	}
	return nil
}


