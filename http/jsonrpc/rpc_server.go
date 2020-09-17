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

func StartRPCServer() error {
	log.Debug()
	http.HandleFunc("/", rpc.Handle)
	rpc.HandleFunc("getbestblockhash", GetBestBlockHash)
	rpc.HandleFunc("getblock", GetBlock)
	rpc.HandleFunc("getblockcount", GetBlockCount)
	rpc.HandleFunc("getblockhash", GetBlockHash)
	rpc.HandleFunc("getconnectioncount", GetConnectionCount)
	rpc.HandleFunc("getsyncstatus", GetSyncStatus)
	//HandleFunc("getrawmempool", GetRawMemPool)

	rpc.HandleFunc("getrawtransaction", GetRawTransaction)
	rpc.HandleFunc("sendrawtransaction", SendRawTransaction)
	rpc.HandleFunc("getstorage", GetStorage)
	rpc.HandleFunc("getversion", GetNodeVersion)
	rpc.HandleFunc("getnetworkid", GetNetworkId)

	rpc.HandleFunc("getcontractstate", GetContractState)
	rpc.HandleFunc("getmempooltxcount", GetMemPoolTxCount)
	rpc.HandleFunc("getmempooltxstate", GetMemPoolTxState)
	rpc.HandleFunc("getmempooltxhashlist", GetMemPoolTxHashList)
	rpc.HandleFunc("getsmartcodeevent", GetSmartCodeEvent)
	rpc.HandleFunc("getblockheightbytxhash", GetBlockHeightByTxHash)

	rpc.HandleFunc("getbalance", GetBalance)
	rpc.HandleFunc("getoep4balance", GetOep4Balance)
	rpc.HandleFunc("getallowance", GetAllowance)
	rpc.HandleFunc("getmerkleproof", GetMerkleProof)
	rpc.HandleFunc("getblocktxsbyheight", GetBlockTxsByHeight)
	rpc.HandleFunc("getgasprice", GetGasPrice)
	rpc.HandleFunc("getunboundong", GetUnboundOng)
	rpc.HandleFunc("getgrantong", GetGrantOng)

	rpc.HandleFunc("getcrosschainmsg", GetCrossChainMsg)
	rpc.HandleFunc("getcrossstatesproof", GetCrossStatesProof)

	err := http.ListenAndServe(":"+strconv.Itoa(int(cfg.DefConfig.Rpc.HttpJsonPort)), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
