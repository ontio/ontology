package httpjsonrpc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"GoOnchain/common"
)

//multiplexer that keeps track of every function to be called on specific rpc call
type ServeMux struct {
	m               map[string]func(*http.Request, map[string]interface{}) map[string]interface{}
	defaultFunction func(http.ResponseWriter, *http.Request)
}

//an instance of the multiplexer
var mainMux ServeMux

func InitServeMux() {
	mainMux.m = make(map[string]func(*http.Request, map[string]interface{}) map[string]interface{})
}

//a function to register functions to be called for specific rpc calls
func HandleFunc(pattern string, handler func(*http.Request, map[string]interface{}) map[string]interface{}) {
	mainMux.m[pattern] = handler
}

//a function to be called if the request is not a HTTP JSON RPC call
func SetDefaultFunc(def func(http.ResponseWriter, *http.Request)) {
	mainMux.defaultFunction = def
}

//this is the funciton that should be called in order to answer an rpc call
//should be registered like "http.HandleFunc("/", httpjsonrpc.Handle)"
func Handle(w http.ResponseWriter, r *http.Request) {
	//JSON RPC commands should be POSTs
	if r.Method != "POST" {
		if mainMux.defaultFunction != nil {
			log.Printf("HTTP JSON RPC Handle - Method!=\"POST\"")
			mainMux.defaultFunction(w, r)
			return
		} else {
			log.Panicf("HTTP JSON RPC Handle - Method!=\"POST\"")
			return
		}
	}

	//We must check if there is Request Body to read
	if r.Body == nil {
		if mainMux.defaultFunction != nil {
			log.Printf("HTTP JSON RPC Handle - Request body is nil")
			mainMux.defaultFunction(w, r)
			return
		} else {
			log.Panicf("HTTP JSON RPC Handle - Request body is nil")
			return
		}
	}

	//read the body of the request
	body, err := ioutil.ReadAll(r.Body)
	//log.Println(r)
	//log.Println(body)
	if err != nil {
		log.Fatalf("HTTP JSON RPC Handle - ioutil.ReadAll: %v", err)
		return
	}
	request := make(map[string]interface{})
	//unmarshal the request
	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Fatalf("HTTP JSON RPC Handle - json.Unmarshal: %v", err)
		return
	}
	//log.Println(request["method"])

	//get the corresponding function
	function, ok := mainMux.m[request["method"].(string)]

	if ok {
		response := function(r, request)
		//response from the program is encoded
		data, err := json.Marshal(response)
		if err != nil {
			log.Fatalf("HTTP JSON RPC Handle - json.Marshal: %v", err)
			return
		}
		//result is printed to the output
		w.Write(data)
	} else { //if the function does not exist
		log.Println("HTTP JSON RPC Handle - No function to call for", request["method"])

		//if you don't want to send an error, send something else:
		// data, err := json.Marshal(map[string]interface{}{
		//        	"result": "OK!",
		//        	"error": nil,
		//        	"id": request["id"],
		// })

		//an error json is created
		data, err := json.Marshal(map[string]interface{}{
			"result": nil,
			"error": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
				"data":    "The called method was not found on the server",
			},
			"id": request["id"],
		})
		if err != nil {
			log.Fatalf("HTTP JSON RPC Handle - json.Marshal: %v", err)
			return
		}
		w.Write(data)
	}
}

func getBalance(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getBestBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()

	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getInfo(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()

	//if err := json.Unmarshal(byt, &dat); err != nil {
	//	panic(err)
	//}

	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]

	// for range params
	//	param1 := params[0]
	log.Println(params)

	// data, err := json.Marshal(map[string]interface{}{
	// 	"method": method,
	// 	"id":     id,
	// 	"params": params,
	// })

	//if err != nil {
	//	log.Println("Parse the Getinfo request erro")
	//}

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type GetBlock struct {
	hash    interface{}
	verbose int
}

func getBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getBlockCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type GetBlockHash struct {
	index int //FIXME the maxium index overflow int?
}

func getBlockHash(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getConnectionCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getRawMemPool(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type GetRawTransaction struct {
	txid    interface{}
	verbose int
}

func getRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type GetTxOut struct {
	txid interface{}
	n    interface{}
}

func getTxOut(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()
	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type SendRawTransaction struct {
	hex interface{}
}

func sendRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()

	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type SendToAddress struct {
	asset_id interface{}
	address  interface{}
	value    float64 // Fixme max value overflow
	fee      float64
}

func sendToAddress(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()

	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type SubmitBlock struct {
	hex interface{}
}

func submitBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	common.Trace()

	id := cmd["id"].(float64)
	log.Println(id)

	params := cmd["params"]
	log.Println(params)
	//param1 := params[0]
	//log.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type GetBalance struct {
	assetId string
}

func StartServer() {
	common.Trace()
	InitServeMux()
	http.HandleFunc("/", Handle)
	HandleFunc("getinfo", getInfo)
	HandleFunc("sendtoaddress", sendToAddress)
	err := http.ListenAndServe("localhost:20332", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
