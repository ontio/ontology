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
	"github.com/ontio/ontology/smartcontract/states"
	"math"
	"math/big"
	"reflect"
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	neovm "github.com/ontio/ontology/smartcontract/service/neovm"
	vm "github.com/ontio/ontology/vm/neovm"
)

// NewDeployTransaction returns a deploy Transaction
func NewDeployTransaction(code []byte, name, version, author, email, desp string, vmType byte) *types.MutableTransaction {
	//TODO: check arguments
	DeployCodePayload := &payload.DeployCode{
		Code:        code,
		VmType:      vmType,
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
		TxType:  types.InvokeNeo,
		Payload: invokeCodePayload,
	}
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

func BuildNativeInvokeCode(contractAddress common.Address, version byte, method string, params []interface{}) ([]byte, error) {
	builder := vm.NewParamsBuilder(new(bytes.Buffer))
	err := BuildNeoVMParam(builder, params)
	if err != nil {
		return nil, err
	}
	builder.EmitPushByteArray([]byte(method))
	builder.EmitPushByteArray(contractAddress[:])
	builder.EmitPushInteger(new(big.Int).SetInt64(int64(version)))
	builder.Emit(vm.SYSCALL)
	builder.EmitPushByteArray([]byte(neovm.NATIVE_INVOKE_NAME))
	return builder.ToArray(), nil
}

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
					err := BuildNeoVMParam(builder, []interface{}{field.Interface()})
					if err != nil {
						return err
					}
					builder.Emit(vm.DUPFROMALTSTACK)
					builder.Emit(vm.SWAP)
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

//build param bytes for wasm contract
func BuildWasmVMInvokeCode(contractAddress common.Address, params []interface{}) ([]byte, error) {
	contract := &states.WasmContractParam{}
	contract.Address = contractAddress
	//bf := bytes.NewBuffer(nil)
	argbytes, err := BuildWasmContractParam(params)
	if err != nil {
		return nil, fmt.Errorf("build wasm contract param failed:%s", err)
	}
	contract.Args = argbytes
	sink := common.NewZeroCopySink(nil)
	contract.Serialization(sink)
	return sink.Bytes(), nil

}

//build param bytes for wasm contract
func BuildWasmContractParam(params []interface{}) ([]byte, error) {
	bf := common.NewZeroCopySink(nil)
	for _, param := range params {
		switch param.(type) {
		case string:
			bf.WriteString(param.(string))
		case int:
			bf.WriteInt32(param.(int32))
		case int64:
			bf.WriteInt64(param.(int64))
		case uint16:
			bf.WriteUint16(param.(uint16))
		case uint32:
			bf.WriteUint32(param.(uint32))
		case uint64:
			bf.WriteUint64(param.(uint64))
		case []byte:
			bf.WriteVarBytes(param.([]byte))
		case common.Uint256:
			bf.WriteHash(param.(common.Uint256))
		case common.Address:
			bf.WriteAddress(param.(common.Address))
		case byte:
			bf.WriteByte(param.(byte))
		case []interface{}:
			val, err := BuildWasmContractParam(param.([]interface{}))
			if err != nil {
				return nil, err
			}
			bf.WriteBytes(val)
		default:
			return nil, fmt.Errorf("not a supported type :%v\n", param)
		}
	}
	return bf.Bytes(), nil

}

func NewWasmVMInvokeTransaction(gasPrice, gasLimit uint64, contractAddress common.Address, params []interface{}) (*types.MutableTransaction, error) {
	invokeCode, err := BuildWasmVMInvokeCode(contractAddress, params)
	if err != nil {
		return nil, err
	}
	return NewWasmSmartContractTransaction(gasPrice, gasLimit, invokeCode)
}

func NewWasmSmartContractTransaction(gasPrice, gasLimit uint64, invokeCode []byte) (*types.MutableTransaction, error) {
	invokePayload := &payload.InvokeCode{
		Code: invokeCode,
	}
	tx := &types.MutableTransaction{
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		TxType:   types.InvokeWasm,
		Nonce:    uint32(time.Now().Unix()),
		Payload:  invokePayload,
		Sigs:     nil,
	}
	return tx, nil
}
