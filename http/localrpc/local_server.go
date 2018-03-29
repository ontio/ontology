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

package localrpc

import (
	"net/http"
	"strconv"

	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/base/rpc"
)

const (
	localHost string = "127.0.0.1"
	LocalDir  string = "/local"
)

func StartLocalServer() {
	log.Debug()
	http.HandleFunc(LocalDir, Handle)

	HandleFunc("getneighbor", GetNeighbor)
	HandleFunc("getnodestate", GetNodeState)
	HandleFunc("startconsensus", StartConsensus)
	HandleFunc("stopconsensus", StopConsensus)
	//HandleFunc("sendsampletransaction", SendSampleTransaction)
	HandleFunc("setdebuginfo", SetDebugInfo)

	// TODO: only listen to local host
	err := http.ListenAndServe(":"+strconv.Itoa(Parameters.HttpLocalPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
