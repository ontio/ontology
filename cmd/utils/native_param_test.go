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
	"github.com/ontio/ontology/cmd/abi"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm"
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
	addr := common.Address([20]byte{})
	address := addr.ToBase58()

	params := []interface{}{
		"Hello, World",
		"12",
		"true",
		address,
		"a757b22282b43e0852c48feae0892af19e48da8627296ef7a051993afb316b9b",
		"128",
		hex.EncodeToString([]byte("foo")),
		[]interface{}{"1", "2", "3", "4", "5", "6"},
		[]interface{}{"bar", "10"},
	}
	builder := neovm.NewParamsBuilder(new(bytes.Buffer))
	err := ParseNativeFuncParam(builder, "", params, paramAbi)
	if err != nil {
		t.Errorf("ParseNativeParam error:%s", err)
		return
	}
}
