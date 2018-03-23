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

package payload

import (
	"io"

	"github.com/Ontology/common/serialization"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

const BookkeeperPayloadVersion byte = 0x00

type BookkeeperAction byte

const (
	BookkeeperAction_ADD BookkeeperAction = 0
	BookkeeperAction_SUB BookkeeperAction = 1
)

type Bookkeeper struct {
	PubKey *crypto.PubKey
	Action BookkeeperAction
	Cert   []byte
	Issuer *crypto.PubKey
}

func (self *Bookkeeper) Serialize(w io.Writer) error {
	if err := self.PubKey.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], PubKey Serialize failed.")
	}
	if err := serialization.WriteVarBytes(w, []byte{byte(self.Action)}); err != nil {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], Action Serialize failed.")
	}
	if err := serialization.WriteVarBytes(w, self.Cert); err != nil {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], Cert Serialize failed.")
	}
	if err := self.Issuer.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], Issuer Serialize failed.")
	}
	return nil
}

func (self *Bookkeeper) Deserialize(r io.Reader) error {
	self.PubKey = new(crypto.PubKey)
	err := self.PubKey.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], PubKey Deserialize failed.")
	}
	var p [1]byte
	n, err := r.Read(p[:])
	if n == 0 {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], Action Deserialize failed.")
	}
	self.Action = BookkeeperAction(p[0])
	self.Cert, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], Cert Deserialize failed.")
	}
	self.Issuer = new(crypto.PubKey)
	err = self.Issuer.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Bookkeeper], Issuer Deserialize failed.")
	}

	return nil
}
