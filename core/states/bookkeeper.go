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
	"bytes"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
)

type BookkeeperState struct {
	StateBase
	CurrBookkeeper []*crypto.PubKey
	NextBookkeeper []*crypto.PubKey
}

func (this *BookkeeperState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.CurrBookkeeper)))
	for _, v := range this.CurrBookkeeper {
		v.Serialize(w)
	}
	serialization.WriteUint32(w, uint32(len(this.NextBookkeeper)))
	for _, v := range this.NextBookkeeper {
		v.Serialize(w)
	}
	return nil
}

func (this *BookkeeperState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(BookkeeperState)
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
		p := new(crypto.PubKey)
		err = p.DeSerialize(r)
		if err != nil {
			return err
		}
		this.CurrBookkeeper = append(this.CurrBookkeeper, p)
	}

	n, err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}
	for i := 0; i < int(n); i++ {
		p := new(crypto.PubKey)
		err = p.DeSerialize(r)
		if err != nil {
			return err
		}
		this.NextBookkeeper = append(this.NextBookkeeper, p)
	}
	return nil
}

func (v *BookkeeperState) ToArray() []byte {
	b := new(bytes.Buffer)
	v.Serialize(b)
	return b.Bytes()
}

