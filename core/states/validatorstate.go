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
	"io"
	. "github.com/Ontology/errors"
)

type ValidatorState struct {
	StateBase
	PublicKey *crypto.PubKey
}

func (this *ValidatorState) Serialize(w io.Writer) error {
	this.StateBase.Serialize(w)
	if err := this.PublicKey.Serialize(w); err != nil {
		return err
	}
	return nil
}

func (this *ValidatorState) Deserialize(r io.Reader) error {
	if this == nil {
		this = new(ValidatorState)
	}
	err := this.StateBase.Deserialize(r)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[ValidatorState], StateBase Deserialize failed.")
	}
	pk := new(crypto.PubKey)
	if err := pk.DeSerialize(r); err != nil {
		return NewDetailErr(err, ErrNoCode, "[ValidatorState], PublicKey Deserialize failed.")
	}
	this.PublicKey = pk
	return nil
}