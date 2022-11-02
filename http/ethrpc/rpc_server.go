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
package ethrpc

import (
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/http/base/actor"
	backend2 "github.com/ontio/ontology/http/ethrpc/backend"
	"github.com/ontio/ontology/http/ethrpc/eth"
	filters2 "github.com/ontio/ontology/http/ethrpc/filters"
	"github.com/ontio/ontology/http/ethrpc/net"
	"github.com/ontio/ontology/http/ethrpc/utils"
	"github.com/ontio/ontology/http/ethrpc/web3"
	tp "github.com/ontio/ontology/txnpool/proc"
)

var (
	vhosts = []string{
		"*",
	}
	// just like 20336/20334's Access-Control-Allow-Origin
	cors = []string{
		"*",
	}
)

func StartEthServer(txpool *tp.TXPoolServer) error {
	log.Root().SetHandler(utils.OntLogHandler())
	server := rpc.NewServer()
	if err := server.RegisterName("eth", eth.NewEthereumAPI(txpool)); err != nil {
		return err
	}

	backend := backend2.NewBloomBackend()
	err := backend.StartBloomHandlers(ledgerstore.BloomBitsBlocks, actor.GetIndexStore())
	if err != nil {
		return err
	}

	if err := server.RegisterName("eth", filters2.NewPublicFilterAPI(backend)); err != nil {
		return err
	}
	if err := server.RegisterName("net", net.NewPublicNetAPI()); err != nil {
		return err
	}
	if err := server.RegisterName("web3", web3.NewAPI()); err != nil {
		return err
	}

	// add cors wrapper
	wrappedCORSHandler := node.NewHTTPHandlerStack(server, cors, vhosts)

	err = http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.EthJsonPort)), wrappedCORSHandler)
	if err != nil {
		return err
	}

	return nil
}
