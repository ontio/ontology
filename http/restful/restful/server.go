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

package restful

import (
	cfg "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/http/base/rest"
	berr "github.com/Ontology/http/base/error"
	"github.com/Ontology/http/websocket"
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type handler func(map[string]interface{}) map[string]interface{}
type Action struct {
	sync.RWMutex
	name    string
	handler handler
}
type restServer struct {
	router           *Router
	listener         net.Listener
	server           *http.Server
	postMap          map[string]Action
	getMap           map[string]Action
}

const (
	Api_GetGenBlockTime = "/api/v1/node/generateblocktime"
	Api_GetconnCount  = "/api/v1/node/connectioncount"
	Api_GetblkTxsByHeight = "/api/v1/block/transactions/height/:height"
	Api_Getblkbyheight = "/api/v1/block/details/height/:height"
	Api_Getblkbyhash = "/api/v1/block/details/hash/:hash"
	Api_Getblkheight = "/api/v1/block/height"
	Api_Getblkhash = "/api/v1/block/hash/:height"
	Api_GetTransaction = "/api/v1/transaction/:hash"
	Api_SendRawTx = "/api/v1/transaction"
	Api_GetStorage = "/api/v1/storage/:hash/:key"
	Api_GetBalanceByAddr = "/api/v1/balance/:addr"
	Api_WebsocketState       = "/api/v1/config/websocket/state"
	Api_Restart              = "/api/v1/restart"
	Api_GetContractState     = "/api/v1/contract/:hash"
	Api_GetSmtCodeEvtByHgt  = "/api/v1/smartcode/event/height/:height"
	Api_GetSmtCodeEvtByHash = "/api/v1/smartcode/event/txhash/:hash"
	Api_GetBlkHeightByTxHash = "/api/v1/block/height/txhash/:hash"
)

func InitRestServer() rest.ApiServer {
	rt := &restServer{}

	rt.router = NewRouter()
	rt.registryMethod()
	rt.initGetHandler()
	rt.initPostHandler()
	return rt
}

func (this *restServer) Start() error {
	if cfg.Parameters.HttpRestPort == 0 {
		log.Fatal("Not configure HttpRestPort port ")
		return nil
	}

	tlsFlag := false
	if tlsFlag || cfg.Parameters.HttpRestPort % 1000 == rest.TlsPort {
		var err error
		this.listener, err = this.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		this.listener, err = net.Listen("tcp", ":" + strconv.Itoa(cfg.Parameters.HttpRestPort))
		if err != nil {
			log.Fatal("net.Listen: ", err.Error())
			return err
		}
	}
	this.server = &http.Server{Handler: this.router}
	err := this.server.Serve(this.listener)

	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
		return err
	}

	return nil
}
func (this *restServer) setWebsocketState(cmd map[string]interface{}) map[string]interface{} {
	resp := rest.ResponsePack(berr.SUCCESS)
	startFlag, ok := cmd["Open"].(bool)
	if !ok {
		resp["Error"] = berr.INVALID_PARAMS
		return resp
	}
	if b, ok := cmd["PushBlock"].(bool); ok {
		websocket.SetWsPushBlockFlag(b)
	}
	if b, ok := cmd["PushRawBlock"].(bool); ok {
		websocket.SetPushRawBlockFlag(b)
	}
	if b, ok := cmd["PushBlockTxs"].(bool); ok {
		websocket.SetPushBlockTxsFlag(b)
	}
	if wsPort, ok := cmd["Port"].(float64); ok && wsPort != 0 {
		cfg.Parameters.HttpWsPort = int(wsPort)
	}
	if startFlag {
		websocket.ReStartServer()
	} else {
		websocket.Stop()
	}
	var result = make(map[string]interface{})
	result["Open"] = startFlag
	result["Port"] = cfg.Parameters.HttpWsPort
	result["PushBlock"] = websocket.GetWsPushBlockFlag()
	result["PushRawBlock"] = websocket.GetPushRawBlockFlag()
	result["PushBlockTxs"] = websocket.GetPushBlockTxsFlag()
	resp["Result"] = result
	return resp
}
func (this *restServer) registryMethod() {

	getMethodMap := map[string]Action{
		Api_GetGenBlockTime:  {name: "getgenerateblocktime", handler: rest.GetGenerateBlockTime},
		Api_GetconnCount:  {name: "getconnectioncount", handler: rest.GetConnectionCount},
		Api_GetblkTxsByHeight: {name: "getblocktxsbyheight", handler: rest.GetBlockTxsByHeight},
		Api_Getblkbyheight:    {name: "getblockbyheight", handler: rest.GetBlockByHeight},
		Api_Getblkbyhash:         {name: "getblockbyhash", handler: rest.GetBlockByHash},
		Api_Getblkheight:         {name: "getblockheight", handler: rest.GetBlockHeight},
		Api_Getblkhash:           {name: "getblockhash", handler: rest.GetBlockHash},
		Api_GetTransaction:       {name: "gettransaction", handler: rest.GetTransactionByHash},
		Api_GetContractState:     {name: "getcontract", handler: rest.GetContractState},
		Api_Restart:              {name: "restart", handler: this.Restart},
		Api_GetSmtCodeEvtByHgt:    {name: "getsmartcodeeventbyheight", handler: rest.GetSmartCodeEventByHeight},
		Api_GetSmtCodeEvtByHash:    {name: "getsmartcodeeventbyhash", handler: rest.GetSmartCodeEventByTxHash},
		Api_GetBlkHeightByTxHash: {name: "getblockheightbytxhash", handler: rest.GetBlockHeightByTxHash},
		Api_GetStorage:           {name: "getstorage", handler: rest.GetStorage},
		Api_GetBalanceByAddr:    {name: "getbalance", handler: rest.GetBalance},
	}

	sendRawTransaction := func(cmd map[string]interface{}) map[string]interface{} {
		resp := rest.SendRawTransaction(cmd)
		if userid, ok := resp["Userid"].(string); ok && len(userid) > 0 {
			if result, ok := resp["Result"].(string); ok {
				websocket.SetTxHashMap(result, userid)
			}
			delete(resp, "Userid")
		}
		return resp
	}
	postMethodMap := map[string]Action{
		Api_SendRawTx:          {name: "sendrawtransaction", handler: sendRawTransaction},
		Api_WebsocketState:     {name: "setwebsocketstate", handler: this.setWebsocketState},
	}
	this.postMap = postMethodMap
	this.getMap = getMethodMap
}
func (this *restServer) getPath(url string) string {

	if strings.Contains(url, strings.TrimRight(Api_GetblkTxsByHeight, ":height")) {
		return Api_GetblkTxsByHeight
	} else if strings.Contains(url, strings.TrimRight(Api_Getblkbyheight, ":height")) {
		return Api_Getblkbyheight
	} else if strings.Contains(url, strings.TrimRight(Api_Getblkhash, ":height")) {
		return Api_Getblkhash
	} else if strings.Contains(url, strings.TrimRight(Api_Getblkbyhash, ":hash")) {
		return Api_Getblkbyhash
	} else if strings.Contains(url, strings.TrimRight(Api_GetTransaction, ":hash")) {
		return Api_GetTransaction
	} else if strings.Contains(url, strings.TrimRight(Api_GetContractState, ":hash")) {
		return Api_GetContractState
	} else if strings.Contains(url, strings.TrimRight(Api_GetSmtCodeEvtByHgt, ":height")) {
		return Api_GetSmtCodeEvtByHgt
	} else if strings.Contains(url, strings.TrimRight(Api_GetSmtCodeEvtByHash, ":hash")) {
		return Api_GetSmtCodeEvtByHash
	} else if strings.Contains(url, strings.TrimRight(Api_GetBlkHeightByTxHash, ":hash")) {
		return Api_GetBlkHeightByTxHash
	} else if strings.Contains(url, strings.TrimRight(Api_GetStorage, ":hash/:key")) {
		return Api_GetStorage
	} else if strings.Contains(url, strings.TrimRight(Api_GetBalanceByAddr, ":addr")) {
		return Api_GetBalanceByAddr
	}
	return url
}
func (this *restServer) getParams(r *http.Request, url string, req map[string]interface{}) map[string]interface{} {
	switch url {
	case Api_GetGenBlockTime:
	case Api_GetconnCount:
	case Api_GetblkTxsByHeight:
		req["Height"] = getParam(r, "height")
	case Api_Getblkbyheight:
		req["Raw"], req["Height"] = r.FormValue("raw"), getParam(r, "height")
	case Api_Getblkbyhash:
		req["Raw"], req["Hash"] = r.FormValue("raw"), getParam(r, "hash")
	case Api_Getblkheight:
	case Api_Getblkhash:
		req["Height"] = getParam(r, "height")
	case Api_GetTransaction:
		req["Hash"], req["Raw"] = getParam(r, "hash"), r.FormValue("raw")
	case Api_GetContractState:
		req["Hash"], req["Raw"] = getParam(r, "hash"), r.FormValue("raw")
	case Api_Restart:
	case Api_SendRawTx:
		userid := r.FormValue("userid")
		req["Userid"] = userid
		if len(userid) == 0 {
			req["Userid"] = getParam(r, "userid")
		}
		req["PreExec"] = r.FormValue("preExec")
	case Api_GetStorage:
		req["Hash"], req["Key"] = getParam(r, "hash"), getParam(r,"key")
	case Api_GetSmtCodeEvtByHgt:
		req["Height"] = getParam(r, "height")
	case Api_GetSmtCodeEvtByHash:
		req["Hash"] = getParam(r, "hash")
	case Api_GetBlkHeightByTxHash:
		req["Hash"] = getParam(r, "hash")
	case Api_GetBalanceByAddr:
		req["Addr"] = getParam(r, "addr")
	case Api_WebsocketState:
	default:
	}
	return req
}
func (this *restServer) initGetHandler() {

	for k, _ := range this.getMap {
		this.router.Get(k, func(w http.ResponseWriter, r *http.Request) {

			var req = make(map[string]interface{})
			var resp map[string]interface{}

			url := this.getPath(r.URL.Path)
			if h, ok := this.getMap[url]; ok {
				req = this.getParams(r, url, req)
				resp = h.handler(req)
				resp["Action"] = h.name
			} else {
				resp = rest.ResponsePack(berr.INVALID_METHOD)
			}
			this.response(w, resp)
		})
	}
}
func (this *restServer) initPostHandler() {
	for k, _ := range this.postMap {
		this.router.Post(k, func(w http.ResponseWriter, r *http.Request) {

			body, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			var req = make(map[string]interface{})
			var resp map[string]interface{}

			url := this.getPath(r.URL.Path)
			if h, ok := this.postMap[url]; ok {
				if err := json.Unmarshal(body, &req); err == nil {
					req = this.getParams(r, url, req)
					resp = h.handler(req)
					resp["Action"] = h.name
				} else {
					resp = rest.ResponsePack(berr.ILLEGAL_DATAFORMAT)
					resp["Action"] = h.name
				}
			} else {
				resp = rest.ResponsePack(berr.INVALID_METHOD)
			}
			this.response(w, resp)
		})
	}
	//Options
	for k, _ := range this.postMap {
		this.router.Options(k, func(w http.ResponseWriter, r *http.Request) {
			this.write(w, []byte{})
		})
	}

}
func (this *restServer) write(w http.ResponseWriter, data []byte) {
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(data)
}
func (this *restServer) response(w http.ResponseWriter, resp map[string]interface{}) {
	resp["Desc"] = berr.ErrMap[resp["Error"].(int64)]
	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal("HTTP Handle - json.Marshal: %v", err)
		return
	}
	this.write(w, data)
}
func (this *restServer) Stop() {
	if this.server != nil {
		this.server.Shutdown(context.Background())
		log.Error("Close restful ")
	}
}
func (this *restServer) Restart(cmd map[string]interface{}) map[string]interface{} {
	go func() {
		time.Sleep(time.Second)
		this.Stop()
		time.Sleep(time.Second)
		go this.Start()
	}()

	var resp = rest.ResponsePack(berr.SUCCESS)
	return resp
}
func (this *restServer) initTlsListen() (net.Listener, error) {

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

	log.Info("TLS listen port is ", strconv.Itoa(cfg.Parameters.HttpRestPort))
	listener, err := tls.Listen("tcp", ":" + strconv.Itoa(cfg.Parameters.HttpRestPort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
