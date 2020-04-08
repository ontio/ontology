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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ontio/ontology/common/config"
	rpcerr "github.com/ontio/ontology/http/base/error"
)

//JsonRpc version
const JSON_RPC_VERSION = "2.0"

const (
	ERROR_INVALID_PARAMS   = rpcerr.INVALID_PARAMS
	ERROR_ONTOLOGY_COMMON  = 10000
	ERROR_ONTOLOGY_SUCCESS = 0
)

type OntologyError struct {
	ErrorCode int64
	Error     error
}

func NewOntologyError(err error, errCode ...int64) *OntologyError {
	ontErr := &OntologyError{Error: err}
	if len(errCode) > 0 {
		ontErr.ErrorCode = errCode[0]
	} else {
		ontErr.ErrorCode = ERROR_ONTOLOGY_COMMON
	}
	if err == nil {
		ontErr.ErrorCode = ERROR_ONTOLOGY_SUCCESS
	}
	return ontErr
}

//JsonRpcRequest object in rpc
type JsonRpcRequest struct {
	Version string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

//JsonRpcResponse object response for JsonRpcRequest
type JsonRpcResponse struct {
	Error  int64           `json:"error"`
	Desc   string          `json:"desc"`
	Result json.RawMessage `json:"result"`
}

func sendRpcRequest(method string, params []interface{}) ([]byte, *OntologyError) {
	rpcReq := &JsonRpcRequest{
		Version: JSON_RPC_VERSION,
		Id:      "cli",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, NewOntologyError(fmt.Errorf("JsonRpcRequest json.Marshal error:%s", err))
	}

	addr := fmt.Sprintf("http://localhost:%d", config.DefConfig.Rpc.HttpJsonPort)
	resp, err := http.Post(addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, NewOntologyError(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NewOntologyError(fmt.Errorf("read rpc response body error:%s", err))
	}
	rpcRsp := &JsonRpcResponse{}
	err = json.Unmarshal(body, rpcRsp)
	if err != nil {
		return nil, NewOntologyError(fmt.Errorf("json.Unmarshal JsonRpcResponse:%s error:%s", body, err))
	}
	if rpcRsp.Error != 0 {
		return nil, NewOntologyError(fmt.Errorf("\n %s ", string(body)), rpcRsp.Error)
	}
	return rpcRsp.Result, nil
}
