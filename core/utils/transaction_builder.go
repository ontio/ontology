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

package utils

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"math"
)

type TxStruct struct {
	Address []byte `json:"address"`
	Method  []byte `json:"method"`
	Version int    `json:"version"`
	Args    []byte `json:"args"`
}

func (txs *TxStruct) Serialize() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	err := serialization.WriteVarBytes(buffer, txs.Address)
	if err != nil {
		return nil, err
	}
	err = serialization.WriteVarBytes(buffer, txs.Method)
	if err != nil {
		return nil, err
	}
	err = serialization.WriteUint32(buffer, uint32(txs.Version))
	if err != nil {
		return nil, err
	}
	err = serialization.WriteVarBytes(buffer, txs.Args)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (txs *TxStruct) Deserialize(data []byte) error {

	buffer := bytes.NewBuffer(data)
	address, err := serialization.ReadVarBytes(buffer)
	if err != nil {
		return err
	}

	method, err := serialization.ReadVarBytes(buffer)
	if err != nil {
		return err
	}
	version, err := serialization.ReadUint32(buffer)
	if err != nil {
		return err
	}

	args, err := serialization.ReadVarBytes(buffer)
	if err != nil {
		return err
	}

	txs.Args = args
	txs.Version = int(version)
	txs.Method = method
	txs.Address = address

	return nil
}

// NewDeployTransaction returns a deploy Transaction
func NewDeployTransaction(code []byte, name, version, author, email, desp string, needStorage bool) *types.MutableTransaction {
	//TODO: check arguments
	DeployCodePayload := &payload.DeployCode{
		Code:        code,
		NeedStorage: needStorage,
		Name:        name,
		Version:     version,
		Author:      author,
		Email:       email,
		Description: desp,
	}

	return &types.MutableTransaction{
		TxType:  types.Deploy,
		Payload: DeployCodePayload,
	}
}

// NewInvokeTransaction returns an invoke Transaction
func NewInvokeTransaction(code []byte) *types.MutableTransaction {
	//TODO: check arguments
	invokeCodePayload := &payload.InvokeCode{
		Code: code,
	}

	return &types.MutableTransaction{
		TxType:  types.Invoke,
		Payload: invokeCodePayload,
	}
}

//add for wasm vm native transaction call
func BuildWasmNativeTransaction(addr common.Address, version int, initMethod string, args []byte) *types.MutableTransaction {
	txstruct := TxStruct{
		Address: addr[:],
		Method:  []byte(initMethod),
		Version: version,
		Args:    args,
	}
	bs, err := txstruct.Serialize()
	if err != nil {
		return nil
	}

	tx := NewInvokeTransaction(bs)
	tx.GasLimit = math.MaxUint64
	return tx
}
