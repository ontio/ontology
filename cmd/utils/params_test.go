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
	"testing"
)

func TestParseRawParamsArray(t *testing.T) {
	rawParamStr := "string:foo,[int:0,[bool:true,string:bar],bool:false]"
	res, size, err := parseRawParamsString(rawParamStr)
	if err != nil {
		t.Errorf("TestParseArrayParams error:%s", err)
		return
	}
	if size != len(rawParamStr) {
		t.Errorf("TestParseArrayParams size:%d != %d", size, len(rawParamStr))
		return
	}
	expect := []interface{}{"string" + PARAM_TYPE_SPLIT_INC + "foo", []interface{}{"int" + PARAM_TYPE_SPLIT_INC + "0", []interface{}{"bool" + PARAM_TYPE_SPLIT_INC + "true", "string" + PARAM_TYPE_SPLIT_INC + "bar"}, "bool" + PARAM_TYPE_SPLIT_INC + "false"}}
	ok, err := arrayEqual(res, expect)
	if err != nil {
		t.Errorf("TestParseArrayParams error:%s", err)
		return
	}
	if !ok {
		t.Errorf("TestParseArrayParams faild, res:%s != %s", res, expect)
		return
	}
}

func TestParseParams(t *testing.T) {
	testByteArray := []byte("HelloWorld")
	testByteArrayParam := hex.EncodeToString(testByteArray)
	rawParamStr := "bytearray:" + testByteArrayParam + ",string:foo,[int:0,[bool:true,string:bar],bool:false]"
	params, err := ParseParams(rawParamStr)
	if err != nil {
		t.Errorf("TestParseParams error:%s", err)
		return
	}
	data, err := json.Marshal(params)
	if err != nil {
		t.Errorf("json.Marshal error:%s", err)
		return
	}
	fmt.Printf("%s\n", data)

	expect := []interface{}{testByteArray, "foo", []interface{}{0, []interface{}{true, "bar"}, false}}
	ok, err := arrayEqual(params, expect)
	if err != nil {
		t.Errorf("TestParseParams error:%s", err)
		return
	}
	if !ok {
		t.Errorf("TestParseParams faild, res:%s != %s", params, expect)
		return
	}
}

func arrayEqual(a1, a2 []interface{}) (bool, error) {
	data1, err := json.Marshal(a1)
	if err != nil {
		return false, fmt.Errorf("json.Marshal:%s error:%s", a1, err)
	}
	data2, err := json.Marshal(a2)
	if err != nil {
		return false, fmt.Errorf("json.Marshal:%s error:%s", a2, err)
	}
	return string(data1) == string(data2), nil
}
