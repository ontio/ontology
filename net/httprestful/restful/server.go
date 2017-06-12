package restful

import (
	. "DNA/common/config"
	"DNA/common/log"
	. "DNA/net/httprestful/common"
	Err "DNA/net/httprestful/error"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
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
	checkAccessToken func(auth_type, access_token string) bool
}

const (
	Api_Getconnectioncount = "/api/v1/node/connectioncount"
	Api_Getblockbyheight   = "/api/v1/block/details/height/:height"
	Api_Getblockbyhash     = "/api/v1/block/details/hash/:hash"
	Api_Getblockheight     = "/api/v1/block/height"
	Api_Gettransaction     = "/api/v1/transaction/:hash"
	Api_Getasset           = "/api/v1/asset/:hash"
	Api_Restart            = "/api/v1/restart"
	Api_SendRawTransaction = "/api/v1/transaction"
	Api_OauthServerAddr    = "/api/v1/config/oauthserver/addr"
	Api_NoticeServerAddr   = "/api/v1/config/noticeserver/addr"
	Api_NoticeServerState  = "/api/v1/config/noticeserver/state"
)

func InitRestServer(checkAccessToken func(string, string) bool) ApiServer {
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
	if tlsFlag {
		rt.listener, _ = rt.initTlsListen()
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

func (rt *restServer) registryMethod() {

	getMethodMap := map[string]Action{
		Api_Getconnectioncount: {name: "getconnectioncount", handler: GetConnectionCount},
		Api_Getblockbyheight:   {name: "getblockbyheight", handler: GetBlockByHeight},
		Api_Getblockbyhash:     {name: "getblockbyhash", handler: GetBlockByHash},
		Api_Getblockheight:     {name: "getblockheight", handler: GetBlockHeight},
		Api_Gettransaction:     {name: "gettransaction", handler: GetTransactionByHash},
		Api_Getasset:           {name: "getasset", handler: GetAssetByHash},
		Api_OauthServerAddr:    {name: "getoauthserveraddr", handler: GetOauthServerAddr},
		Api_NoticeServerAddr:   {name: "getnoticeserveraddr", handler: GetNoticeServerAddr},
		Api_Restart:            {name: "restart", handler: rt.Restart},
	}

	postMethodMap := map[string]Action{
		Api_SendRawTransaction: {name: "sendrawtransaction", handler: SendRawTransaction},
		Api_OauthServerAddr:    {name: "setoauthserveraddr", handler: SetOauthServerAddr},
		Api_NoticeServerAddr:   {name: "setnoticeserveraddr", handler: SetNoticeServerAddr},
		Api_NoticeServerState:  {name: "setpostblock", handler: SetPushBlockFlag},
	}
	rt.postMap = postMethodMap
	rt.getMap = getMethodMap
}
func (rt *restServer) getPath(url string) string {

	if strings.Contains(url, strings.TrimRight(Api_Getblockbyheight, ":height")) {
		return Api_Getblockbyheight
	} else if strings.Contains(url, strings.TrimRight(Api_Getblockbyhash, ":hash")) {
		return Api_Getblockbyhash
	} else if strings.Contains(url, strings.TrimRight(Api_Gettransaction, ":hash")) {
		return Api_Gettransaction
	} else if strings.Contains(url, strings.TrimRight(Api_Getasset, ":hash")) {
		return Api_Getasset
	}
	return url
}

func (rt *restServer) initGetHandler() {

	for k, _ := range rt.getMap {
		rt.router.Get(k, func(w http.ResponseWriter, r *http.Request) {

			var reqMsg = make(map[string]interface{})
			var data []byte
			var err error
			var resp map[string]interface{}
			access_token := r.FormValue("access_token")
			auth_type := r.FormValue("auth_type")

			if !rt.checkAccessToken(auth_type, access_token) && r.URL.Path != Api_OauthServerAddr {
				resp = ResponsePack(Err.INVALID_TOKEN)
				goto ResponseWrite
			}
			if h, ok := rt.getMap[rt.getPath(r.URL.Path)]; ok {

				reqMsg["Height"] = getParam(r, "height")
				reqMsg["Hash"] = getParam(r, "hash")
				resp = h.handler(reqMsg)
				resp["Action"] = h.name
			} else {
				resp = ResponsePack(Err.INVALID_METHOD)
			}
		ResponseWrite:
			resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
			data, err = json.Marshal(resp)
			if err != nil {
				log.Fatal("HTTP Handle - json.Marshal: %v", err)
				return
			}
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("content-type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Connection", "close")
			w.Write([]byte(data))
		})
	}
}
func (rt *restServer) initPostHandler() {
	for k, _ := range rt.postMap {
		rt.router.Post(k, func(w http.ResponseWriter, r *http.Request) {

			body, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			var reqMsg = make(map[string]interface{})
			var data []byte
			var err error

			access_token := r.FormValue("access_token")
			auth_type := r.FormValue("auth_type")
			var resp map[string]interface{}
			if !rt.checkAccessToken(auth_type, access_token) && r.URL.Path != Api_OauthServerAddr {
				resp = ResponsePack(Err.INVALID_TOKEN)
				data, _ = json.Marshal(resp)
				goto ResponseWrite
			}

			if h, ok := rt.postMap[rt.getPath(r.URL.Path)]; ok {

				if err = json.Unmarshal(body, &reqMsg); err == nil {

					resp = h.handler(reqMsg)
					resp["Action"] = h.name

				} else {
					resp = ResponsePack(Err.ILLEGAL_DATAFORMAT)
					resp["Action"] = h.name
					data, _ = json.Marshal(resp)
				}
			}
		ResponseWrite:
			resp["Desc"] = Err.ErrMap[resp["Error"].(int64)]
			data, err = json.Marshal(resp)
			if err != nil {
				log.Fatal("HTTP Handle - json.Marshal: %v", err)
				return
			}
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("content-type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Connection", "close")
			w.Write([]byte(data))
		})
	}
	//Options
	for k, _ := range rt.postMap {
		rt.router.Options(k, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("content-type", "application/json")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write([]byte{})
		})
	}

}
func (rt *restServer) Stop() {
	rt.server.Shutdown(context.Background())
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
	//TODO TLS test, Cert Parameters
	CertPath := Parameters.CertPath
	KeyPath := Parameters.KeyPath
	CAPath := Parameters.CAPath

	// load cert
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}
	// load root ca
	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		log.Error("read ca fail", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	ret := pool.AppendCertsFromPEM(caData)
	if !ret {
		return nil, errors.New("failed to parse root certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}

	log.Info("TLS listen port is ", strconv.Itoa(Parameters.HttpRestPort))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(Parameters.HttpRestPort), tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
