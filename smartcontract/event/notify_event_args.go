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

package event

import (
	"encoding/json"
	"github.com/ontio/ontology/common"
)

const (
	CONTRACT_STATE_FAIL    byte = 0
	CONTRACT_STATE_SUCCESS byte = 1
)

// NotifyEventInfo describe smart contract event notify info struct
type NotifyEventInfo struct {
	ContractAddress common.Address
	States          interface{}
}

func (self NotifyEventInfo) MarshalJSON() ([]byte, error) {
	type Internal struct {
		ContractAddress string
		States          interface{}
	}
	in := Internal{
		ContractAddress: self.ContractAddress.ToHexString(),
		States:          self.States,
	}
	return json.Marshal(in)
}

func (self *NotifyEventInfo) UnmarshalJSON(data []byte) error {
	type Internal struct {
		ContractAddress string
		States          interface{}
	}
	in := Internal{}
	err := json.Unmarshal(data, &in)
	if err != nil {
		return err
	}
	self.States = in.States
	contractAddr, err := common.AddressFromHexString(in.ContractAddress)
	if err != nil {
		return err
	}
	self.ContractAddress = contractAddr
	return nil
}

type ExecuteNotify struct {
	TxHash      common.Uint256
	State       byte
	GasConsumed uint64
	Notify      []*NotifyEventInfo
}

func (self ExecuteNotify) MarshalJSON() ([]byte, error) {
	type Internal struct {
		TxHash      string
		State       byte
		GasConsumed uint64
		Notify      []*NotifyEventInfo
	}
	in := Internal{
		TxHash:      self.TxHash.ToHexString(),
		State:       self.State,
		GasConsumed: self.GasConsumed,
		Notify:      self.Notify,
	}
	return json.Marshal(in)
}

func (self *ExecuteNotify) UnmarshalJSON(data []byte) error {
	type Internal struct {
		TxHash      string
		State       byte
		GasConsumed uint64
		Notify      []*NotifyEventInfo
	}

	in := Internal{}
	err := json.Unmarshal(data, &in)
	if err != nil {
		return err
	}
	self.Notify = in.Notify
	hash, err := common.Uint256FromHexString(in.TxHash)
	if err != nil {
		return err
	}
	self.TxHash = hash
	self.GasConsumed = in.GasConsumed
	self.State = in.State
	return nil
}
