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
	"time"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/states"
	vm "github.com/ontio/ontology/vm/neovm"
)

const NATIVE_INVOKE_NAME = "Ontology.Native.Invoke" // copy from smartcontract/service/neovm/config.go to avoid cycle dependences

// NewDeployTransaction returns a deploy Transaction
func NewDeployTransaction(code []byte, name, version, author, email, desp string, vmType payload.VmType) (*types.MutableTransaction, error) {
	//TODO: check arguments
	depCode, err := payload.NewDeployCode(code, vmType, name, version, author, email, desp)
	if err != nil {
		return nil, err
	}
	return &types.MutableTransaction{
		TxType:  types.Deploy,
		Payload: depCode,
	}, nil
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
	builder.EmitPushByteArray([]byte(NATIVE_INVOKE_NAME))

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
	builder.EmitPushByteArray([]byte(NATIVE_INVOKE_NAME))
	return builder.ToArray(), nil
}

//BuildNeoVMInvokeCode build NeoVM Invoke code for params
func BuildNeoVMInvokeCode(smartContractAddress common.Address, params []interface{}) ([]byte, error) {
	builder := vm.NewParamsBuilder(new(bytes.Buffer))
	err := BuildNeoVMParam(builder, params)
	if err != nil {
		return nil, err
	}
	args := append(builder.ToArray(), 0x67)
	args = append(args, smartContractAddress[:]...)
	return args, nil
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

	return common.SerializeToBytes(contract), nil
}

//build param bytes for wasm contract
func BuildWasmContractParam(params []interface{}) ([]byte, error) {
	bf := common.NewZeroCopySink(nil)
	for _, param := range params {
		switch val := param.(type) {
		case string:
			bf.WriteString(val)
		case int:
			bf.WriteI128(common.I128FromInt64(int64(val)))
		case int64:
			bf.WriteI128(common.I128FromInt64(int64(val)))
		case uint16:
			bf.WriteI128(common.I128FromUint64(uint64(val)))
		case uint32:
			bf.WriteI128(common.I128FromUint64(uint64(val)))
		case uint64:
			bf.WriteI128(common.I128FromUint64(uint64(val)))
		case *big.Int:
			bint, err := common.I128FromBigInt(val)
			if err != nil {
				return nil, err
			}
			bf.WriteI128(bint)
		case big.Int:
			bint, err := common.I128FromBigInt(&val)
			if err != nil {
				return nil, err
			}
			bf.WriteI128(bint)
		case []byte:
			bf.WriteVarBytes(val)
		case common.Uint256:
			bf.WriteHash(val)
		case common.Address:
			bf.WriteAddress(val)
		case byte:
			bf.WriteByte(val)
		case bool:
			bf.WriteBool(val)
		case []interface{}:
			// actually if different type will pass tuple to wasm. or will pass array.
			vnum := len(val)
			bf.WriteVarUint(uint64(vnum))
			value, err := BuildWasmContractParam(val)
			if err != nil {
				return nil, err
			}
			bf.WriteBytes(value)
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
