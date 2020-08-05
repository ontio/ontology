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
	"time"

	"github.com/ontio/ontology/common/log"
	bactor "github.com/ontio/ontology/http/base/actor"
	"github.com/ontio/ontology/http/base/common"
	berr "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/http/base/rpc"
)

func GetNeighbor(params []interface{}) map[string]interface{} {
	addr := bactor.GetNeighborAddrs()
	return rpc.ResponseSuccess(addr)
}

func GetNodeState(params []interface{}) map[string]interface{} {
	t := time.Now().UnixNano()
	port := bactor.GetNodePort()
	id := bactor.GetID()
	ver := bactor.GetVersion()
	tpe := bactor.GetNodeType()
	relay := bactor.GetRelayState()
	height := bactor.GetCurrentBlockHeight()
	txnCnt, err := bactor.GetTxnCount()
	if err != nil {
		return rpc.ResponsePack(berr.INTERNAL_ERROR, false)
	}
	n := common.NodeInfo{
		NodeTime:    t,
		NodePort:    port,
		ID:          id,
		NodeVersion: ver,
		NodeType:    tpe,
		Relay:       relay,
		Height:      height,
		TxnCnt:      txnCnt,
	}
	return rpc.ResponseSuccess(n)
}

func StartConsensus(params []interface{}) map[string]interface{} {
	if err := bactor.ConsensusSrvStart(); err != nil {
		return rpc.ResponsePack(berr.INTERNAL_ERROR, false)
	}
	return rpc.ResponsePack(berr.SUCCESS, true)
}

func StopConsensus(params []interface{}) map[string]interface{} {
	if err := bactor.ConsensusSrvHalt(); err != nil {
		return rpc.ResponsePack(berr.INTERNAL_ERROR, false)
	}
	return rpc.ResponsePack(berr.SUCCESS, true)
}

func SetDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return rpc.ResponsePack(berr.INVALID_PARAMS, "")
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log().SetDebugLevel(int(level)); err != nil {
			return rpc.ResponsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return rpc.ResponsePack(berr.INVALID_PARAMS, "")
	}
	return rpc.ResponsePack(berr.SUCCESS, true)
}
