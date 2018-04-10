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
	"fmt"
	"os"
	"testing"
)

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
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}

func TestRpcGetBlockByHeight(t *testing.T) {
	var req = reqPacking("getblock", []interface{}{1})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}
func TestRpcGetBlockByHash(t *testing.T) {
	var req = reqPacking("getblock", []interface{}{"ce536bccd56a5a26781b38845ea95551c1a9b622905bde724d0a08fa11c3fb04"})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}
func TestRpcGetBlockHeight(t *testing.T) {
	var req = reqPacking("getblockcount", []interface{}{})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}
func TestRpcGetTx(t *testing.T) {
	var req = reqPacking("getrawtransaction", []interface{}{"9661a4ae48b98c92e73ce2f685dfd1dabe878bfa6b3c73cd2892124ea1c9985e"})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}
func TestRpcGetContract(t *testing.T) {
	var req = reqPacking("getcontractstate", []interface{}{"ff00000000000000000000000000000000000001"})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}
func TestRpcGetEventByHeight(t *testing.T) {
	var req = reqPacking("getcontractstate", []interface{}{1})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}
func TestRpcTxBlockHeight(t *testing.T) {
	var req = reqPacking("getblockheightbytxhash", []interface{}{"aa"})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
}
func TestRpcGetStorage(t *testing.T) {
	var req = reqPacking("getstorage", []interface{}{"ff00000000000000000000000000000000000001", "0121dca8ffcba308e697ee9e734ce686f4181658"})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
	//value, _ := common.HexToBytes(resp["Result"].(string))
	//fmt.Println(new(big.Int).SetBytes(value))
}
func TestRpcSendRawTx(t *testing.T) {
	var req = reqPacking("sendrawtransaction", []interface{}{""})
	q, _ := json.Marshal(req)
	fmt.Println(string(q))
	resp, err := Request("POST", req, "http://127.0.0.1:20386")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	r, _ := json.Marshal(resp)
	fmt.Println(string(r))
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
