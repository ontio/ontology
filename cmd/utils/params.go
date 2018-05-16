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
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	PARAMS_SPLIT          = ","
	PARAM_TYPE_SPLIT      = ":"
	PARAM_TYPE_ARRAY      = "array"
	PARAM_TYPE_BYTE_ARRAY = "bytearray"
	PARAM_TYPE_STRING     = "string"
	PARAM_TYPE_INTEGER    = "int"
	PARAM_TYPE_BOOLEAN    = "bool"
	PARAM_LEFT_BRACKET    = "["
	PARAM_RIGHT_BRACKET   = "]"
)

//ParseParams return interface{} array of encode params item.
//A param item compose of type and value, type can be: bytearray, string, int, bool
//Param type and param value split with ":", such as int:10
//Param array can be express with "[]", such [int:10,string:foo], param array can be nested, such as [int:10,[int:12,bool:true]]
//A raw params example: string:foo,[int:0,[bool:true,string:bar],bool:false]
func ParseParams(rawParamStr string) ([]interface{}, error) {
	rawParams, _, err := parseRawParamsString(rawParamStr)
	if err != nil {
		return nil, err
	}
	return parseRawParams(rawParams)
}

func parseRawParamsString(rawParamStr string) ([]interface{}, int, error) {
	if len(rawParamStr) == 0 {
		return nil, 0, nil
	}
	rawParamItems := make([]interface{}, 0)
	curRawParam := ""
	index := 0
	totalSize := len(rawParamStr)
	for i := 0; i < totalSize; i++ {
		s := string(rawParamStr[i])
		switch s {
		case PARAMS_SPLIT:
			curRawParam = strings.TrimSpace(curRawParam)
			if len(curRawParam) > 0 {
				rawParamItems = append(rawParamItems, curRawParam)
				curRawParam = ""
			}
		case PARAM_LEFT_BRACKET:
			if index == totalSize-1 {
				return rawParamItems, 0, nil
			}
			//clear current param as invalid input
			curRawParam = ""
			items, size, err := parseRawParamsString(string(rawParamStr[i+1:]))
			if err != nil {
				return nil, 0, fmt.Errorf("parse params error:%s", err)
			}
			if len(items) > 0 {
				rawParamItems = append(rawParamItems, items)
			}
			i += size
		case PARAM_RIGHT_BRACKET:
			curRawParam = strings.TrimSpace(curRawParam)
			if len(curRawParam) > 0 {
				rawParamItems = append(rawParamItems, curRawParam)
			}
			return rawParamItems, i + 1, nil
		default:
			curRawParam = fmt.Sprintf("%s%s", curRawParam, string(s))
		}
	}
	curRawParam = strings.TrimSpace(curRawParam)
	if len(curRawParam) != 0 {
		rawParamItems = append(rawParamItems, curRawParam)
	}
	return rawParamItems, totalSize, nil
}

func parseRawParams(rawParams []interface{}) ([]interface{}, error) {
	if len(rawParams) == 0 {
		return nil, nil
	}
	params := make([]interface{}, 0)
	for _, rawParam := range rawParams {
		switch v := rawParam.(type) {
		case string:
			param, err := parseRawParam(v)
			if err != nil {
				return nil, err
			}
			params = append(params, param)
		case []interface{}:
			res, err := parseRawParams(v)
			if err != nil {
				return nil, err
			}
			params = append(params, res)
		default:
			return nil, fmt.Errorf("unknown param type:%s", reflect.TypeOf(rawParam))
		}
	}
	return params, nil
}

func parseRawParam(rawParam string) (interface{}, error) {
	rawParam = strings.TrimSpace(rawParam)
	rawParam = strings.Trim(rawParam, PARAMS_SPLIT)
	if len(rawParam) == 0 {
		return nil, nil
	}
	ps := strings.Split(rawParam, PARAM_TYPE_SPLIT)
	if len(ps) != 2 {
		return nil, fmt.Errorf("invalid param:%s", rawParam)
	}
	pType := strings.TrimSpace(ps[0])
	pValue := strings.TrimSpace(ps[1])
	return parseRawParamValue(pType, pValue)
}

func parseRawParamValue(pType string, pValue string) (interface{}, error) {
	switch strings.ToLower(pType) {
	case PARAM_TYPE_BYTE_ARRAY:
		value, err := hex.DecodeString(pValue)
		if err != nil {
			return nil, fmt.Errorf("parse byte array param:%s error:%s", pValue, err)
		}
		return value, nil
	case PARAM_TYPE_STRING:
		return pValue, nil
	case PARAM_TYPE_INTEGER:
		value, err := strconv.ParseInt(pValue, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse integer param:%s error:%s", pValue, err)
		}
		return value, nil
	case PARAM_TYPE_BOOLEAN:
		switch strings.ToLower(pValue) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("parse boolean param:%s failed", pValue)
		}
	default:
		return nil, fmt.Errorf("unspport param type:%s", pType)
	}
}

//ParseReturnValue return the value of rawReturnTypeStr type.
//Return type can be: bytearray, string, int, bool.
//Types can be split with "," each other, such as int,string,bool
//Type array can be express with "[]", such [int,string], param array can be nested, such as [int,[int,bool]]
func ParseReturnValue(rawValue interface{}, rawReturnTypeStr string) ([]interface{}, error) {
	returnTypes, _, err := parseRawParamsString(rawReturnTypeStr)
	if err != nil {
		return nil, fmt.Errorf("Parse raw return types:%s error:%s", rawReturnTypeStr, err)
	}
	var rawValues []interface{}
	rawValues, ok := rawValue.([]interface{})
	if !ok {
		rawValues = append(rawValues, rawValue)
	}
	return parseReturnValueArray(rawValues, returnTypes)
}

func parseReturnValueArray(rawValues []interface{}, returnTypes []interface{}) ([]interface{}, error) {
	values := make([]interface{}, 0)
	for i := 0; i < len(rawValues); i++ {
		rawValue := rawValues[i]
		if i == len(returnTypes) {
			return values, nil
		}
		valueType := returnTypes[i]

		var err error
		switch v := rawValue.(type) {
		case string:
			var value interface{}
			vType := valueType.(string)
			switch strings.ToLower(vType) {
			case PARAM_TYPE_BYTE_ARRAY:
				value, err = ParseNeoVMContractReturnTypeByteArray(v)
			case PARAM_TYPE_STRING:
				value, err = ParseNeoVMContractReturnTypeString(v)
			case PARAM_TYPE_INTEGER:
				value, err = ParseNeoVMContractReturnTypeInteger(v)
			case PARAM_TYPE_BOOLEAN:
				value, err = ParseNeoVMContractReturnTypeBool(v)
			default:
				return nil, fmt.Errorf("unknown return type:%s", v)
			}
			values = append(values, value)
			if err != nil {
				return nil, fmt.Errorf("Parse return value:%s type:byte array error:%s", v, err)
			}
		case []interface{}:
			valueTypes, ok := valueType.([]interface{})
			if !ok {
				return nil, fmt.Errorf("Parse return value:%+v types:%s failed, types doesnot match", v, valueType)
			}
			values, err := parseReturnValueArray(v, valueTypes)
			if err != nil {
				return nil, fmt.Errorf("Parese return values:%+v types:%s error:%s", values, valueType, err)
			}
		default:
			return nil, fmt.Errorf("unknown return type:%s", reflect.TypeOf(rawValue))
		}
	}
	return values, nil
}

//NeoVMInvokeParam use to express the param to invoke neovm contract.
//Type can be of array, bytearray, string, int and bool
//If type is one of bytearray, string, int and bool, value must be a string
//If Type is array, value must be []*NeoVMInvokeParam
//Example:
//[]interface{}{
//	&NeoVMInvokeParam{
//		Type:  "string",
//		Value: "foo",
//	},
//	&NeoVMInvokeParam{
//		Type: "array",
//		Value: []interface{}{
//			&NeoVMInvokeParam{
//				Type:  "int",
//				Value: "0",
//			},
//			&NeoVMInvokeParam{
//				Type:  "bool",
//				Value: "true",
//			},
//		},
//	},
//}
type NeoVMInvokeParam struct {
	Type  string
	Value interface{} //string or []*NeoVMInvokeParam
}

//ParseNeoVMInvokeParams parse params to []interface, rawParams is array of NeoVMInvokeParam
func ParseNeoVMInvokeParams(rawParams []interface{}) ([]interface{}, error) {
	if len(rawParams) == 0 {
		return nil, nil
	}
	params := make([]interface{}, 0)
	for _, rawParam := range rawParams {
		rawParamItem, ok := rawParam.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid param %v", rawParam)
		}
		for k, v := range rawParamItem {
			rawParamItem[strings.ToLower(k)] = v
		}
		pType, ok := rawParamItem["type"]
		if !ok {
			return nil, fmt.Errorf("invalid param %v", rawParamItem)
		}
		pt, ok := pType.(string)
		if !ok {
			return nil, fmt.Errorf("invalid param %v", rawParamItem)
		}
		pValue, ok := rawParamItem["value"]
		if !ok {
			return nil, fmt.Errorf("invalid param %v", rawParamItem)
		}
		switch pv := pValue.(type) {
		case string:
			param, err := parseRawParamValue(pt, pv)
			if err != nil {
				return nil, fmt.Errorf("Parse Param type:%s value:%s error:%s", pType, pv, err)
			}
			params = append(params, param)
		case []interface{}:
			ps, err := ParseNeoVMInvokeParams(pv)
			if err != nil {
				return nil, err
			}
			if len(ps) > 0 {
				params = append(params, ps)
			}
		default:
			return nil, fmt.Errorf("invalid param %v", rawParamItem)
		}
	}
	return params, nil
}
