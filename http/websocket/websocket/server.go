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
	"errors"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	Err "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/http/base/rest"
	"github.com/ontio/ontology/http/websocket/session"
)

type handler func(map[string]interface{}) map[string]interface{}
type Handler struct {
	handler  handler
	pushFlag bool
}

type WsServer struct {
	sync.RWMutex
	Upgrader     websocket.Upgrader
	listener     net.Listener
	server       *http.Server
	SessionList  *session.SessionList
	ActionMap    map[string]Handler
	TxHashMap    map[string]string //key: txHash   value:sessionid
	BroadcastMap map[string]string
}

func InitWsServer() *WsServer {
	ws := &WsServer{
		Upgrader:     websocket.Upgrader{},
		SessionList:  session.NewSessionList(),
		TxHashMap:    make(map[string]string),
		BroadcastMap: make(map[string]string),
	}
	return ws
}

func (this *WsServer) Start() error {
	if cfg.Parameters.HttpWsPort == 0 {
		log.Error("Not configure HttpWsPort port ")
		return nil
	}
	this.registryMethod()
	this.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	tlsFlag := false
	if tlsFlag || cfg.Parameters.HttpWsPort%1000 == rest.TLS_PORT {
		var err error
		this.listener, err = this.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		this.listener, err = net.Listen("tcp", ":"+strconv.Itoa(cfg.Parameters.HttpWsPort))
		if err != nil {
			log.Fatal("net.Listen: ", err.Error())
			return err
		}
	}
	var done = make(chan bool)
	go this.checkSessionsTimeout(done)

	this.server = &http.Server{Handler: http.HandlerFunc(this.webSocketHandler)}
	err := this.server.Serve(this.listener)

	done <- true
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
		return err
	}
	return nil

}

func (this *WsServer) registryMethod() {
	gettxhashmap := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		this.Lock()
		defer this.Unlock()
		resp["Result"] = len(this.TxHashMap)
		return resp
	}
	sendRawTransaction := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.SendRawTransaction(cmd)
		if userid, ok := resp["Userid"].(string); ok && len(userid) > 0 {
			if result, ok := resp["Result"].(string); ok {
				this.SetTxHashMap(result, userid)
			}
			delete(resp, "Userid")
		}
		return resp
	}
	heartbeat := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		if b, ok := cmd["Broadcast"].(bool); ok && b == true {
			if userid, ok := cmd["Userid"].(string); ok {
				this.BroadcastMap[userid] = userid
			}
		}
		resp["Action"] = "heartbeat"
		resp["Result"] = cmd["Userid"]
		return resp
	}
	getsessioncount := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.ResponsePack(Err.SUCCESS)
		resp["Action"] = "getsessioncount"
		resp["Result"] = this.SessionList.GetSessionCount()
		return resp
	}
	actionMap := map[string]Handler{
		"getconnectioncount": {handler: rest.GetConnectionCount},
		"getblockbyheight":   {handler: rest.GetBlockByHeight},
		"getblockbyhash":     {handler: rest.GetBlockByHash},
		"getblockheight":     {handler: rest.GetBlockHeight},
		"gettransaction":     {handler: rest.GetTransactionByHash},
		"sendrawtransaction": {handler: sendRawTransaction},
		"heartbeat":          {handler: heartbeat},

		"gettxhashmap":    {handler: gettxhashmap},
		"getsessioncount": {handler: getsessioncount},
	}
	this.ActionMap = actionMap
}

func (this *WsServer) Stop() {
	if this.server != nil {
		this.server.Shutdown(context.Background())
		log.Error("Close websocket ")
	}
}
func (this *WsServer) Restart() {
	go func() {
		time.Sleep(time.Second)
		this.Stop()
		time.Sleep(time.Second)
		go this.Start()
	}()
}

func (this *WsServer) checkSessionsTimeout(done chan bool) {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			var closeList []*session.Session
			this.SessionList.ForEachSession(func(v *session.Session) {
				if v.SessionTimeoverCheck() {
					resp := rest.ResponsePack(Err.SESSION_EXPIRED)
					this.response(v.GetSessionId(), resp)
					closeList = append(closeList, v)
				}
			})
			for _, s := range closeList {
				this.SessionList.CloseSession(s)
			}

		case <-done:
			return
		}
	}

}

func (this *WsServer) webSocketHandler(w http.ResponseWriter, r *http.Request) {
	wsConn, err := this.Upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Error("websocket Upgrader: ", err)
		return
	}
	defer wsConn.Close()
	nsSession, err := this.SessionList.NewSession(wsConn)
	if err != nil {
		log.Error("websocket NewSession:", err)
		return
	}

	defer func() {
		this.deleteTxHashs(nsSession.GetSessionId())
		this.SessionList.CloseSession(nsSession)
		if err := recover(); err != nil {
			log.Fatal("websocket recover:", err)
		}
	}()

	for {
		_, bysMsg, err := wsConn.ReadMessage()
		if err == nil {
			if this.OnDataHandle(nsSession, bysMsg, r) {
				nsSession.UpdateActiveTime()
			}
			continue
		}
		e, ok := err.(net.Error)
		if !ok || !e.Timeout() {
			log.Error("websocket conn:", err)
			return
		}
	}
}
func (this *WsServer) IsValidMsg(reqMsg map[string]interface{}) bool {
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
func (this *WsServer) OnDataHandle(curSession *session.Session, bysMsg []byte, r *http.Request) bool {

	var req = make(map[string]interface{})

	if err := json.Unmarshal(bysMsg, &req); err != nil {
		resp := rest.ResponsePack(Err.ILLEGAL_DATAFORMAT)
		this.response(curSession.GetSessionId(), resp)
		log.Error("websocket OnDataHandle:", err)
		return false
	}
	actionName, ok := req["Action"].(string)
	if !ok {
		resp := rest.ResponsePack(Err.INVALID_METHOD)
		this.response(curSession.GetSessionId(), resp)
		return false
	}
	action, ok := this.ActionMap[actionName]
	if !ok {
		resp := rest.ResponsePack(Err.INVALID_METHOD)
		this.response(curSession.GetSessionId(), resp)
		return false
	}
	if !this.IsValidMsg(req) {
		resp := rest.ResponsePack(Err.INVALID_PARAMS)
		this.response(curSession.GetSessionId(), resp)
		return true
	}
	if height, ok := req["Height"].(float64); ok {
		req["Height"] = strconv.FormatInt(int64(height), 10)
	}
	if raw, ok := req["Raw"].(float64); ok {
		req["Raw"] = strconv.FormatInt(int64(raw), 10)
	}

	req["Userid"] = curSession.GetSessionId()
	resp := action.handler(req)
	resp["Action"] = actionName
	if txHash, ok := resp["Result"].(string); ok && action.pushFlag {
		this.Lock()
		defer this.Unlock()
		this.TxHashMap[txHash] = curSession.GetSessionId()
	}
	this.response(curSession.GetSessionId(), resp)

	return true
}
func (this *WsServer) SetTxHashMap(txhash string, sessionid string) {
	this.Lock()
	defer this.Unlock()
	this.TxHashMap[txhash] = sessionid
}
func (this *WsServer) deleteTxHashs(sSessionId string) {
	this.Lock()
	defer this.Unlock()
	for k, v := range this.TxHashMap {
		if v == sSessionId {
			delete(this.TxHashMap, k)
		}
	}
}
func (this *WsServer) response(sSessionId string, resp map[string]interface{}) {
	resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error("Websocket response:", err)
		return
	}
	this.send(sSessionId, data)
}
func (this *WsServer) PushTxResult(txHashStr string, resp map[string]interface{}) {
	this.Lock()
	defer this.Unlock()
	sSessionId := this.TxHashMap[txHashStr]
	delete(this.TxHashMap, txHashStr)
	if len(sSessionId) > 0 {
		this.response(sSessionId, resp)
	}
	//broadcast BroadcastMap
	for _, v := range this.BroadcastMap {
		this.response(v, resp)
	}
}

func (this *WsServer) BroadcastResult(resp map[string]interface{}) {
	resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
	data, err := json.Marshal(resp)
	if err != nil {
		log.Error("Websocket PushResult:", err)
		return
	}
	this.broadcast(data)
}
func (this *WsServer) send(sSessionId string, data []byte) error {
	session := this.SessionList.GetSessionById(sSessionId)
	if session == nil {
		return errors.New("websocket sessionId Not Exist:" + sSessionId)
	}
	return session.Send(data)
}
func (this *WsServer) broadcast(data []byte) error {
	this.SessionList.ForEachSession(func(v *session.Session) {
		v.Send(data)
	})
	return nil
}

func (this *WsServer) initTlsListen() (net.Listener, error) {

	certPath := cfg.Parameters.RestCertPath
	keyPath := cfg.Parameters.RestKeyPath

	// load cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	log.Info("TLS listen port is ", strconv.Itoa(cfg.Parameters.HttpWsPort))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(cfg.Parameters.HttpWsPort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
