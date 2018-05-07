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
	"github.com/ontio/ontology/common"
	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/events/message"
	bactor "github.com/ontio/ontology/http/base/actor"
	bcomn "github.com/ontio/ontology/http/base/common"
	Err "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/http/base/rest"
	"github.com/ontio/ontology/http/websocket/websocket"
	"github.com/ontio/ontology/smartcontract/event"
)

var ws *websocket.WsServer

func StartServer() {
	bactor.SubscribeEvent(message.TOPIC_SAVE_BLOCK_COMPLETE, sendBlock2WSclient)
	bactor.SubscribeEvent(message.TOPIC_SMART_CODE_EVENT, pushSmartCodeEvent)
	go func() {
		ws = websocket.InitWsServer()
		ws.Start()
	}()
}
func sendBlock2WSclient(v interface{}) {
	if cfg.DefConfig.Ws.HttpWsPort != 0 {
		go func() {
			pushBlock(v)
			pushBlockTransactions(v)
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

func pushSmartCodeEvent(v interface{}) {
	if ws == nil {
		return
	}
	rs, ok := v.(types.SmartCodeEvent)
	if !ok {
		log.Errorf("[PushSmartCodeEvent]", "SmartCodeEvent err")
		return
	}
	go func() {
		switch object := rs.Result.(type) {
		case []*event.NotifyEventInfo:
			evts := []bcomn.NotifyEventInfo{}
			var contractAddrs = make(map[string]bool)
			for _, v := range object {
				txhash := v.TxHash
				evts = append(evts, bcomn.NotifyEventInfo{common.ToHexString(txhash[:]), v.ContractAddress.ToHexString(), v.States})
				contractAddrs[v.ContractAddress.ToHexString()] = true
			}
			pushEvent(contractAddrs, rs.TxHash, rs.Error, rs.Action, evts)
		case *event.LogEventArgs:
			type logEventArgs struct {
				TxHash          string
				ContractAddress string
				Message         string
			}
			hash := object.TxHash
			addr := object.ContractAddress.ToHexString()
			pushEvent(map[string]bool{addr:true}, rs.TxHash, rs.Error, rs.Action,
				logEventArgs{common.ToHexString(hash[:]), addr, object.Message})
		default:
		}
	}()
}

func pushEvent(contractAddrs map[string]bool, txHash string, errcode int64, action string, result interface{}) {
	if ws != nil {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Result"] = result
		resp["Error"] = errcode
		resp["Action"] = action
		resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
		ws.PushTxResult(contractAddrs,txHash, resp)
		ws.BroadcastToSubscribers(contractAddrs, websocket.WSTOPIC_EVENT, resp)
	}
}

func pushBlock(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(types.Block); ok {
		resp["Action"] = "sendrawblock"
		w := bytes.NewBuffer(nil)
		block.Serialize(w)
		resp["Result"] = common.ToHexString(w.Bytes())
		ws.BroadcastToSubscribers(nil,websocket.WSTOPIC_RAW_BLOCK, resp)

		resp["Action"] = "sendjsonblock"
		resp["Result"] = bcomn.GetBlockInfo(&block)
		ws.BroadcastToSubscribers(nil,websocket.WSTOPIC_JSON_BLOCK, resp)
	}
}
func pushBlockTransactions(v interface{}) {
	if ws == nil {
		return
	}
	resp := rest.ResponsePack(Err.SUCCESS)
	if block, ok := v.(types.Block); ok {
		resp["Result"] = rest.GetBlockTransactions(&block)
		resp["Action"] = "sendblocktxhashs"
		ws.BroadcastToSubscribers(nil,websocket.WSTOPIC_TXHASHS, resp)
	}
}
