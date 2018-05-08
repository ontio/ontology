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

package test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

var addr = "http://127.0.0.1:20334"

func TestGenerateblocktime(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/node/generateblocktime")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestConnectioncount(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/node/connectioncount")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetBlockTxs(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/block/transactions/height/3000")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetBlockByHeight(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/block/details/height/13?raw=0")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetBlockByHash(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/block/details/hash/6e2c2afacc0ac9e5699bc7f92194ca37d23340ebfc6c9301aa74dc70eb69c280")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetBlockHeight(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/block/height")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetTx(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/transaction/7372a2ea037c13e9b0d8f020b07b8c041acde3f6e7c8326c8ff638c08120bee9")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetContract(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/contract/ff00000000000000000000000000000000000001")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetEventByHeight(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/smartcode/event/transactions/11")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestTxBlockHeight(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/block/height/txhash/d0378f808ecc19d61143ade0be0044203666851f9cfe254d748958c790901ca7")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestGetStorage(t *testing.T) {
	resp, err := Request("GET", nil, addr+"/api/v1/storage/ff00000000000000000000000000000000000001/0121dca8ffcba308e697ee9e734ce686f4181658")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestSendRawTx(t *testing.T) {
	var req = map[string]interface{}{
		"Action":  "sendrawtransaction",
		"Version": "1.0.0",
		"Data":    "",
	}
	resp, err := Request("POST", req, addr+"/api/v1/transaction")
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestAll(t *testing.T) {
	TestGenerateblocktime(t)
	TestConnectioncount(t)
	TestGetBlockTxs(t)
	TestGetBlockByHeight(t)
	TestGetBlockByHash(t)
	TestGetBlockHeight(t)
	TestGetTx(t)
	TestGetContract(t)
	TestGetEventByHeight(t)
	TestTxBlockHeight(t)
	TestGetStorage(t)
	TestSendRawTx(t)
}

func Request(method string, cmd map[string]interface{}, url string) (map[string]interface{}, error) {
	hClient := &http.Client{}
	var repMsg = make(map[string]interface{})
	var response *http.Response
	var err error
	switch method {
	case "GET":
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return repMsg, err
		}
		response, err = hClient.Do(req)
	case "POST":
		data, err := json.Marshal(cmd)
		if err != nil {
			return repMsg, err
		}
		reqData := bytes.NewBuffer(data)
		req, err := http.NewRequest("POST", url, reqData)
		if err != nil {
			return repMsg, err
		}
		req.Header.Set("Content-type", "application/json")
		response, err = hClient.Do(req)
	default:
		return repMsg, err
	}
	if response != nil {
		defer response.Body.Close()

		body, _ := ioutil.ReadAll(response.Body)
		if err := json.Unmarshal(body, &repMsg); err == nil {
			return repMsg, err
		}
	}
	if err != nil {
		return repMsg, err
	}
	return repMsg, err
}
