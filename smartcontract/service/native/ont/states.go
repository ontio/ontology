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

package ont

import (
	"fmt"
	"io"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/vm/neovm/types"
)

// Transfers
type Transfers struct {
	States []*State
}

func (this *Transfers) Serialize(w io.Writer) error {
	if err := serialization.WriteVarUint(w, uint64(len(this.States))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize States length error!")
	}
	for _, v := range this.States {
		if err := v.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize States error!")
		}
	}
	return nil
}

func (this *Transfers) Deserialize(r io.Reader) error {
	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize states length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		state := new(State)
		if err := state.Deserialize(r); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize states error!")
		}
		this.States = append(this.States, state)
	}
	return nil
}

type State struct {
	From  common.Address
	To    common.Address
	Value uint64
}

func (this *State) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.From[:]); err != nil {
		return fmt.Errorf("[State] serialize from error:%v", err)
	}
	if err := serialization.WriteVarBytes(w, this.To[:]); err != nil {
		return fmt.Errorf("[State] serialize to error:%v", err)
	}
	if err := serialization.WriteVarBytes(w, types.BigIntToBytes(big.NewInt(int64(this.Value)))); err != nil {
		return fmt.Errorf("[State] serialize value error:%v", err)
	}
	return nil
}

func (this *State) Deserialize(r io.Reader) error {
	from, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[State] deserialize from error:%v", err)
	}
	this.From, err = common.AddressParseFromBytes(from)
	if err != nil {
		return fmt.Errorf("[State] address parse from bytes error:%v", err)
	}
	to, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[State] deserialize to error:%v", err)
	}
	this.To, err = common.AddressParseFromBytes(to)
	if err != nil {
		return fmt.Errorf("[State] address parse from bytes error:%v", err)
	}

	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[State] Deserialize value error:%v", err)
	}

	this.Value = types.BigIntFromBytes(value).Uint64()
	return nil
}

type TransferFrom struct {
	Sender common.Address
	From   common.Address
	To     common.Address
	Value  uint64
}

func (this *TransferFrom) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Sender[:]); err != nil {
		return fmt.Errorf("[TransferFrom] serialize sender error:%v", err)
	}
	if err := serialization.WriteVarBytes(w, this.From[:]); err != nil {
		return fmt.Errorf("[TransferFrom] serialize from error:%v", err)
	}
	if err := serialization.WriteVarBytes(w, this.To[:]); err != nil {
		return fmt.Errorf("[TransferFrom] serialize to error:%v", err)
	}
	if err := serialization.WriteVarBytes(w, types.BigIntToBytes(big.NewInt(int64(this.Value)))); err != nil {
		return fmt.Errorf("[TransferFrom] serialize value error:%v", err)
	}
	return nil
}

func (this *TransferFrom) Deserialize(r io.Reader) error {
	sender, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[TransferFrom] deserialize sender error:%v", err)
	}
	this.Sender, err = common.AddressParseFromBytes(sender)
	if err != nil {
		return fmt.Errorf("[TransferFrom] address parse from bytes error:%v", err)
	}

	from, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[TransferFrom] deserialize from error:%v", err)
	}
	this.From, err = common.AddressParseFromBytes(from)
	if err != nil {
		return fmt.Errorf("[TransferFrom] address parse from bytes error:%v", err)
	}

	to, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[TransferFrom] deserialize to error:%v", err)
	}
	this.To, err = common.AddressParseFromBytes(to)
	if err != nil {
		return fmt.Errorf("[TransferFrom] address parse from bytes error:%v", err)
	}

	value, err := serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize value error!")
	}

	this.Value = types.BigIntFromBytes(value).Uint64()
	return nil
}
