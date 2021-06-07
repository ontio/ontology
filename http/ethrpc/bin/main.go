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
package main

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ontio/ontology/http/ethrpc"
)

func main() {

	startEthRpc()
}

func Ensure(err error) {
	if err != nil {
		panic(err)
	}
}

func startEthRpc() {
	calculator := new(ethrpc.EthereumAPI)
	server := rpc.NewServer()
	err := server.RegisterName("eth", calculator)
	Ensure(err)
	netRpcService := new(ethrpc.PublicNetAPI)
	err = server.RegisterName("net", netRpcService)
	Ensure(err)
	fmt.Printf("listen on 8545")
	err = http.ListenAndServe("0.0.0.0:8545", server)
	Ensure(err)
}
