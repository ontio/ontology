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

// Package jsonrpc privides a function to start json rpc server
package jsonrpc

import (
	"fmt"
	"net/http"
	"strconv"

	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/http/base/rpc"
)

func NewRPCHandler() *rpc.ServeMux {
	mux := rpc.NewServeMux()
	mux.HandleFunc("getbestblockhash", GetBestBlockHash)
	mux.HandleFunc("getblock", GetBlock)
	mux.HandleFunc("getblockcount", GetBlockCount)
	mux.HandleFunc("getblockhash", GetBlockHash)
	mux.HandleFunc("getconnectioncount", GetConnectionCount)
	mux.HandleFunc("getsyncstatus", GetSyncStatus)
	//HandleFunc("getrawmempool", GetRawMemPool)

	mux.HandleFunc("getrawtransaction", GetRawTransaction)
	mux.HandleFunc("sendrawtransaction", SendRawTransaction)
	mux.HandleFunc("getstorage", GetStorage)
	mux.HandleFunc("getversion", GetNodeVersion)
	mux.HandleFunc("getnetworkid", GetNetworkId)

	mux.HandleFunc("getcontractstate", GetContractState)
	mux.HandleFunc("getmempooltxcount", GetMemPoolTxCount)
	mux.HandleFunc("getmempooltxstate", GetMemPoolTxState)
	mux.HandleFunc("getmempooltxhashlist", GetMemPoolTxHashList)
	mux.HandleFunc("getsmartcodeevent", GetSmartCodeEvent)
	mux.HandleFunc("getblockheightbytxhash", GetBlockHeightByTxHash)

	mux.HandleFunc("getbalance", GetBalance)
	mux.HandleFunc("getbalancev2", GetBalanceV2)
	mux.HandleFunc("getoep4balance", GetOep4Balance)
	mux.HandleFunc("getallowance", GetAllowance)
	mux.HandleFunc("getallowancev2", GetAllowanceV2)
	mux.HandleFunc("getmerkleproof", GetMerkleProof)
	mux.HandleFunc("getblocktxsbyheight", GetBlockTxsByHeight)
	mux.HandleFunc("getgasprice", GetGasPrice)
	mux.HandleFunc("getunboundong", GetUnboundOng)
	mux.HandleFunc("getgrantong", GetGrantOng)

	mux.HandleFunc("getcrosschainmsg", GetCrossChainMsg)
	mux.HandleFunc("getcrossstatesproof", GetCrossStatesProof)
	mux.HandleFunc("getcrossstatesleafhashes", GetCrossStatesLeafHashes)

	return mux
}

func StartRPCServer() error {
	log.Debug()

	rpcMux := NewRPCHandler()
	mux := http.NewServeMux()
	mux.Handle("/", rpcMux)
	err := http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.HttpJsonPort)), mux)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
