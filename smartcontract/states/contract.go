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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
)

// Invoke smart contract struct
// Param Version: invoke smart contract version, default 0
// Param Address: invoke on blockchain smart contract by address
// Param Method: invoke smart contract method, default ""
// Param Args: invoke smart contract arguments
type Contract struct {
	Version byte
	Address common.Address
	Method  string
	Args    []byte
}

// Serialize contract
func (this *Contract) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, this.Version); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Version serialize error!")
	}
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Address serialize error!")
	}
	if err := serialization.WriteVarBytes(w, []byte(this.Method)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Method serialize error!")
	}
	if err := serialization.WriteVarBytes(w, this.Args); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Args serialize error!")
	}
	return nil
}

// Deserialize contract
func (this *Contract) Deserialize(r io.Reader) error {
	var err error
	this.Version, err = serialization.ReadByte(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Version deserialize error!")
	}

	if err := this.Address.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Address deserialize error!")
	}

	method, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Method deserialize error!")
	}
	this.Method = string(method)

	this.Args, err = serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[Contract] Args deserialize error!")
	}
	return nil
}

type PreExecResult struct {
	State  byte
	Gas    uint64
	Result interface{}
}
