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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ontio/ontology/cmd/abi"
)

func NewNeovmContractAbi(abiData []byte) (*abi.NeovmContractAbi, error) {
	abi := &abi.NeovmContractAbi{}
	err := json.Unmarshal(abiData, abi)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal NeovmContractAbi error:%s", err)
	}
	return abi, nil
}

func ParseNeovmFunc(rawParams []string, funcAbi *abi.NeovmContractFunctionAbi) ([]interface{}, error) {
	res := make([]interface{}, 0)
	funcName := convertNeovmFuncName(funcAbi.Name)
	res = append(res, funcName)
	params, err := ParseNeovmParam(rawParams, funcAbi.Parameters)
	if err != nil {
		return nil, err
	}
	res = append(res, params)
	return res, nil
}

//Neovm func name in Camel-Case. For example: transfer, transferFrom
func convertNeovmFuncName(name string) string {
	if name == "" {
		return name
	}
	data := []byte(name)
	data[0] = strings.ToLower(string(data[0]))[0]
	return string(data)
}

func ParseNeovmParam(params []string, paramsAbi []*abi.NeovmContractParamsAbi) ([]interface{}, error) {
	if len(params) != len(paramsAbi) {
		return nil, fmt.Errorf("abi param not match")
	}
	val := make([]interface{}, 0)
	for i, rawParam := range params {
		paramAbi := paramsAbi[i]
		rawParam = strings.TrimSpace(rawParam)
		var res interface{}
		var err error
		switch strings.ToLower(paramAbi.Type) {
		case abi.NEOVM_PARAM_TYPE_INTEGER:
			res, err = ParseNeovmParamInteger(rawParam)
		case abi.NEOVM_PARAM_TYPE_BOOL:
			res, err = ParseNeovmParamBoolean(rawParam)
		case abi.NEOVM_PARAM_TYPE_STRING:
			res, err = ParseNeovmParamString(rawParam)
		case abi.NEOVM_PARAM_TYPE_BYTE_ARRAY:
			res, err = ParseNeovmParamByteArray(rawParam)
		default:
			return nil, fmt.Errorf("unknown param type:%s", paramAbi.Type)
		}
		if err != nil {
			return nil, fmt.Errorf("parse param:%s value:%s type:%s error:%s", paramAbi.Name, rawParam, paramAbi.Type, err)
		}
		val = append(val, res)
	}
	return val, nil
}

func ParseNeovmParamString(param string) (interface{}, error) {
	return param, nil
}

func ParseNeovmParamInteger(param string) (interface{}, error) {
	if param == "" {
		return nil, fmt.Errorf("invalid integer")
	}
	value, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse integer param:%s error:%s", param, err)
	}
	return value, nil
}

func ParseNeovmParamBoolean(param string) (interface{}, error) {
	var res bool
	switch strings.ToLower(param) {
	case "true":
		res = true
	case "false":
		res = false
	default:
		return nil, fmt.Errorf("parse boolean param:%s failed", param)
	}
	return res, nil
}

func ParseNeovmParamByteArray(param string) (interface{}, error) {
	res, err := hex.DecodeString(param)
	if err != nil {
		return nil, fmt.Errorf("parse byte array param:%s error:%s", param, err)
	}
	return res, nil
}
