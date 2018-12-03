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
	"math"
	"math/big"
	"reflect"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	neovm "github.com/ontio/ontology/smartcontract/service/neovm"
	vm "github.com/ontio/ontology/vm/neovm"
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
func BuildNativeTransaction(addr common.Address, initMethod string, args []byte) *types.MutableTransaction {
	bf := new(bytes.Buffer)
	builder := vm.NewParamsBuilder(bf)
	builder.EmitPushByteArray(args)
	builder.EmitPushByteArray([]byte(initMethod))
	builder.EmitPushByteArray(addr[:])
	builder.EmitPushInteger(big.NewInt(0))
	builder.Emit(vm.SYSCALL)
	builder.EmitPushByteArray([]byte(neovm.NATIVE_INVOKE_NAME))

	tx := NewInvokeTransaction(builder.ToArray())
	tx.GasLimit = math.MaxUint64
	return tx
}

//add for wasm vm native transaction call
func BuildNativeInvokeCode(contractAddress common.Address, version byte, method string, params []interface{}) ([]byte, error) {
	bf := bytes.NewBuffer(nil)

	for _, p := range params {
		switch p.(type) {
		case common.Address:
			utils.WriteAddress(bf, p.(common.Address))
		case uint64:
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
			for _, s := range p.([]interface{}) {
				serialization.WriteVarBytes(bf, []byte(s.(string)))
			}

		default:
			log.Errorf("[BuildNativeInvokeCode] unrecongnized params:%v\n", p)
		}
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

//func BuildNativeInvokeCode(contractAddress common.Address, version byte, method string, params []interface{}) ([]byte, error) {
//	builder := vm.NewParamsBuilder(new(bytes.Buffer))
//	err := BuildNeoVMParam(builder, params)
//	if err != nil {
//		return nil, err
//	}
//	builder.EmitPushByteArray([]byte(method))
//	builder.EmitPushByteArray(contractAddress[:])
//	builder.EmitPushInteger(new(big.Int).SetInt64(int64(version)))
//	builder.Emit(vm.SYSCALL)
//	builder.EmitPushByteArray([]byte(neovm.NATIVE_INVOKE_NAME))
//	return builder.ToArray(), nil
//}

//buildNeoVMParamInter build neovm invoke param code
func BuildNeoVMParam(builder *vm.ParamsBuilder, smartContractParams []interface{}) error {
	//VM load params in reverse order
	for i := len(smartContractParams) - 1; i >= 0; i-- {
		switch v := smartContractParams[i].(type) {
		case bool:
			builder.EmitPushBool(v)
		case byte:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case uint32:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case int64:
			builder.EmitPushInteger(big.NewInt(int64(v)))
		case common.Fixed64:
			builder.EmitPushInteger(big.NewInt(int64(v.GetData())))
		case uint64:
			val := big.NewInt(0)
			builder.EmitPushInteger(val.SetUint64(uint64(v)))
		case string:
			builder.EmitPushByteArray([]byte(v))
		case *big.Int:
			builder.EmitPushInteger(v)
		case []byte:
			builder.EmitPushByteArray(v)
		case common.Address:
			builder.EmitPushByteArray(v[:])
		case common.Uint256:
			builder.EmitPushByteArray(v.ToArray())
		case []interface{}:
			err := BuildNeoVMParam(builder, v)
			if err != nil {
				return err
			}
			builder.EmitPushInteger(big.NewInt(int64(len(v))))
			builder.Emit(vm.PACK)
		default:
			object := reflect.ValueOf(v)
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
				err := BuildNeoVMParam(builder, []interface{}{ps})
				if err != nil {
					return err
				}
			case "struct":
				builder.EmitPushInteger(big.NewInt(0))
				builder.Emit(vm.NEWSTRUCT)
				builder.Emit(vm.TOALTSTACK)
				for i := 0; i < object.NumField(); i++ {
					field := object.Field(i)
					builder.Emit(vm.DUPFROMALTSTACK)
					err := BuildNeoVMParam(builder, []interface{}{field.Interface()})
					if err != nil {
						return err
					}
					builder.Emit(vm.APPEND)
				}
				builder.Emit(vm.FROMALTSTACK)
			default:
				return fmt.Errorf("unsupported param:%s", v)
			}
		}
	}
	return nil
}
