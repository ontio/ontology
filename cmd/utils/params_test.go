package utils

import (
	"encoding/json"
	"fmt"
	"testing"
	"encoding/hex"
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
	expect := []interface{}{"string:foo", []interface{}{"int:0", []interface{}{"bool:true", "string:bar"}, "bool:false"}}
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
	rawParamStr := "bytearray:"+testByteArrayParam+",string:foo,[int:0,[bool:true,string:bar],bool:false]"
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
