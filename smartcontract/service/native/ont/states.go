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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"io"
)

// Transfers
type Transfers struct {
	Version byte
	States  []*State
}

func (this *Transfers) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, byte(this.Version)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize version error!")
	}
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
	version, err := serialization.ReadByte(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize version error!")
	}
	this.Version = version

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
	Version byte
	From    common.Address
	To      common.Address
	Value   uint64
}

func (this *State) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, byte(this.Version)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize version error!")
	}
	if err := this.From.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize From error!")
	}
	if err := this.To.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize To error!")
	}
	if err := serialization.WriteUint64(w, this.Value); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Serialize Value error!")
	}
	return nil
}

func (this *State) Deserialize(r io.Reader) error {
	version, err := serialization.ReadByte(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize version error!")
	}
	this.Version = version

	if err := this.From.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize from error!")
	}

	if err := this.To.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize to error!")
	}

	value, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[State] Deserialize value error!")
	}

	this.Value = value
	return nil
}

type TransferFrom struct {
	Version byte
	Sender  common.Address
	From    common.Address
	To      common.Address
	Value   uint64
}

func (this *TransferFrom) Serialize(w io.Writer) error {
	if err := serialization.WriteByte(w, byte(this.Version)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize version error!")
	}
	if err := this.Sender.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize sender error!")
	}
	if err := this.From.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize from error!")
	}
	if err := this.To.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize to error!")
	}
	if err := serialization.WriteUint64(w, this.Value); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Serialize value error!")
	}
	return nil
}

func (this *TransferFrom) Deserialize(r io.Reader) error {
	version, err := serialization.ReadByte(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize version error!")
	}
	this.Version = version

	if err := this.Sender.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize sender error!")
	}

	if err := this.From.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize from error!")
	}

	if err := this.To.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize to error!")
	}

	value, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TransferFrom] Deserialize value error!")
	}

	this.Value = value
	return nil
}
