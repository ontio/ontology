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
	"context"
	"crypto/tls"
	"encoding/json"
	cfg "github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	berr "github.com/ontio/ontology/http/base/error"
	"github.com/ontio/ontology/http/base/rest"
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
	router   *Router
	listener net.Listener
	server   *http.Server
	postMap  map[string]Action
	getMap   map[string]Action
}

const (
	GET_GEN_BLK_TIME      = "/api/v1/node/generateblocktime"
	GET_CONN_COUNT        = "/api/v1/node/connectioncount"
	GET_BLK_TXS_BY_HEIGHT = "/api/v1/block/transactions/height/:height"
	GET_BLK_BY_HEIGHT     = "/api/v1/block/details/height/:height"
	GET_BLK_BY_HASH       = "/api/v1/block/details/hash/:hash"
	GET_BLK_HEIGHT        = "/api/v1/block/height"
	GET_BLK_HASH          = "/api/v1/block/hash/:height"
	GET_TX                = "/api/v1/transaction/:hash"
	GET_STORAGE           = "/api/v1/storage/:hash/:key"
	GET_BALANCE           = "/api/v1/balance/:addr"
	GET_CONTRACT_STATE    = "/api/v1/contract/:hash"
	GET_SMTCOCE_EVT_TXS   = "/api/v1/smartcode/event/transactions/:height"
	GET_SMTCOCE_EVTS      = "/api/v1/smartcode/event/txhash/:hash"
	GET_BLK_HGT_BY_TXHASH = "/api/v1/block/height/txhash/:hash"
	GET_MERKLE_PROOF      = "/api/v1/merkleproof/:hash"
	GET_GAS_PRICE         = "/api/v1/gasprice"
	GET_ALLOWANCE         = "/api/v1/allowance/:asset/:from/:to"

	POST_RAW_TX = "/api/v1/transaction"
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
	retPort := int(cfg.DefConfig.Restful.HttpRestPort)
	if retPort == 0 {
		log.Fatal("Not configure HttpRestPort port ")
		return nil
	}

	tlsFlag := false
	if tlsFlag || retPort%1000 == rest.TLS_PORT {
		var err error
		this.listener, err = this.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		this.listener, err = net.Listen("tcp", ":"+strconv.Itoa(retPort))
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

func (this *restServer) registryMethod() {

	getMethodMap := map[string]Action{
		GET_GEN_BLK_TIME:      {name: "getgenerateblocktime", handler: rest.GetGenerateBlockTime},
		GET_CONN_COUNT:        {name: "getconnectioncount", handler: rest.GetConnectionCount},
		GET_BLK_TXS_BY_HEIGHT: {name: "getblocktxsbyheight", handler: rest.GetBlockTxsByHeight},
		GET_BLK_BY_HEIGHT:     {name: "getblockbyheight", handler: rest.GetBlockByHeight},
		GET_BLK_BY_HASH:       {name: "getblockbyhash", handler: rest.GetBlockByHash},
		GET_BLK_HEIGHT:        {name: "getblockheight", handler: rest.GetBlockHeight},
		GET_BLK_HASH:          {name: "getblockhash", handler: rest.GetBlockHash},
		GET_TX:                {name: "gettransaction", handler: rest.GetTransactionByHash},
		GET_CONTRACT_STATE:    {name: "getcontract", handler: rest.GetContractState},
		GET_SMTCOCE_EVT_TXS:   {name: "getsmartcodeeventbyheight", handler: rest.GetSmartCodeEventTxsByHeight},
		GET_SMTCOCE_EVTS:      {name: "getsmartcodeeventbyhash", handler: rest.GetSmartCodeEventByTxHash},
		GET_BLK_HGT_BY_TXHASH: {name: "getblockheightbytxhash", handler: rest.GetBlockHeightByTxHash},
		GET_STORAGE:           {name: "getstorage", handler: rest.GetStorage},
		GET_BALANCE:           {name: "getbalance", handler: rest.GetBalance},
		GET_ALLOWANCE:         {name: "getallowance", handler: rest.GetAllowance},
		GET_MERKLE_PROOF:      {name: "getmerkleproof", handler: rest.GetMerkleProof},
		GET_GAS_PRICE:         {name: "getgasprice", handler: rest.GetGasPrice},
	}

	postMethodMap := map[string]Action{
		POST_RAW_TX: {name: "sendrawtransaction", handler: rest.SendRawTransaction},
	}
	this.postMap = postMethodMap
	this.getMap = getMethodMap
}
func (this *restServer) getPath(url string) string {

	if strings.Contains(url, strings.TrimRight(GET_BLK_TXS_BY_HEIGHT, ":height")) {
		return GET_BLK_TXS_BY_HEIGHT
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_BY_HEIGHT, ":height")) {
		return GET_BLK_BY_HEIGHT
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_HASH, ":height")) {
		return GET_BLK_HASH
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_BY_HASH, ":hash")) {
		return GET_BLK_BY_HASH
	} else if strings.Contains(url, strings.TrimRight(GET_TX, ":hash")) {
		return GET_TX
	} else if strings.Contains(url, strings.TrimRight(GET_CONTRACT_STATE, ":hash")) {
		return GET_CONTRACT_STATE
	} else if strings.Contains(url, strings.TrimRight(GET_SMTCOCE_EVT_TXS, ":height")) {
		return GET_SMTCOCE_EVT_TXS
	} else if strings.Contains(url, strings.TrimRight(GET_SMTCOCE_EVTS, ":hash")) {
		return GET_SMTCOCE_EVTS
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_HGT_BY_TXHASH, ":hash")) {
		return GET_BLK_HGT_BY_TXHASH
	} else if strings.Contains(url, strings.TrimRight(GET_STORAGE, ":hash/:key")) {
		return GET_STORAGE
	} else if strings.Contains(url, strings.TrimRight(GET_BALANCE, ":addr")) {
		return GET_BALANCE
	} else if strings.Contains(url, strings.TrimRight(GET_MERKLE_PROOF, ":hash")) {
		return GET_MERKLE_PROOF
	} else if strings.Contains(url, strings.TrimRight(GET_ALLOWANCE, ":asset/:from/:to")) {
		return GET_ALLOWANCE
	}
	return url
}
func (this *restServer) getParams(r *http.Request, url string, req map[string]interface{}) map[string]interface{} {
	switch url {
	case GET_GEN_BLK_TIME:
	case GET_CONN_COUNT:
	case GET_BLK_TXS_BY_HEIGHT:
		req["Height"] = getParam(r, "height")
	case GET_BLK_BY_HEIGHT:
		req["Raw"], req["Height"] = r.FormValue("raw"), getParam(r, "height")
	case GET_BLK_BY_HASH:
		req["Raw"], req["Hash"] = r.FormValue("raw"), getParam(r, "hash")
	case GET_BLK_HEIGHT:
	case GET_BLK_HASH:
		req["Height"] = getParam(r, "height")
	case GET_TX:
		req["Hash"], req["Raw"] = getParam(r, "hash"), r.FormValue("raw")
	case GET_CONTRACT_STATE:
		req["Hash"], req["Raw"] = getParam(r, "hash"), r.FormValue("raw")
	case POST_RAW_TX:
		req["PreExec"] = r.FormValue("preExec")
	case GET_STORAGE:
		req["Hash"], req["Key"] = getParam(r, "hash"), getParam(r, "key")
	case GET_SMTCOCE_EVT_TXS:
		req["Height"] = getParam(r, "height")
	case GET_SMTCOCE_EVTS:
		req["Hash"] = getParam(r, "hash")
	case GET_BLK_HGT_BY_TXHASH:
		req["Hash"] = getParam(r, "hash")
	case GET_BALANCE:
		req["Addr"] = getParam(r, "addr")
	case GET_MERKLE_PROOF:
		req["Hash"] = getParam(r, "hash")
	case GET_ALLOWANCE:
		req["Asset"] = getParam(r, "asset")
		req["From"], req["To"] = getParam(r, "from"), getParam(r, "to")
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

	certPath := cfg.DefConfig.Restful.HttpCertPath
	keyPath := cfg.DefConfig.Restful.HttpKeyPath

	// load cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	restPort := strconv.Itoa(int(cfg.DefConfig.Restful.HttpRestPort))
	log.Info("TLS listen port is ", restPort)
	listener, err := tls.Listen("tcp", ":"+restPort, tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
