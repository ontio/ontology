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
	"context"
	"crypto/tls"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ontio/ontology/common"
	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	Err "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/http/base/rest"
	"github.com/ontio/ontology/http/websocket/session"
)

const (
	WSTOPIC_EVENT      = 1
	WSTOPIC_JSON_BLOCK = 2
	WSTOPIC_RAW_BLOCK  = 3
	WSTOPIC_TXHASHS    = 4
)

type handler func(map[string]interface{}) map[string]interface{}
type Handler struct {
	handler  handler
	pushFlag bool
}
type subscribe struct {
	ConstractsFilter      []string `json:"ConstractsFilter"`
	SubscribeEvent        bool     `json:"SubscribeEvent"`
	SubscribeJsonBlock    bool     `json:"SubscribeJsonBlock"`
	SubscribeRawBlock     bool     `json:"SubscribeRawBlock"`
	SubscribeBlockTxHashs bool     `json:"SubscribeBlockTxHashs"`
}
type WsServer struct {
	sync.RWMutex
	Upgrader     websocket.Upgrader
	listener     net.Listener
	server       *http.Server
	SessionList  *session.SessionList
	ActionMap    map[string]Handler
	TxHashMap    map[string]string //key: txHash   value:sessionid
	SubscribeMap map[string]subscribe
}

func InitWsServer() *WsServer {
	ws := &WsServer{
		Upgrader:     websocket.Upgrader{},
		SessionList:  session.NewSessionList(),
		TxHashMap:    make(map[string]string),
		SubscribeMap: make(map[string]subscribe),
	}
	return ws
}

func (self *WsServer) Start() error {
	wsPort := int(cfg.DefConfig.Ws.HttpWsPort)
	if wsPort == 0 {
		log.Error("Not configure HttpWsPort port ")
		return nil
	}
	self.registryMethod()
	self.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	tlsFlag := false
	if tlsFlag || wsPort%1000 == rest.TLS_PORT {
		var err error
		self.listener, err = self.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		self.listener, err = net.Listen("tcp", ":"+strconv.Itoa(wsPort))
		if err != nil {
			log.Fatal("net.Listen: ", err.Error())
			return err
		}
	}
	var done = make(chan bool)
	go self.checkSessionsTimeout(done)

	self.server = &http.Server{Handler: http.HandlerFunc(self.webSocketHandler)}
	err := self.server.Serve(self.listener)

	done <- true
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
		return err
	}
	return nil

}

func (self *WsServer) registryMethod() {

	heartbeat := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		self.Lock()
		defer self.Unlock()

		sessionId, _ := cmd["SessionId"].(string)
		sub := self.SubscribeMap[sessionId]
		resp["Action"] = "heartbeat"
		resp["Result"] = sub
		return resp
	}
	subscribe := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		self.Lock()
		defer self.Unlock()

		sessionId, _ := cmd["SessionId"].(string)
		sub := self.SubscribeMap[sessionId]
		if b, ok := cmd["SubscribeEvent"].(bool); ok {
			sub.SubscribeEvent = b
		}
		if b, ok := cmd["SubscribeJsonBlock"].(bool); ok {
			sub.SubscribeJsonBlock = b
		}
		if b, ok := cmd["SubscribeRawBlock"].(bool); ok {
			sub.SubscribeRawBlock = b
		}
		if b, ok := cmd["SubscribeBlockTxHashs"].(bool); ok {
			sub.SubscribeBlockTxHashs = b
		}
		if ctsf, ok := cmd["ConstractsFilter"].([]interface{}); ok {
			sub.ConstractsFilter = []string{}
			for _, v := range ctsf {
				if addr, k := v.(string); k {
					sub.ConstractsFilter = append(sub.ConstractsFilter, addr)
				}
			}
		}
		self.SubscribeMap[sessionId] = sub

		resp["Action"] = "subscribe"
		resp["Result"] = sub
		return resp
	}
	getsessioncount := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Action"] = "getsessioncount"
		resp["Result"] = self.SessionList.GetSessionCount()
		return resp
	}
	actionMap := map[string]Handler{
		"getblockheightbytxhash":    {handler: rest.GetBlockHeightByTxHash},
		"getsmartcodeeventbyhash":   {handler: rest.GetSmartCodeEventByTxHash},
		"getsmartcodeeventbyheight": {handler: rest.GetSmartCodeEventTxsByHeight},
		"getcontract":               {handler: rest.GetContractState},
		"getbalance":                {handler: rest.GetBalance},
		"getconnectioncount":        {handler: rest.GetConnectionCount},
		"getblockbyheight":          {handler: rest.GetBlockByHeight},
		"getblockhash":              {handler: rest.GetBlockHash},
		"getblockbyhash":            {handler: rest.GetBlockByHash},
		"getblockheight":            {handler: rest.GetBlockHeight},
		"getgenerateblocktime":      {handler: rest.GetGenerateBlockTime},
		"gettransaction":            {handler: rest.GetTransactionByHash},
		"sendrawtransaction":        {handler: rest.SendRawTransaction, pushFlag: true},
		"heartbeat":                 {handler: heartbeat},
		"subscribe":                 {handler: subscribe},
		"getstorage":                {handler: rest.GetStorage},
		"getallowance":              {handler: rest.GetAllowance},
		"getmerkleproof":            {handler: rest.GetMerkleProof},
		"getblocktxsbyheight":       {handler: rest.GetBlockTxsByHeight},
		"getgasprice":               {handler: rest.GetGasPrice},
		"getunclaimong":             {handler: rest.GetUnclaimOng},
		"getmempooltxcount":         {handler: rest.GetMemPoolTxCount},
		"getmempooltxstate":         {handler: rest.GetMemPoolTxState},

		"getsessioncount": {handler: getsessioncount},
	}
	self.ActionMap = actionMap
}

func (self *WsServer) Stop() {
	if self.server != nil {
		self.server.Shutdown(context.Background())
		log.Infof("Close websocket ")
	}
}
func (self *WsServer) Restart() {
	go func() {
		time.Sleep(time.Second)
		self.Stop()
		time.Sleep(time.Second)
		go self.Start()
	}()
}

func (self *WsServer) checkSessionsTimeout(done chan bool) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var closeList []*session.Session
			self.SessionList.ForEachSession(func(v *session.Session) {
				if v.SessionTimeoverCheck() {
					resp := rest.ResponsePack(Err.SESSION_EXPIRED)
					v.Send(marshalResp(resp))
					closeList = append(closeList, v)
				}
			})
			for _, s := range closeList {
				self.SessionList.CloseSession(s)
			}

		case <-done:
			return
		}
	}

}

func (self *WsServer) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := self.Upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("websocket Upgrader: ", err)
		return
	}
	defer wsConn.Close()
	nsSession, err := self.SessionList.NewSession(wsConn)
	if err != nil {
		log.Error("websocket NewSession:", err)
		return
	}

	defer func() {
		self.deleteTxHashes(nsSession.GetSessionId())
		self.deleteSubscribe(nsSession.GetSessionId())
		self.SessionList.CloseSession(nsSession)
		if err := recover(); err != nil {
			log.Fatal("websocket recover:", err)
		}
	}()

	for {
		_, bysMsg, err := wsConn.ReadMessage()
		if err == nil {
			if self.OnDataHandle(nsSession, bysMsg, r) {
				nsSession.UpdateActiveTime()
			}
			continue
		}
		e, ok := err.(net.Error)
		if !ok || !e.Timeout() {
			log.Infof("websocket conn:", err)
			return
		}
	}
}
func (self *WsServer) IsValidMsg(reqMsg map[string]interface{}) bool {
	if _, ok := reqMsg["Hash"].(string); !ok && reqMsg["Hash"] != nil {
		return false
	}
	if _, ok := reqMsg["Addr"].(string); !ok && reqMsg["Addr"] != nil {
		return false
	}
	if _, ok := reqMsg["Assetid"].(string); !ok && reqMsg["Assetid"] != nil {
		return false
	}
	return true
}
func (self *WsServer) OnDataHandle(curSession *session.Session, bysMsg []byte, r *http.Request) bool {

	var req = make(map[string]interface{})

	if err := json.Unmarshal(bysMsg, &req); err != nil {
		resp := rest.ResponsePack(Err.ILLEGAL_DATAFORMAT)
		curSession.Send(marshalResp(resp))
		log.Infof("websocket OnDataHandle:", err)
		return false
	}
	actionName, ok := req["Action"].(string)
	if !ok {
		resp := rest.ResponsePack(Err.INVALID_METHOD)
		curSession.Send(marshalResp(resp))
		return false
	}
	action, ok := self.ActionMap[actionName]
	if !ok {
		resp := rest.ResponsePack(Err.INVALID_METHOD)
		curSession.Send(marshalResp(resp))
		return false
	}
	if !self.IsValidMsg(req) {
		resp := rest.ResponsePack(Err.INVALID_PARAMS)
		curSession.Send(marshalResp(resp))
		return true
	}
	if height, ok := req["Height"].(float64); ok {
		req["Height"] = strconv.FormatInt(int64(height), 10)
	}
	if raw, ok := req["Raw"].(float64); ok {
		req["Raw"] = strconv.FormatInt(int64(raw), 10)
	}
	req["SessionId"] = curSession.GetSessionId()
	resp := action.handler(req)
	resp["Action"] = actionName
	resp["Id"] = req["Id"]
	if action.pushFlag {
		if error, _ := resp["Error"].(int64); ok && error == 0 {
			if txHash, ok := resp["Result"].(string); ok && len(txHash) == common.UINT256_SIZE*2 {
				//delete timeover txhashs
				txhashs := curSession.RemoveTimeoverTxHashes()
				self.Lock()
				defer self.Unlock()
				for _, v := range txhashs {
					delete(self.TxHashMap, v.TxHash)
				}
				//add new txhash
				self.TxHashMap[txHash] = curSession.GetSessionId()
				curSession.AppendTxHash(txHash)
			}
		}
	}
	curSession.Send(marshalResp(resp))

	return true
}
func (self *WsServer) InsertTxHashMap(txhash string, sessionid string) {
	self.Lock()
	defer self.Unlock()
	self.TxHashMap[txhash] = sessionid
}
func (self *WsServer) deleteTxHashes(sessionId string) {
	self.Lock()
	defer self.Unlock()
	for k, v := range self.TxHashMap {
		if v == sessionId {
			delete(self.TxHashMap, k)
		}
	}
}
func (self *WsServer) deleteSubscribe(sessionId string) {
	self.Lock()
	defer self.Unlock()
	delete(self.SubscribeMap, sessionId)
}

func marshalResp(resp map[string]interface{}) []byte {
	resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
	data, err := json.Marshal(resp)
	if err != nil {
		log.Infof("Websocket marshal json error:", err)
		return nil
	}

	return data
}

func (self *WsServer) PushTxResult(contractAddrs map[string]bool, txHashStr string, resp map[string]interface{}) {
	self.Lock()
	sessionId := self.TxHashMap[txHashStr]
	delete(self.TxHashMap, txHashStr)
	//avoid twice, will send in BroadcastToSubscribers
	sub := self.SubscribeMap[sessionId]
	if sub.SubscribeEvent {
		if len(sub.ConstractsFilter) == 0 {
			self.Unlock()
			return
		}
		for _, addr := range sub.ConstractsFilter {
			if contractAddrs[addr] {
				self.Unlock()
				return
			}
		}
	}
	self.Unlock()

	s := self.SessionList.GetSessionById(sessionId)
	if s != nil {
		s.Send(marshalResp(resp))
	}
}
func (self *WsServer) BroadcastToSubscribers(contractAddrs map[string]bool, sub int, resp map[string]interface{}) {
	// broadcast SubscribeMap
	self.Lock()
	defer self.Unlock()
	data := marshalResp(resp)
	for sid, v := range self.SubscribeMap {
		s := self.SessionList.GetSessionById(sid)
		if s == nil {
			continue
		}
		if sub == WSTOPIC_JSON_BLOCK && v.SubscribeJsonBlock {
			s.Send(data)
		} else if sub == WSTOPIC_RAW_BLOCK && v.SubscribeRawBlock {
			s.Send(data)
		} else if sub == WSTOPIC_TXHASHS && v.SubscribeBlockTxHashs {
			s.Send(data)
		} else if sub == WSTOPIC_EVENT && v.SubscribeEvent {
			if len(v.ConstractsFilter) == 0 {
				s.Send(data)
				continue
			}
			for _, addr := range v.ConstractsFilter {
				if contractAddrs[addr] {
					s.Send(data)
					break
				}
			}
		}
	}
}

func (self *WsServer) initTlsListen() (net.Listener, error) {

	certPath := cfg.DefConfig.Ws.HttpCertPath
	keyPath := cfg.DefConfig.Ws.HttpKeyPath

	// load cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	wsPort := strconv.Itoa(int(cfg.DefConfig.Ws.HttpWsPort))
	log.Info("TLS listen port is ", wsPort)
	listener, err := tls.Listen("tcp", ":"+wsPort, tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
