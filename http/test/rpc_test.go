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
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

var rpcaddr = "http://127.0.0.1:20336"

func reqPacking(method string, params []interface{}) map[string]interface{} {
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}
	return resp
}

func TestRpcConnectioncount(t *testing.T) {
	var req = reqPacking("getblockcount", []interface{}{})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}

func TestRpcGetBlockByHeight(t *testing.T) {
	var req = reqPacking("getblock", []interface{}{1})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcGetBlockByHash(t *testing.T) {
	var req = reqPacking("getblock", []interface{}{"ce536bccd56a5a26781b38845ea95551c1a9b622905bde724d0a08fa11c3fb04"})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcGetBlockHeight(t *testing.T) {
	var req = reqPacking("getblockcount", []interface{}{})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcGetTx(t *testing.T) {
	var req = reqPacking("getrawtransaction", []interface{}{"9661a4ae48b98c92e73ce2f685dfd1dabe878bfa6b3c73cd2892124ea1c9985e"})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcGetContract(t *testing.T) {
	var req = reqPacking("getcontractstate", []interface{}{"ff00000000000000000000000000000000000001"})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcGetEventByHeight(t *testing.T) {
	var req = reqPacking("getsmartcodeevent", []interface{}{1})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcTxBlockHeight(t *testing.T) {
	var req = reqPacking("getblockheightbytxhash", []interface{}{"aa"})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcGetStorage(t *testing.T) {
	var req = reqPacking("getstorage", []interface{}{"ff00000000000000000000000000000000000001", "0121dca8ffcba308e697ee9e734ce686f4181658"})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcSendRawTx(t *testing.T) {
	var req = reqPacking("sendrawtransaction", []interface{}{""})
	resp, err := Request("POST", req, rpcaddr)
	if err != nil {
		assert.Error(t, err)
	}
	r, _ := json.Marshal(resp)
	assert.Contains(t, string(r), "SUCCESS")
}
func TestRpcAll(t *testing.T) {
	TestRpcConnectioncount(t)
	TestRpcGetBlockByHeight(t)
	TestRpcGetBlockByHash(t)
	TestRpcGetBlockHeight(t)
	TestRpcGetTx(t)
	TestRpcGetContract(t)
	TestRpcGetEventByHeight(t)
	TestRpcTxBlockHeight(t)
	TestRpcGetStorage(t)
	TestRpcSendRawTx(t)
}
