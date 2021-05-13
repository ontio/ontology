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

	"github.com/ethereum/go-ethereum/rpc"
	cfg "github.com/ontio/ontology/common/config"
)

func StartEthServer() error {
	calculator := new(EthereumAPI)
	server := rpc.NewServer()
	err := server.RegisterName("eth", calculator)
	if err != nil {
		return err
	}
	netRpcService := new(PublicNetAPI)
	err = server.RegisterName("net", netRpcService)
	if err != nil {
		return err
	}
	err = http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.EthJsonPort)), server)
	if err != nil {
		return err
	}
	return nil
}
