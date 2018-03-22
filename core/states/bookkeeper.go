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

type BookKeeperState struct {
	StateBase
	CurrBookKeeper []*crypto.PubKey
	NextBookKeeper []*crypto.PubKey
}

func (this *BookKeeperState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	serialization.WriteUint32(w, uint32(len(this.CurrBookKeeper)))
	for _, v := range this.CurrBookKeeper {
		v.Serialize(w)
	}
	serialization.WriteUint32(w, uint32(len(this.NextBookKeeper)))
	for _, v := range this.NextBookKeeper {
		v.Serialize(w)
	}
	return nil
}

func (this *BookKeeperState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(BookKeeperState)
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
		this.CurrBookKeeper = append(this.CurrBookKeeper, p)
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
		this.NextBookKeeper = append(this.NextBookKeeper, p)
	}
	return nil
}

func (v *BookKeeperState) ToArray() []byte {
	b := new(bytes.Buffer)
	v.Serialize(b)
	return b.Bytes()
}

