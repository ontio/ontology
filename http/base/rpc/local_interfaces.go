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

package rpc

import (
	"os"
	"path/filepath"

	"github.com/ontio/ontology/common/log"
	bactor "github.com/ontio/ontology/http/base/actor"
	"github.com/ontio/ontology/http/base/common"
	berr "github.com/ontio/ontology/http/base/error"
)

const (
	RANDBYTELEN = 4
)

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir
}

func GetNeighbor(params []interface{}) map[string]interface{} {
	addr := bactor.GetNeighborAddrs()
	return responseSuccess(addr)
}

func GetNodeState(params []interface{}) map[string]interface{} {
	state, err := bactor.GetConnectionState()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	t, err := bactor.GetNodeTime()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	port, err := bactor.GetNodePort()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	id, err := bactor.GetID()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	ver, err := bactor.GetVersion()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	tpe, err := bactor.GetNodeType()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	relay, err := bactor.GetRelayState()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	height := bactor.GetCurrentBlockHeight()
	txnCnt, err := bactor.GetTxnCnt()
	if err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	n := common.NodeInfo{
		NodeState:   uint(state),
		NodeTime:    t,
		NodePort:    port,
		ID:          id,
		NodeVersion: ver,
		NodeType:    tpe,
		Relay:       relay,
		Height:      height,
		TxnCnt:      txnCnt,
	}
	return responseSuccess(n)
}

func StartConsensus(params []interface{}) map[string]interface{} {
	if err := bactor.ConsensusSrvStart(); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responsePack(berr.SUCCESS, true)
}

func StopConsensus(params []interface{}) map[string]interface{} {
	if err := bactor.ConsensusSrvHalt(); err != nil {
		return responsePack(berr.INTERNAL_ERROR, false)
	}
	return responsePack(berr.SUCCESS, true)
}

func SetDebugInfo(params []interface{}) map[string]interface{} {
	if len(params) < 1 {
		return responsePack(berr.INVALID_PARAMS, "")
	}
	switch params[0].(type) {
	case float64:
		level := params[0].(float64)
		if err := log.Log.SetDebugLevel(int(level)); err != nil {
			return responsePack(berr.INVALID_PARAMS, "")
		}
	default:
		return responsePack(berr.INVALID_PARAMS, "")
	}
	return responsePack(berr.SUCCESS, true)
}
