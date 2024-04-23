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
	"errors"

	"github.com/laizy/bigint"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

// TransferStates
type TransferStates struct {
	States         []TransferState
	uint64Wrapping bool
}

type TransferStatesV2 struct {
	States []*TransferStateV2
}

func (this *TransferStates) ToV2() *TransferStatesV2 {
	var states []*TransferStateV2
	for _, s := range this.States {
		states = append(states, s.ToV2())
	}
	return &TransferStatesV2{States: states}
}

func (this *TransferStates) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, uint64(len(this.States)))
	for _, v := range this.States {
		v.Serialization(sink)
	}
}

func (this *TransferStates) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	for i := 0; uint64(i) < n; i++ {
		var state TransferState
		state.uint64Wrapping = this.uint64Wrapping
		if err := state.Deserialization(source); err != nil {
			return err
		}
		this.States = append(this.States, state)
	}
	return nil
}

func (this *TransferStatesV2) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	for i := 0; uint64(i) < n; i++ {
		var state TransferStateV2
		if err := state.Deserialization(source); err != nil {
			return err
		}
		this.States = append(this.States, &state)
	}
	return nil
}

func (this *TransferStatesV2) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, uint64(len(this.States)))
	for _, state := range this.States {
		state.Serialization(sink)
	}
}

type TransferStateV2 struct {
	From  common.Address
	To    common.Address
	Value states.NativeTokenBalance
}

func (this *TransferStateV2) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.From)
	utils.EncodeAddress(sink, this.To)
	utils.EncodeVarBytes(sink, common.BigIntToNeoBytes(this.Value.Balance.BigInt()))
}

func (this *TransferStateV2) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.From, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}

	this.To, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}

	buf, err := utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	value := bigint.New(common.BigIntFromNeoBytes(buf))
	if value.LessThan(bigint.New(0)) {
		return errors.New("nagative value")
	}
	this.Value = states.NativeTokenBalance{Balance: bigint.New(common.BigIntFromNeoBytes(buf))}
	return nil
}

type TransferState struct {
	From           common.Address
	To             common.Address
	Value          uint64
	uint64Wrapping bool
}

func (this *TransferState) ToV2() *TransferStateV2 {
	return &TransferStateV2{
		From:  this.From,
		To:    this.To,
		Value: states.NativeTokenBalanceFromInteger(this.Value),
	}
}

func (this *TransferState) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.From)
	utils.EncodeAddress(sink, this.To)
	utils.EncodeVarUint(sink, this.Value)
}

func (this *TransferState) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.From, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}

	this.To, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}

	if this.uint64Wrapping {
		this.Value, err = utils.DecodeVarUintWrapping(source)
	} else {
		this.Value, err = utils.DecodeVarUint(source)
	}

	return err
}

type TransferFrom struct {
	Sender common.Address
	TransferState
}

func NewTransferFromState(sender, from, to common.Address, value uint64) *TransferFrom {
	return &TransferFrom{
		Sender: sender,
		TransferState: TransferState{
			From:  from,
			To:    to,
			Value: value,
		},
	}
}

func (self *TransferFrom) ToV2() *TransferFromStateV2 {
	return &TransferFromStateV2{
		Sender:          self.Sender,
		TransferStateV2: *self.TransferState.ToV2(),
	}
}

type TransferFromStateV2 struct {
	Sender common.Address
	TransferStateV2
}

func (self *TransferFromStateV2) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, self.Sender)
	self.TransferStateV2.Serialization(sink)
}

func (self *TransferFromStateV2) Deserialization(source *common.ZeroCopySource) error {
	var err error
	self.Sender, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	return self.TransferStateV2.Deserialization(source)
}

func (this *TransferFrom) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Sender)
	utils.EncodeAddress(sink, this.From)
	utils.EncodeAddress(sink, this.To)
	utils.EncodeVarUint(sink, this.Value)
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
