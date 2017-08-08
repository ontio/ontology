package restful

import (
	. "DNA/common/config"
	"DNA/common/log"
	. "DNA/net/httprestful/common"
	Err "DNA/net/httprestful/error"
	"DNA/net/httpwebsocket"
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
	checkAccessToken func(auth_type, access_token string) (string, int64, interface{})
}

const (
	Api_Getconnectioncount  = "/api/v1/node/connectioncount"
	Api_GetblockTxsByHeight = "/api/v1/block/transactions/height/:height"
	Api_Getblockbyheight    = "/api/v1/block/details/height/:height"
	Api_Getblockbyhash      = "/api/v1/block/details/hash/:hash"
	Api_Getblockheight      = "/api/v1/block/height"
	Api_Getblockhash        = "/api/v1/block/hash/:height"
	Api_GetTotalIssued      = "/api/v1/totalissued/:assetid"
	Api_Gettransaction      = "/api/v1/transaction/:hash"
	Api_Getasset            = "/api/v1/asset/:hash"
	Api_GetBalanceByAddr    = "/api/v1/asset/balances/:addr"
	Api_GetBalancebyAsset   = "/api/v1/asset/balance/:addr/:assetid"
	Api_GetUTXObyAsset      = "/api/v1/asset/utxo/:addr/:assetid"
	Api_GetUTXObyAddr       = "/api/v1/asset/utxos/:addr"
	Api_SendRawTx           = "/api/v1/transaction"
	Api_SendRcdTxByTrans    = "/api/v1/custom/transaction/record"
	Api_GetStateUpdate      = "/api/v1/stateupdate/:namespace/:key"
	Api_OauthServerUrl      = "/api/v1/config/oauthserver/url"
	Api_NoticeServerUrl     = "/api/v1/config/noticeserver/url"
	Api_NoticeServerState   = "/api/v1/config/noticeserver/state"
	Api_WebsocketState      = "/api/v1/config/websocket/state"
	Api_Restart             = "/api/v1/restart"
	Api_GetContract         = "/api/v1/contract/:hash"
)

func InitRestServer(checkAccessToken func(string, string) (string, int64, interface{})) ApiServer {
	rt := &restServer{}
	rt.checkAccessToken = checkAccessToken

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
	if tlsFlag || Parameters.HttpRestPort%1000 == TlsPort {
		var err error
		rt.listener, err = rt.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		rt.listener, err = net.Listen("tcp", ":"+strconv.Itoa(Parameters.HttpRestPort))
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
		httpwebsocket.SetWsPushBlockFlag(b)
	}
	if b, ok := cmd["PushRawBlock"].(bool); ok {
		httpwebsocket.SetPushRawBlockFlag(b)
	}
	if b, ok := cmd["PushBlockTxs"].(bool); ok {
		httpwebsocket.SetPushBlockTxsFlag(b)
	}
	if wsPort, ok := cmd["Port"].(float64); ok && wsPort != 0 {
		Parameters.HttpWsPort = int(wsPort)
	}
	if startFlag {
		httpwebsocket.ReStartServer()
	} else {
		httpwebsocket.Stop()
	}
	var result = make(map[string]interface{})
	result["Open"] = startFlag
	result["Port"] = Parameters.HttpWsPort
	result["PushBlock"] = httpwebsocket.GetWsPushBlockFlag()
	result["PushRawBlock"] = httpwebsocket.GetPushRawBlockFlag()
	result["PushBlockTxs"] = httpwebsocket.GetPushBlockTxsFlag()
	resp["Result"] = result
	return resp
}
func (rt *restServer) registryMethod() {

	getMethodMap := map[string]Action{
		Api_Getconnectioncount:  {name: "getconnectioncount", handler: GetConnectionCount},
		Api_GetblockTxsByHeight: {name: "getblocktransactionsbyheight", handler: GetBlockTxsByHeight},
		Api_Getblockbyheight:    {name: "getblockbyheight", handler: GetBlockByHeight},
		Api_Getblockbyhash:      {name: "getblockbyhash", handler: GetBlockByHash},
		Api_Getblockheight:      {name: "getblockheight", handler: GetBlockHeight},
		Api_Getblockhash:        {name: "getblockhash", handler: GetBlockHash},
		Api_GetTotalIssued:      {name: "gettotalissued", handler: GetTotalIssued},
		Api_Gettransaction:      {name: "gettransaction", handler: GetTransactionByHash},
		Api_Getasset:            {name: "getasset", handler: GetAssetByHash},
		Api_GetContract:         {name: "getcontract", handler: GetContract},
		Api_GetUTXObyAddr:       {name: "getutxobyaddr", handler: GetUnspends},
		Api_GetUTXObyAsset:      {name: "getutxobyasset", handler: GetUnspendOutput},
		Api_GetBalanceByAddr:    {name: "getbalancebyaddr", handler: GetBalanceByAddr},
		Api_GetBalancebyAsset:   {name: "getbalancebyasset", handler: GetBalanceByAsset},
		Api_OauthServerUrl:      {name: "getoauthserverurl", handler: GetOauthServerUrl},
		Api_NoticeServerUrl:     {name: "getnoticeserverurl", handler: GetNoticeServerUrl},
		Api_Restart:             {name: "restart", handler: rt.Restart},
		Api_GetStateUpdate:      {name: "getstateupdate", handler: GetStateUpdate},
	}

	sendRawTransaction := func(cmd map[string]interface{}) map[string]interface{} {
		resp := SendRawTransaction(cmd)
		if userid, ok := resp["Userid"].(string); ok && len(userid) > 0 {
			if result, ok := resp["Result"].(string); ok {
				httpwebsocket.SetTxHashMap(result, userid)
			}
			delete(resp, "Userid")
		}
		return resp
	}
	postMethodMap := map[string]Action{
		Api_SendRawTx:         {name: "sendrawtransaction", handler: sendRawTransaction},
		Api_SendRcdTxByTrans:  {name: "sendrecord", handler: SendRecord},
		Api_OauthServerUrl:    {name: "setoauthserverurl", handler: SetOauthServerUrl},
		Api_NoticeServerUrl:   {name: "setnoticeserverurl", handler: SetNoticeServerUrl},
		Api_NoticeServerState: {name: "setpostblock", handler: SetPushBlockFlag},
		Api_WebsocketState:    {name: "setwebsocketstate", handler: rt.setWebsocketState},
	}
	rt.postMap = postMethodMap
	rt.getMap = getMethodMap
}
func (rt *restServer) getPath(url string) string {

	if strings.Contains(url, strings.TrimRight(Api_GetblockTxsByHeight, ":height")) {
		return Api_GetblockTxsByHeight
	} else if strings.Contains(url, strings.TrimRight(Api_Getblockbyheight, ":height")) {
		return Api_Getblockbyheight
	} else if strings.Contains(url, strings.TrimRight(Api_Getblockhash, ":height")) {
		return Api_Getblockhash
	} else if strings.Contains(url, strings.TrimRight(Api_Getblockbyhash, ":hash")) {
		return Api_Getblockbyhash
	} else if strings.Contains(url, strings.TrimRight(Api_GetTotalIssued, ":assetid")) {
		return Api_GetTotalIssued
	} else if strings.Contains(url, strings.TrimRight(Api_Gettransaction, ":hash")) {
		return Api_Gettransaction
	} else if strings.Contains(url, strings.TrimRight(Api_GetContract, ":hash")) {
		return Api_GetContract
	} else if strings.Contains(url, strings.TrimRight(Api_GetBalanceByAddr, ":addr")) {
		return Api_GetBalanceByAddr
	} else if strings.Contains(url, strings.TrimRight(Api_GetBalancebyAsset, ":addr/:assetid")) {
		return Api_GetBalancebyAsset
	} else if strings.Contains(url, strings.TrimRight(Api_GetUTXObyAddr, ":addr")) {
		return Api_GetUTXObyAddr
	} else if strings.Contains(url, strings.TrimRight(Api_GetUTXObyAsset, ":addr/:assetid")) {
		return Api_GetUTXObyAsset
	} else if strings.Contains(url, strings.TrimRight(Api_Getasset, ":hash")) {
		return Api_Getasset
	} else if strings.Contains(url, strings.TrimRight(Api_GetStateUpdate, ":namespace/:key")) {
		return Api_GetStateUpdate
	}
	return url
}
func (rt *restServer) getParams(r *http.Request, url string, req map[string]interface{}) map[string]interface{} {
	switch url {
	case Api_Getconnectioncount:
		break
	case Api_GetblockTxsByHeight:
		req["Height"] = getParam(r, "height")
		break
	case Api_Getblockbyheight:
		req["Raw"] = r.FormValue("raw")
		req["Height"] = getParam(r, "height")
		break
	case Api_Getblockbyhash:
		req["Raw"] = r.FormValue("raw")
		req["Hash"] = getParam(r, "hash")
		break
	case Api_Getblockheight:
		break
	case Api_Getblockhash:
		req["Height"] = getParam(r, "height")
		break
	case Api_GetTotalIssued:
		req["Assetid"] = getParam(r, "assetid")
		break
	case Api_Gettransaction:
		req["Hash"] = getParam(r, "hash")
		req["Raw"] = r.FormValue("raw")
		break
	case Api_GetContract:
		req["Hash"] = getParam(r, "hash")
		req["Raw"] = r.FormValue("raw")
		break
	case Api_Getasset:
		req["Hash"] = getParam(r, "hash")
		req["Raw"] = r.FormValue("raw")
		break
	case Api_GetBalancebyAsset:
		req["Addr"] = getParam(r, "addr")
		req["Assetid"] = getParam(r, "assetid")
		break
	case Api_GetBalanceByAddr:
		req["Addr"] = getParam(r, "addr")
		break
	case Api_GetUTXObyAddr:
		req["Addr"] = getParam(r, "addr")
		break
	case Api_GetUTXObyAsset:
		req["Addr"] = getParam(r, "addr")
		req["Assetid"] = getParam(r, "assetid")
		break
	case Api_Restart:
		break
	case Api_SendRawTx:
		userid := r.FormValue("userid")
		if len(userid) == 0 {
			req["Userid"] = getParam(r, "userid")
		}
		break
	case Api_SendRcdTxByTrans:
		req["Raw"] = r.FormValue("raw")
		break
	case Api_GetStateUpdate:
		req["Namespace"] = getParam(r, "namespace")
		req["Key"] = getParam(r, "key")
		break
	case Api_OauthServerUrl:
	case Api_NoticeServerUrl:
	case Api_NoticeServerState:
	case Api_WebsocketState:
		break
	default:
	}
	return req
}
func (rt *restServer) initGetHandler() {

	for k, _ := range rt.getMap {
		rt.router.Get(k, func(w http.ResponseWriter, r *http.Request) {

			var req = make(map[string]interface{})
			var resp map[string]interface{}
			access_token := r.FormValue("access_token")
			auth_type := r.FormValue("auth_type")

			CAkey, errCode, result := rt.checkAccessToken(auth_type, access_token)
			url := rt.getPath(r.URL.Path)
			if errCode > 0 && r.URL.Path != Api_OauthServerUrl {
				resp = ResponsePack(errCode)
				resp["Result"] = result
				rt.response(w, resp)
				return
			}
			if h, ok := rt.getMap[url]; ok {
				req["CAkey"] = CAkey
				req = rt.getParams(r, url, req)
				resp = h.handler(req)
				resp["Action"] = h.name
			} else {
				resp = ResponsePack(Err.INVALID_METHOD)
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
			access_token := r.FormValue("access_token")
			auth_type := r.FormValue("auth_type")

			CAkey, errCode, result := rt.checkAccessToken(auth_type, access_token)
			url := rt.getPath(r.URL.Path)
			if errCode > 0 && r.URL.Path != Api_OauthServerUrl {
				resp = ResponsePack(errCode)
				resp["Result"] = result
				rt.response(w, resp)
				return
			}
			if h, ok := rt.postMap[url]; ok {
				if err := json.Unmarshal(body, &req); err == nil {
					req["CAkey"] = CAkey
					req = rt.getParams(r, url, req)
					resp = h.handler(req)
					resp["Action"] = h.name
				} else {
					resp = ResponsePack(Err.ILLEGAL_DATAFORMAT)
					resp["Action"] = h.name
				}
			} else {
				resp = ResponsePack(Err.INVALID_METHOD)
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
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(Parameters.HttpRestPort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
