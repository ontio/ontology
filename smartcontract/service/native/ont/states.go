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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// Transfers
type Transfers struct {
	States []State
}

func (this *Transfers) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.States))); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize States length error!")
	}
	for _, v := range this.States {
		if err := v.Serialize(w); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Serialize States error!")
		}
	}
	return nil
}

func (this *Transfers) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, uint64(len(this.States)))
	for _, v := range this.States {
		v.Serialization(sink)
	}
}

func (this *Transfers) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize states length error!")
	}
	for i := 0; uint64(i) < n; i++ {
		var state State
		if err := state.Deserialize(r); err != nil {
			return errors.NewDetailErr(err, errors.ErrNoCode, "[TokenTransfer] Deserialize states error!")
		}
		this.States = append(this.States, state)
	}
	return nil
}

func (this *Transfers) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	for i := 0; uint64(i) < n; i++ {
		var state State
		if err := state.Deserialization(source); err != nil {
			return err
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
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("[State] serialize from error:%v", err)
	}
	if err := utils.WriteAddress(w, this.To); err != nil {
		return fmt.Errorf("[State] serialize to error:%v", err)
	}
	if err := utils.WriteVarUint(w, this.Value); err != nil {
		return fmt.Errorf("[State] serialize value error:%v", err)
	}
	return nil
}

func (this *State) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.From)
	utils.EncodeAddress(sink, this.To)
	utils.EncodeVarUint(sink, this.Value)
}

func (this *State) Deserialize(r io.Reader) error {
	var err error
	this.From, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("[State] deserialize from error:%v", err)
	}
	this.To, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("[State] deserialize to error:%v", err)
	}

	this.Value, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	return nil
}

func (this *State) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.From, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}

	this.To, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}

	this.Value, err = utils.DecodeVarUint(source)

	return err
}

type TransferFrom struct {
	Sender common.Address
	From   common.Address
	To     common.Address
	Value  uint64
}

func (this *TransferFrom) Serialize(w io.Writer) error {
	if err := utils.WriteAddress(w, this.Sender); err != nil {
		return fmt.Errorf("[TransferFrom] serialize sender error:%v", err)
	}
	if err := utils.WriteAddress(w, this.From); err != nil {
		return fmt.Errorf("[TransferFrom] serialize from error:%v", err)
	}
	if err := utils.WriteAddress(w, this.To); err != nil {
		return fmt.Errorf("[TransferFrom] serialize to error:%v", err)
	}
	if err := utils.WriteVarUint(w, this.Value); err != nil {
		return fmt.Errorf("[TransferFrom] serialize value error:%v", err)
	}
	return nil
}

func (this *TransferFrom) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Sender)
	utils.EncodeAddress(sink, this.From)
	utils.EncodeAddress(sink, this.To)
	utils.EncodeVarUint(sink, this.Value)
}

func (this *TransferFrom) Deserialize(r io.Reader) error {
	var err error
	this.Sender, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("[TransferFrom] deserialize sender error:%v", err)
	}

	this.From, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("[TransferFrom] deserialize from error:%v", err)
	}

	this.To, err = utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("[TransferFrom] deserialize to error:%v", err)
	}

	this.Value, err = utils.ReadVarUint(r)
	if err != nil {
		return err
	}
	return nil
}

func (this *TransferFrom) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Sender, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.From, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.To, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Value, err = utils.DecodeVarUint(source)

	return err
}
