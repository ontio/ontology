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
	"testing"
)

func TestParseNativeParam(t *testing.T) {
	paramAbi := []*abi.NativeContractParamAbi{
		{
			Name: "Param1",
			Type: "String",
		},
		{
			Name: "Param2",
			Type: "Int",
		},
		{
			Name: "Param3",
			Type: "Bool",
		},
		{
			Name: "Param4",
			Type: "Address",
		},
		{
			Name: "Param5",
			Type: "Uint256",
		},
		{
			Name: "Param6",
			Type: "Byte",
		},
		{
			Name: "Param7",
			Type: "ByteArray",
		},
		{
			Name: "Param8",
			Type: "Array",
			SubType: []*abi.NativeContractParamAbi{
				{
					Name: "",
					Type: "Int",
				},
			},
		},
		{
			Name: "Param9",
			Type: "Struct",
			SubType: []*abi.NativeContractParamAbi{
				{
					Name: "Param9_0",
					Type: "String",
				},
				{
					Name: "Param9_1",
					Type: "Int",
				},
			},
		},
	}
	params := []interface{}{
		"Hello, World",
		"12",
		"true",
		"TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe",
		"a757b22282b43e0852c48feae0892af19e48da8627296ef7a051993afb316b9b",
		"128",
		hex.EncodeToString([]byte("foo")),
		[]interface{}{"1", "2", "3", "4", "5", "6"},
		[]interface{}{"bar", "10"},
	}

	_, err := ParseNativeParam(params, paramAbi)
	if err != nil {
		t.Errorf("ParseNativeParam error:%s", err)
		return
	}
}

func TestParseNativeParamAddress(t *testing.T) {
	address := "TA587BCw7HFwuUuzY1wg2HXCN7cHBPaXSe"
	data, err := ParseNativeParamAddress(address)
	if err != nil {
		t.Errorf("TestParseNativeParamAddress error:%s", err)
		return
	}
	address1 := new(common.Address)
	err = address1.Deserialize(bytes.NewBuffer(data))
	if err != nil {
		t.Errorf("TestParseNativeParamAddress error:%s", err)
		return
	}
	if address != address1.ToBase58() {
		t.Errorf("TestParseNativeParamAddress address %s != %s", address1.ToBase58(), address)
		return
	}
}

func TestParseNativeParamByte(t *testing.T) {
	param := byte(1)
	data, err := ParseNativeParamByte(fmt.Sprintf("%v", param))
	if err != nil {
		t.Errorf("TestParseNativeParamByte error:%s", err)
		return
	}
	b, err := serialization.ReadByte(bytes.NewReader(data))
	if err != nil {
		t.Errorf("TestParseNativeParamByte error:%s", err)
		return
	}
	if b != param {
		t.Errorf("TestParseNativeParamByte byte:%v != %v", b, param)
		return
	}
}

func TestParseNativeParamBool(t *testing.T) {
	param := true
	data, err := ParseNativeParamBool(fmt.Sprintf("%v", param))
	if err != nil {
		t.Errorf("TestParseNativeParamBool error:%s", err)
		return
	}
	b, err := serialization.ReadBool(bytes.NewReader(data))
	if err != nil {
		t.Errorf("TestParseNativeParamBool error:%s", err)
		return
	}
	if param != b {
		t.Errorf("TestParseNativeParamBool bool:%v != %v", b, param)
		return
	}
}

func TestParseNativeParamInteger(t *testing.T) {
	param := 1234
	data, err := ParseNativeParamInteger(fmt.Sprintf("%v", param))
	if err != nil {
		t.Errorf("TestParseNativeParamInteger error:%s", err)
		return
	}
	i, err := serialization.ReadVarUint(bytes.NewReader(data), 0)
	if err != nil {
		t.Errorf("TestParseNativeParamInteger error:%s", err)
		return
	}
	if int(i) != param {
		t.Errorf("TestParseNativeParamInteger int:%v != %v", i, param)
		return
	}
}

func TestParseNativeParamUint256(t *testing.T) {
	txHash := "a757b22282b43e0852c48feae0892af19e48da8627296ef7a051993afb316b9b"
	data, err := ParseNativeParamUint256(txHash)
	if err != nil {
		t.Errorf("TestParseNativeParamUint256 error:%s", err)
		return
	}
	u := &common.Uint256{}
	err = u.Deserialize(bytes.NewReader(data))
	if err != nil {
		t.Errorf("TestParseNativeParamUint256 error:%s", err)
		return
	}
	uStr := hex.EncodeToString(u.ToArray())
	if uStr != txHash {
		t.Errorf("TestParseNativeParamUint256 uint256:%s != %s", uStr, txHash)
		return
	}
}

func TestParseNativeParamByteArray(t *testing.T) {
	param := hex.EncodeToString([]byte("HelloWorld"))
	data, err := ParseNativeParamByteArray(param)
	if err != nil {
		t.Errorf("TestParseNativeParamByteArray error:%s", err)
		return
	}
	data, err = serialization.ReadVarBytes(bytes.NewReader(data))
	if err != nil {
		t.Errorf("TestParseNativeParamByteArray error:%s", err)
		return
	}
	by := hex.EncodeToString(data)
	if by != param {
		t.Errorf("TestParseNativeParamByteArray %s != %s", by, param)
		return
	}
}

func TestParseNativeParamString(t *testing.T) {
	param := "HelloWorld"
	data, err := ParseNativeParamString(param)
	if err != nil {
		t.Errorf("TestParseNativeParamString error:%s", err)
		return
	}
	str, err := serialization.ReadString(bytes.NewReader(data))
	if err != nil {
		t.Errorf("TestParseNativeParamString error:%s", err)
		return
	}
	if param != str {
		t.Errorf("TestParseNativeParamString %s != %s", str, param)
		return
	}
}
