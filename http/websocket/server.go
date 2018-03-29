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

package websocket

import (
	"bytes"

	"github.com/Ontology/common"
	cfg "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/events/message"
	bactor "github.com/Ontology/http/base/actor"
	bcomn "github.com/Ontology/http/base/common"
	Err "github.com/Ontology/http/base/error"
	"github.com/Ontology/http/base/rest"
	"github.com/Ontology/http/websocket/websocket"
)

var ws *websocket.WsServer
var (
	pushBlockFlag    bool = false
	pushRawBlockFlag bool = false
	pushBlockTxsFlag bool = false
)

func StartServer() {
	bactor.SubscribeEvent(message.TOPIC_SAVE_BLOCK_COMPLETE, SendBlock2WSclient)
	bactor.SubscribeEvent(message.TOPIC_SMART_CODE_EVENT, PushSmartCodeEvent)
	go func() {
		ws = websocket.InitWsServer()
		ws.Start()
	}()
}
func SendBlock2WSclient(v interface{}) {
	if cfg.Parameters.HttpWsPort != 0 && pushBlockFlag {
		go func() {
			PushBlock(v)
		}()
	}
	if cfg.Parameters.HttpWsPort != 0 && pushBlockTxsFlag {
		go func() {
			PushBlockTransactions(v)
		}()
	}
}
func Stop() {
	if ws == nil {
		return
	}
	ws.Stop()
}
func ReStartServer() {
	if ws == nil {
		ws = websocket.InitWsServer()
		ws.Start()
		return
	}
	ws.Restart()
}
func GetWsPushBlockFlag() bool {
	return pushBlockFlag
}
func SetWsPushBlockFlag(b bool) {
	pushBlockFlag = b
}
func GetPushRawBlockFlag() bool {
	return pushRawBlockFlag
}
func SetPushRawBlockFlag(b bool) {
	pushRawBlockFlag = b
}
func GetPushBlockTxsFlag() bool {
	return pushBlockTxsFlag
}
func SetPushBlockTxsFlag(b bool) {
	pushBlockTxsFlag = b
}
func SetTxHashMap(txhash string, sessionid string) {
	if ws == nil {
		return
	}
	ws.SetTxHashMap(txhash, sessionid)
}

func PushSmartCodeEvent(v interface{}) {
	if ws != nil {
		log.Info("[PushSmartCodeEvent]", v)
		rs, ok := v.(types.SmartCodeEvent)
		if !ok {
			log.Errorf("[PushSmartCodeEvent]", "SmartCodeEvent err")
			return
		}
		go func() {
			PushEvent(rs.TxHash, rs.Error, rs.Action, rs.Result)
		}()
	}
}

func PushEvent(txHash string, errcode int64, action string, result interface{}) {
	if ws != nil {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Result"] = result
		resp["Error"] = errcode
		resp["Action"] = action
		resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
		ws.PushTxResult(txHash, resp)
		//ws.BroadcastResult(resp)
	}
}

func PushBlock(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(*types.Block); ok {
		if pushRawBlockFlag {
			w := bytes.NewBuffer(nil)
			block.Serialize(w)
			resp["Result"] = common.ToHexString(w.Bytes())
		} else {
			resp["Result"] = bcomn.GetBlockInfo(block)
		}
		resp["Action"] = "sendrawblock"
		ws.BroadcastResult(resp)
	}
}
func PushBlockTransactions(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(*types.Block); ok {
		if pushBlockTxsFlag {
			resp["Result"] = rest.GetBlockTransactions(block)
		}
		resp["Action"] = "sendblocktransactions"
		ws.BroadcastResult(resp)
	}
}
