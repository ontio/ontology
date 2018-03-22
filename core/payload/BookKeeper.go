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

const BookKeeperPayloadVersion byte = 0x00

type BookKeeperAction byte

const (
	BookKeeperAction_ADD BookKeeperAction = 0
	BookKeeperAction_SUB BookKeeperAction = 1
)

type BookKeeper struct {
	PubKey *crypto.PubKey
	Action BookKeeperAction
	Cert   []byte
	Issuer *crypto.PubKey
}

func (self *BookKeeper) Serialize(w io.Writer) error {
	if err := self.PubKey.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], PubKey Serialize failed.")
	}
	if err := serialization.WriteVarBytes(w, []byte{byte(self.Action)}); err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Action Serialize failed.")
	}
	if err := serialization.WriteVarBytes(w, self.Cert); err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Cert Serialize failed.")
	}
	if err := self.Issuer.Serialize(w); err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Issuer Serialize failed.")
	}
	return nil
}

func (self *BookKeeper) Deserialize(r io.Reader) error {
	self.PubKey = new(crypto.PubKey)
	err := self.PubKey.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], PubKey Deserialize failed.")
	}
	var p [1]byte
	n, err := r.Read(p[:])
	if n == 0 {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Action Deserialize failed.")
	}
	self.Action = BookKeeperAction(p[0])
	self.Cert, err = serialization.ReadVarBytes(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Cert Deserialize failed.")
	}
	self.Issuer = new(crypto.PubKey)
	err = self.Issuer.DeSerialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BookKeeper], Issuer Deserialize failed.")
	}

	return nil
}
