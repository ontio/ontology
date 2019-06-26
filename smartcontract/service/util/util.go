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
package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ontio/ontology/common"
	cstate "github.com/ontio/ontology/smartcontract/states"
	"github.com/ontio/ontology/vm/neovm"
	neotypes "github.com/ontio/ontology/vm/neovm/types"
)

const (
	ByteArrayType byte = 0x00
	AddressType   byte = 0x01
	BooleanType   byte = 0x02
	UsizeType     byte = 0x03
	Int64Type     byte = 0x04
	Uint64Type    byte = 0x05
	Uint256Type   byte = 0x06
	ListType      byte = 0x07

	MAX_PARAM_LENGTH      = 1024
	VERSION          byte = 0
)

var ERROR_PARAM_FORMAT = fmt.Errorf("error param format")
var ERROR_PARAM_TOO_LONG = fmt.Errorf("param length is exceeded")
var ERROR_PARAM_NOT_SUPPORTED_TYPE = fmt.Errorf("error param format:not supported type")

//input byte array should be the following format
// version(1byte) + type(1byte) + usize( bytearray or list) (4 bytes) + data...

func DeserializeInput(input []byte) ([]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	if len(input) == 0 {
		return nil, ERROR_PARAM_FORMAT
	}
	if len(input) > MAX_PARAM_LENGTH {
		return nil, ERROR_PARAM_TOO_LONG
	}
	version := input[0]
	//current only support "0" version
	if version != VERSION {
		return nil, ERROR_PARAM_FORMAT
	}

	paramlist := make([]interface{}, 0)
	err := anaylzeInput(input[1:], &paramlist)
	if err != nil {
		return nil, err
	}

	return paramlist, nil
}

func anaylzeInput(input []byte, ret *[]interface{}) error {

	if input == nil || len(input) == 0 {
		return nil
	}

	switch input[0] {
	case ByteArrayType:
		//usize is 4 bytes
		if len(input[1:]) < 4 {
			return ERROR_PARAM_FORMAT
		}

		sizebytes := input[1:5]
		size := binary.LittleEndian.Uint32(sizebytes)
		if size == 0 {
			return ERROR_PARAM_FORMAT
		}

		if len(input[5:]) < int(size) {
			return ERROR_PARAM_FORMAT
		}

		bs := input[5 : 5+size]
		*ret = append(*ret, bs)
		return anaylzeInput(input[5+size:], ret)

	case AddressType:
		if len(input[1:]) < 20 {
			return ERROR_PARAM_FORMAT
		}
		addrbytes := input[1:21]
		address, err := common.AddressParseFromBytes(addrbytes)
		if err != nil {
			return err
		}
		*ret = append(*ret, address)
		return anaylzeInput(input[21:], ret)

	case BooleanType:
		if len(input[1:]) < 1 {
			return ERROR_PARAM_FORMAT
		}
		boolbyte := input[1]
		boolvalue := true
		if boolbyte != byte(1) {
			boolvalue = false
		}
		*ret = append(*ret, boolvalue)
		return anaylzeInput(input[2:], ret)
	case UsizeType:
		if len(input[1:]) < 4 {
			return ERROR_PARAM_FORMAT
		}
		i32bytes := input[1:5]
		i32 := binary.LittleEndian.Uint32(i32bytes)
		*ret = append(*ret, i32)
		return anaylzeInput(input[5:], ret)
	case Int64Type:
		if len(input[1:]) < 8 {
			return ERROR_PARAM_FORMAT
		}
		i64bytes := input[1:9]
		tmpbf := bytes.NewBuffer(i64bytes)
		var x int64
		binary.Read(tmpbf, binary.LittleEndian, &x)
		*ret = append(*ret, x)
		return anaylzeInput(input[9:], ret)
	case Uint64Type:
		if len(input[1:]) < 8 {
			return ERROR_PARAM_FORMAT
		}
		ui64bytes := input[1:9]
		ui64 := binary.LittleEndian.Uint64(ui64bytes)
		*ret = append(*ret, ui64)
		return anaylzeInput(input[9:], ret)
	case Uint256Type:
		if len(input[1:]) < 32 {
			return ERROR_PARAM_FORMAT
		}
		u256bytes := input[1:33]
		u256, err := common.Uint256ParseFromBytes(u256bytes)
		if err != nil {
			return err
		}
		*ret = append(*ret, u256)
		return anaylzeInput(input[33:], ret)
	case ListType:
		if len(input[1:]) < 4 {
			return ERROR_PARAM_FORMAT
		}

		sizebytes := input[1:5]
		size := binary.LittleEndian.Uint32(sizebytes)
		list := make([]interface{}, 0)
		rest, err := anaylzeList(input[5:], int(size), &list)

		if err != nil {
			return err
		}

		*ret = append(*ret, list)
		return anaylzeInput(rest, ret)
	default:
		return ERROR_PARAM_NOT_SUPPORTED_TYPE
	}

}

func anaylzeList(input []byte, listsize int, list *[]interface{}) ([]byte, error) {
	if input == nil || len(input) == 0 {
		return nil, nil
	}

	for i := 0; i < listsize; i++ {
		switch input[0] {
		case ByteArrayType:
			//usize is 4 bytes
			if len(input[1:]) < 4 {
				return nil, ERROR_PARAM_FORMAT
			}
			sizebytes := input[1:5]
			size := binary.LittleEndian.Uint32(sizebytes)
			if size == 0 {
				return nil, ERROR_PARAM_FORMAT
			}

			if len(input[5:]) < int(size) {
				return nil, ERROR_PARAM_FORMAT
			}
			bs := input[5 : 5+size]
			*list = append(*list, bs)
			input = input[5+size:]

		case AddressType:
			if len(input[1:]) < 20 {
				return nil, ERROR_PARAM_FORMAT
			}
			addrbytes := input[1:21]
			address, err := common.AddressParseFromBytes(addrbytes)
			if err != nil {
				return nil, err
			}
			*list = append(*list, address)
			input = input[21:]

		case BooleanType:
			if len(input[1:]) < 1 {
				return nil, ERROR_PARAM_FORMAT
			}
			boolbyte := input[1]
			boolvalue := true
			if boolbyte != byte(1) {
				boolvalue = false
			}
			*list = append(*list, boolvalue)
			input = input[2:]
		case UsizeType:
			if len(input[1:]) < 4 {
				return nil, ERROR_PARAM_FORMAT
			}
			i32bytes := input[1:5]
			i32 := binary.LittleEndian.Uint32(i32bytes)
			*list = append(*list, i32)
			input = input[5:]
		case Int64Type:
			if len(input[1:]) < 8 {
				return nil, ERROR_PARAM_FORMAT
			}
			i64bytes := input[1:9]
			tmpbf := bytes.NewBuffer(i64bytes)
			var x int64
			binary.Read(tmpbf, binary.LittleEndian, &x)
			*list = append(*list, x)
			input = input[9:]
		case Uint64Type:
			if len(input[1:]) < 8 {
				return nil, ERROR_PARAM_FORMAT
			}
			ui64bytes := input[1:9]
			ui64 := binary.LittleEndian.Uint64(ui64bytes)
			*list = append(*list, ui64)
			input = input[9:]
		case Uint256Type:
			if len(input[1:]) < 32 {
				return nil, ERROR_PARAM_FORMAT
			}
			u256bytes := input[1:33]
			u256, err := common.Uint256ParseFromBytes(u256bytes)
			if err != nil {
				return nil, err
			}
			*list = append(*list, u256)
			input = input[33:]
		case ListType:
			if len(input[1:]) < 4 {
				return nil, ERROR_PARAM_FORMAT
			}
			sizebytes := input[1:5]
			size := binary.LittleEndian.Uint32(sizebytes)
			sublist := make([]interface{}, 0)
			bs := input[5:]
			rest, err := anaylzeList(bs, int(size), &sublist)
			if err != nil {
				return nil, err
			}

			*list = append(*list, sublist)
			input = rest
		default:
			return nil, ERROR_PARAM_FORMAT
		}
	}

	return input, nil
}

func CreateNeoInvokeParam(contractAddress common.Address, input []byte) ([]byte, error) {

	list, err := DeserializeInput(input)
	if err != nil {
		return nil, err
	}

	if list == nil {
		return nil, nil
	}

	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err = BuildNeoVMParam(builder, list)
	if err != nil {
		return nil, err
	}
	args := append(builder.ToArray(), 0x67)
	args = append(args, contractAddress[:]...)
	return args, nil
}

//buildNeoVMParamInter build neovm invoke param code
func BuildNeoVMParam(builder *neovm.ParamsBuilder, smartContractParams []interface{}) error {
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
			builder.Emit(neovm.PACK)
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
				builder.Emit(neovm.NEWSTRUCT)
				builder.Emit(neovm.TOALTSTACK)
				for i := 0; i < object.NumField(); i++ {
					field := object.Field(i)
					builder.Emit(neovm.DUPFROMALTSTACK)
					err := BuildNeoVMParam(builder, []interface{}{field.Interface()})
					if err != nil {
						return err
					}
					builder.Emit(neovm.APPEND)
				}
				builder.Emit(neovm.FROMALTSTACK)
			default:
				return fmt.Errorf("unsupported param:%s", v)
			}
		}
	}
	return nil
}

//build param bytes for wasm contract
func BuildWasmVMInvokeCode(contractAddress common.Address, params []interface{}) ([]byte, error) {
	contract := &cstate.WasmContractParam{}
	contract.Address = contractAddress
	//bf := bytes.NewBuffer(nil)
	bf := common.NewZeroCopySink(nil)
	argbytes, err := buildWasmContractParam(params, bf)
	if err != nil {
		return nil, fmt.Errorf("build wasm contract param failed:%s", err)
	}
	contract.Args = argbytes
	sink := common.NewZeroCopySink(nil)
	contract.Serialization(sink)
	return sink.Bytes(), nil

}

//build param bytes for wasm contract
func buildWasmContractParam(params []interface{}, bf *common.ZeroCopySink) ([]byte, error) {
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
			buildWasmContractParam(param.([]interface{}), bf)
		default:
			return nil, fmt.Errorf("not a supported type :%v\n", param)
		}
	}
	return bf.Bytes(), nil

}

func BuildResultFromNeo(item neotypes.StackItems, bf *bytes.Buffer) error {

	switch item.(type) {
	case *neotypes.ByteArray:
		bs, err := item.GetByteArray()
		if err != nil {
			return err
		}
		bf.WriteByte(byte(ByteArrayType))
		size := uint32(len(bs))
		tmpbytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(tmpbytes, size)
		bf.Write(tmpbytes)
		bf.Write(bs)

	case *neotypes.Integer:
		val, err := item.GetBigInteger()
		if err != nil {
			return err
		}
		bf.WriteByte(byte(Int64Type))
		tmpbf := bytes.NewBuffer(nil)
		err = binary.Write(tmpbf, binary.LittleEndian, val.Int64())
		if err != nil {
			return err
		}
		bf.Write(tmpbf.Bytes())
	case *neotypes.Boolean:
		val, err := item.GetBoolean()
		if err != nil {
			return err
		}
		bf.WriteByte(byte(BooleanType))
		if val {
			bf.WriteByte(byte(1))
		} else {
			bf.WriteByte(byte(0))
		}
	case *neotypes.Array:
		val, err := item.GetArray()
		if err != nil {
			return err
		}
		if val == nil {
			return fmt.Errorf("get array error")
		}

		bf.WriteByte(byte(ListType))
		size := uint32(len(val))
		tmpbs := make([]byte, 4)
		binary.LittleEndian.PutUint32(tmpbs, size)
		bf.Write(tmpbs)
		for _, si := range val {
			err = BuildResultFromNeo(si, bf)
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("not a supported return type")
	}
	return nil
}
