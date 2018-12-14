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
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"math"
	"reflect"
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
		SideChainID: config.DefConfig.Genesis.SideChainID,
		Version:     types.TX_VERSION,
		TxType:      types.Deploy,
		Payload:     DeployCodePayload,
	}
}

// NewInvokeTransaction returns an invoke Transaction
func NewInvokeTransaction(code []byte) *types.MutableTransaction {
	//TODO: check arguments
	invokeCodePayload := &payload.InvokeCode{
		Code: code,
	}

	return &types.MutableTransaction{
		SideChainID: config.DefConfig.Genesis.SideChainID,
		Version:     types.TX_VERSION,
		TxType:      types.Invoke,
		Payload:     invokeCodePayload,
	}
}

//add for wasm vm native transaction call
func BuildWasmNativeTransaction(addr common.Address, version int, initMethod string, args interface{}) *types.MutableTransaction {

	p := []interface{}{args}

	bs, err := BuildNativeInvokeCode(addr, byte(version), initMethod, p)
	if err != nil {
		return nil
	}
	bf := bytes.NewBuffer(bs)

	tx := NewInvokeTransaction(bf.Bytes())
	tx.GasLimit = math.MaxUint64
	return tx
}

//add for wasm vm native transaction call
func BuildNativeInvokeCode(contractAddress common.Address, version byte, method string, params []interface{}) ([]byte, error) {
	bf := bytes.NewBuffer(nil)

	err := buildParam(params, bf)
	if err != nil {
		return nil, err
	}
	txstruct := TxStruct{
		Address: contractAddress[:],
		Method:  []byte(method),
		Version: int(version),
		Args:    bf.Bytes(),
	}

	bs, err := txstruct.Serialize()
	if err != nil {
		return nil, err
	}
	return bs, nil
}

//add for wasm vm invoke code
func BuildWasmInvokeCode(contractAddress common.Address, params []interface{}) ([]byte, error) {
	bf := bytes.NewBuffer(nil)
	if len(params) < 1 {
		return nil, errors.NewErr("params count error")
	}

	//method := params[0].(string)

	return bf.Bytes(), nil
}

func buildParam(params []interface{}, bf *bytes.Buffer) error {

	for _, p := range params {
		switch p.(type) {
		case common.Address:
			utils.WriteAddress(bf, p.(common.Address))
		case uint64, int, int32, int64:
			utils.WriteVarUint(bf, p.(uint64))
		case []*ont.State:
			utils.WriteVarUint(bf, uint64(len(p.([]*ont.State))))
			for _, s := range p.([]*ont.State) {
				utils.WriteAddress(bf, s.From)
				utils.WriteAddress(bf, s.To)
				utils.WriteVarUint(bf, s.Value)
			}
		case *ont.TransferFrom:
			tmp := p.(*ont.TransferFrom)
			utils.WriteAddress(bf, tmp.Sender)
			utils.WriteAddress(bf, tmp.From)
			utils.WriteAddress(bf, tmp.To)
			utils.WriteVarUint(bf, tmp.Value)

		case []string:
			utils.WriteVarUint(bf, uint64(len(p.([]string))))
			for _, s := range p.([]string) {
				serialization.WriteVarBytes(bf, []byte(s))
			}
		case string:
			serialization.WriteVarBytes(bf, []byte(p.(string)))
		case []byte:
			serialization.WriteVarBytes(bf, p.([]byte))
		case []interface{}:
			utils.WriteVarUint(bf, uint64(len(p.([]interface{}))))

			err := buildParam(p.([]interface{}), bf)
			if err != nil {
				return err
			}
			//}

		default:
			object := reflect.ValueOf(p)
			kind := object.Kind().String()
			if kind == "ptr" {
				object = object.Elem()
				kind = object.Kind().String()
			}
			switch kind {
			case "slice":
				ps := make([]interface{}, 0)
				for i := 0; i < object.Len(); i++ {
					ps = append(ps, object.Index(i).Interface())
				}
				return buildParam([]interface{}{ps}, bf)
			case "struct":
				for i := 0; i < object.NumField(); i++ {
					field := object.Field(i)

					err := buildParam([]interface{}{field.Interface()}, bf)
					if err != nil {
						return err
					}
				}
			default:
				return fmt.Errorf("unsupported param:%s", p)
			}

		}
	}

	return nil
}
