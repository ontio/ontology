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

package server

import (
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/cmd/server/common"
	"github.com/ontio/ontology/common/log"
	"io/ioutil"
	"net/http"
)

var DefCliRpcSvr = NewCliRpcServer()

type CliRpcServer struct {
	port       uint
	handlers   map[string]func(req *common.CliRpcRequest, resp *common.CliRpcResponse)
	httpSvr    *http.Server
	httpSvtMux *http.ServeMux
}

func NewCliRpcServer() *CliRpcServer {
	return &CliRpcServer{
		handlers: make(map[string]func(req *common.CliRpcRequest, resp *common.CliRpcResponse)),
	}
}

func (this *CliRpcServer) Start(port uint) {
	this.port = port
	this.httpSvtMux = http.NewServeMux()
	this.httpSvr = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
		Handler: this.httpSvtMux,
	}
	this.httpSvtMux.HandleFunc("/cli", this.Handler)
	err := this.httpSvr.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			return
		}
		panic(fmt.Sprintf("httpSvr.ListenAndServe error:%s", err))
	}
}

func (this *CliRpcServer) RegHandler(method string, handler func(req *common.CliRpcRequest, resp *common.CliRpcResponse)) {
	this.handlers[method] = handler
}

func (this *CliRpcServer) GetHandler(method string) func(req *common.CliRpcRequest, resp *common.CliRpcResponse) {
	handler, ok := this.handlers[method]
	if !ok {
		return nil
	}
	return handler
}

func (this *CliRpcServer) Handler(w http.ResponseWriter, r *http.Request) {
	resp := &common.CliRpcResponse{}
	defer func() {
		w.WriteHeader(http.StatusOK)

		if resp.ErrorInfo == "" {
			resp.ErrorInfo = common.GetCLIErrorDesc(resp.ErrorCode)
		}
		data, err := json.Marshal(resp)
		if err != nil {
			log.Error("CliRpcServer json.Marshal JsonRpcResponse:%+v error:%s", resp, err)
			return
		}
		_, err = w.Write(data)
		if err != nil {
			log.Error("CliRpcServer Write:%s error %s", data, err)
			return
		}
		log.Infof("[CliRpcResponse]%s", data)
	}()

	if r.Method != http.MethodPost {
		resp.ErrorCode = common.CLIERR_HTTP_METHOD_INVALID
		return
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("CliRpcServer read body error:%s", err)
		resp.ErrorCode = common.CLIERR_INVALID_REQUEST
		resp.ErrorInfo = "invalid body"
		return
	}
	defer r.Body.Close()

	log.Infof("[CliRpcRequest]%s", data)
	req := &common.CliRpcRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		log.Errorf("CliRpcServer json.Unmarshal JsonRpcRequest:%s error:%s", data, err)
		resp.ErrorCode = common.CLIERR_INVALID_PARAMS
		return
	}

	resp.Method = req.Method
	resp.Qid = req.Qid

	handler := this.GetHandler(req.Method)
	if handler == nil {
		resp.ErrorCode = common.CLIERR_UNSUPPORT_METHOD
		return
	}

	handler(req, resp)
}

func (this *CliRpcServer) Close() {
	err := this.httpSvr.Close()
	if err != nil {
		log.Error("httpSvr close error:%s", err)
	}
}
