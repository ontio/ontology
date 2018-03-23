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
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/base/rest"
	Err "github.com/Ontology/http/base/error"
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

func InitRestServer() ApiServer {
	rt := &restServer{}

	rt.router = NewRouter()
	rt.registryMethod()
	rt.initGetHandler()
	rt.initPostHandler()
	return rt
}

func (rt *restServer) Start() error {
	if Parameters.HttpRestPort == 0 {
		log.Fatal("Not configure HttpRestPort port ")
		return nil
	}

	tlsFlag := false
	if tlsFlag || Parameters.HttpRestPort % 1000 == TlsPort {
		var err error
		rt.listener, err = rt.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		rt.listener, err = net.Listen("tcp", ":" + strconv.Itoa(Parameters.HttpRestPort))
		if err != nil {
			log.Fatal("net.Listen: ", err.Error())
			return err
		}
	}
	rt.server = &http.Server{Handler: rt.router}
	err := rt.server.Serve(rt.listener)

	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
		return err
	}

	return nil
}
func (rt *restServer) setWebsocketState(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(Err.SUCCESS)
	startFlag, ok := cmd["Open"].(bool)
	if !ok {
		resp["Error"] = Err.INVALID_PARAMS
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
		Parameters.HttpWsPort = int(wsPort)
	}
	if startFlag {
		websocket.ReStartServer()
	} else {
		websocket.Stop()
	}
	var result = make(map[string]interface{})
	result["Open"] = startFlag
	result["Port"] = Parameters.HttpWsPort
	result["PushBlock"] = websocket.GetWsPushBlockFlag()
	result["PushRawBlock"] = websocket.GetPushRawBlockFlag()
	result["PushBlockTxs"] = websocket.GetPushBlockTxsFlag()
	resp["Result"] = result
	return resp
}
func (rt *restServer) registryMethod() {

	getMethodMap := map[string]Action{
		Api_GetGenBlockTime:  {name: "getgenerateblocktime", handler: GetGenerateBlockTime},
		Api_GetconnCount:  {name: "getconnectioncount", handler: GetConnectionCount},
		Api_GetblkTxsByHeight: {name: "getblocktxsbyheight", handler: GetBlockTxsByHeight},
		Api_Getblkbyheight:    {name: "getblockbyheight", handler: GetBlockByHeight},
		Api_Getblkbyhash:         {name: "getblockbyhash", handler: GetBlockByHash},
		Api_Getblkheight:         {name: "getblockheight", handler: GetBlockHeight},
		Api_Getblkhash:           {name: "getblockhash", handler: GetBlockHash},
		Api_GetTransaction:       {name: "gettransaction", handler: GetTransactionByHash},
		Api_GetContractState:     {name: "getcontract", handler: GetContractState},
		Api_Restart:              {name: "restart", handler: rt.Restart},
		Api_GetSmtCodeEvtByHgt:    {name: "getsmartcodeeventbyheight", handler: GetSmartCodeEventByHeight},
		Api_GetSmtCodeEvtByHash:    {name: "getsmartcodeeventbyhash", handler: GetSmartCodeEventByTxHash},
		Api_GetBlkHeightByTxHash: {name: "getblockheightbytxhash", handler: GetBlockHeightByTxHash},
		Api_GetStorage:           {name: "getstorage", handler: GetStorage},
		Api_GetBalanceByAddr:    {name: "getbalance", handler: GetBalance},
	}

	sendRawTransaction := func(cmd map[string]interface{}) map[string]interface{} {
		resp := SendRawTransaction(cmd)
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
		Api_WebsocketState:     {name: "setwebsocketstate", handler: rt.setWebsocketState},
	}
	rt.postMap = postMethodMap
	rt.getMap = getMethodMap
}
func (rt *restServer) getPath(url string) string {

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
func (rt *restServer) getParams(r *http.Request, url string, req map[string]interface{}) map[string]interface{} {
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
func (rt *restServer) initGetHandler() {

	for k, _ := range rt.getMap {
		rt.router.Get(k, func(w http.ResponseWriter, r *http.Request) {

			var req = make(map[string]interface{})
			var resp map[string]interface{}

			url := rt.getPath(r.URL.Path)
			if h, ok := rt.getMap[url]; ok {
				req = rt.getParams(r, url, req)
				resp = h.handler(req)
				resp["Action"] = h.name
			} else {
				resp = RspInvalidMethod
			}
			rt.response(w, resp)
		})
	}
}
func (rt *restServer) initPostHandler() {
	for k, _ := range rt.postMap {
		rt.router.Post(k, func(w http.ResponseWriter, r *http.Request) {

			body, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			var req = make(map[string]interface{})
			var resp map[string]interface{}

			url := rt.getPath(r.URL.Path)
			if h, ok := rt.postMap[url]; ok {
				if err := json.Unmarshal(body, &req); err == nil {
					req = rt.getParams(r, url, req)
					resp = h.handler(req)
					resp["Action"] = h.name
				} else {
					resp = RspIllegalDataFormat
					resp["Action"] = h.name
				}
			} else {
				resp = RspInvalidMethod
			}
			rt.response(w, resp)
		})
	}
	//Options
	for k, _ := range rt.postMap {
		rt.router.Options(k, func(w http.ResponseWriter, r *http.Request) {
			rt.write(w, []byte{})
		})
	}

}
func (rt *restServer) write(w http.ResponseWriter, data []byte) {
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(data)
}
func (rt *restServer) response(w http.ResponseWriter, resp map[string]interface{}) {
	resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal("HTTP Handle - json.Marshal: %v", err)
		return
	}
	rt.write(w, data)
}
func (rt *restServer) Stop() {
	if rt.server != nil {
		rt.server.Shutdown(context.Background())
		log.Error("Close restful ")
	}
}
func (rt *restServer) Restart(cmd map[string]interface{}) map[string]interface{} {
	go func() {
		time.Sleep(time.Second)
		rt.Stop()
		time.Sleep(time.Second)
		go rt.Start()
	}()

	var resp = ResponsePack(Err.SUCCESS)
	return resp
}
func (rt *restServer) initTlsListen() (net.Listener, error) {

	CertPath := Parameters.RestCertPath
	KeyPath := Parameters.RestKeyPath

	// load cert
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	log.Info("TLS listen port is ", strconv.Itoa(Parameters.HttpRestPort))
	listener, err := tls.Listen("tcp", ":" + strconv.Itoa(Parameters.HttpRestPort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
