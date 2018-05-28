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
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/cmd/abi"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"strconv"
	"strings"
)

func ParseNativeParam(params []interface{}, paramAbi []*abi.NativeContractParamAbi) ([]byte, error) {
	if len(params) != len(paramAbi) {
		return nil, fmt.Errorf("abi unmatch")
	}
	var data []byte
	var err error
	buf := bytes.NewBuffer(nil)
	for i, param := range params {
		paramAbi := paramAbi[i]
		switch strings.ToLower(paramAbi.Type) {
		case abi.NATIVE_PARAM_TYPE_STRUCT:
			data, err = ParseNativeParamStruct(param, paramAbi)
			if err != nil {
				return nil, fmt.Errorf("param:%s parse:%v error:%s", paramAbi.Name, param)
			}
		case abi.NATIVE_PARAM_TYPE_ARRAY:
			data, err = ParseNativeParamArray(param, paramAbi)
		default:
			rawParam, ok := param.(string)
			if !ok {
				return nil, fmt.Errorf("param:%v assert to string failed", param)
			}
			switch strings.ToLower(paramAbi.Type) {
			case abi.NATIVE_PARAM_TYPE_ADDRESS:
				data, err = ParseNativeParamAddress(rawParam)
			case abi.NATIVE_PARAM_TYPE_BOOL:
				data, err = ParseNativeParamBool(rawParam)
			case abi.NATIVE_PARAM_TYPE_BYTE:
				data, err = ParseNativeParamByte(rawParam)
			case abi.NATIVE_PARAM_TYPE_BYTEARRAY:
				data, err = ParseNativeParamByteArray(rawParam)
			case abi.NATIVE_PARAM_TYPE_INTEGER:
				data, err = ParseNativeParamInteger(rawParam)
			case abi.NATIVE_PARAM_TYPE_STRING:
				data, err = ParseNativeParamString(rawParam)
			case abi.NATIVE_PARAM_TYPE_UINT256:
				data, err = ParseNativeParamUint256(rawParam)
			default:
				return nil, fmt.Errorf("unknown param type:%s", paramAbi.Type)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("param:%s parse:%v error:%s", paramAbi.Name, param, err)
		}
		_, err = buf.Write(data)
		if err != nil {
			return nil, fmt.Errorf("buf write error:%s", err)
		}
	}

	return buf.Bytes(), nil
}

func ParseNativeParamStruct(param interface{}, structAbi *abi.NativeContractParamAbi) ([]byte, error) {
	params, ok := param.([]interface{})
	if !ok {
		return nil, fmt.Errorf("assert to []interface{} failed")
	}
	return ParseNativeParam(params, structAbi.SubType)
}

func ParseNativeParamArray(param interface{}, arrayAbi *abi.NativeContractParamAbi) ([]byte, error) {
	params, ok := param.([]interface{})
	if !ok {
		return nil, fmt.Errorf("assert to []interface{} failed")
	}
	abis := make([]*abi.NativeContractParamAbi, 0, len(params))
	for i := 0; i < len(params); i++ {
		abis = append(abis, &abi.NativeContractParamAbi{
			Name:    fmt.Sprintf("%s_%d", arrayAbi.Name, i),
			Type:    arrayAbi.SubType[0].Type,
			SubType: arrayAbi.SubType[0].SubType,
		})
	}
	data, err := ParseNativeParam(params, abis)
	if err != nil {
		return nil, fmt.Errorf("parse array error:%s", err)
	}
	buf := bytes.NewBuffer(nil)
	serialization.WriteVarUint(buf, uint64(len(params)))
	_, err = buf.Write(data)
	if err != nil {
		return nil, fmt.Errorf("parse array error:%s", err)
	}
	return buf.Bytes(), nil
}

func ParseNativeParamByte(param string) ([]byte, error) {
	i, err := strconv.ParseInt(param, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse int error:%s", err)
	}
	buf := bytes.NewBuffer(nil)
	err = serialization.WriteByte(buf, byte(i))
	if err != nil {
		return nil, fmt.Errorf("write byte error:%s", err)
	}
	return buf.Bytes(), nil
}

func ParseNativeParamByteArray(param string) ([]byte, error) {
	data, err := hex.DecodeString(param)
	if err != nil {
		return nil, fmt.Errorf("hex decode string error:%s", err)
	}
	buf := bytes.NewBuffer(nil)
	err = serialization.WriteVarBytes(buf, data)
	if err != nil {
		return nil, fmt.Errorf("write bytes error:%s", err)
	}
	return buf.Bytes(), err
}

func ParseNativeParamUint256(param string) ([]byte, error) {
	data, err := hex.DecodeString(param)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	uint256, err := common.Uint256ParseFromBytes(data)
	if err != nil {
		return nil, fmt.Errorf("Uint256ParseFromBytes error:%s", err)
	}
	buf := bytes.NewBuffer(nil)
	err = uint256.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("uint256 serialize error:%s", err)
	}
	return buf.Bytes(), nil
}

func ParseNativeParamString(param string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := serialization.WriteString(buf, param)
	if err != nil {
		return nil, fmt.Errorf("write string error:%s", err)
	}
	return buf.Bytes(), nil
}

func ParseNativeParamInteger(param string) ([]byte, error) {
	i, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse int error:%s", err)
	}
	buf := bytes.NewBuffer(nil)
	err = serialization.WriteVarUint(buf, uint64(i))
	if err != nil {
		return nil, fmt.Errorf("write int error:%s", err)
	}
	return buf.Bytes(), nil
}

func ParseNativeParamBool(param string) ([]byte, error) {
	var b bool
	switch strings.ToLower(param) {
	case "true":
		b = true
	case "false":
		b = false
	default:
		return nil, fmt.Errorf("invalid bool value")
	}
	buf := bytes.NewBuffer(nil)
	err := serialization.WriteBool(buf, b)
	if err != nil {
		return nil, fmt.Errorf("write bool error:%s", err)
	}
	return buf.Bytes(), nil
}

func ParseNativeParamAddress(param string) ([]byte, error) {
	addr, err := common.AddressFromBase58(param)
	if err != nil {
		return nil, fmt.Errorf("AddressFromBase58 error:%s", err)
	}
	buf := bytes.NewBuffer(nil)
	err = addr.Serialize(buf)
	if err != nil {
		return nil, fmt.Errorf("address serialize error:%s", err)
	}
	return buf.Bytes(), nil
}
