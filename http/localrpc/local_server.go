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

// Package localrpc privides a function to start local rpc server
package localrpc

import (
	"fmt"
	"net/http"
	"strconv"

	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/http/jsonrpc"
)

const (
	LOCAL_HOST string = "127.0.0.1"
)

var LocalRpcMux = jsonrpc.NewRPCHandler()

func StartLocalServer() error {
	log.Debug()
	rpcMux := LocalRpcMux

	rpcMux.HandleFunc("getneighbor", GetNeighbor)
	rpcMux.HandleFunc("getnodestate", GetNodeState)
	rpcMux.HandleFunc("startconsensus", StartConsensus)
	rpcMux.HandleFunc("stopconsensus", StopConsensus)
	rpcMux.HandleFunc("setdebuginfo", SetDebugInfo)

	mux := http.NewServeMux()
	mux.Handle("/", rpcMux)
	err := http.ListenAndServe(LOCAL_HOST+":"+strconv.Itoa(int(cfg.DefConfig.Rpc.HttpLocalPort)), mux)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
