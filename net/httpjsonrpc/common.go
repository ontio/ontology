package httpjsonrpc

import (
	"GoOnchain/consensus/dbft"
	"GoOnchain/core/ledger"
	tx "GoOnchain/core/transaction"
	. "GoOnchain/net/protocol"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func init() {
	mainMux.m = make(map[string]func(*http.Request, map[string]interface{}) map[string]interface{})
}

//an instance of the multiplexer
var mainMux ServeMux
var node Noder
var dBFT *dbft.DbftService

//multiplexer that keeps track of every function to be called on specific rpc call
type ServeMux struct {
	m               map[string]func(*http.Request, map[string]interface{}) map[string]interface{}
	defaultFunction func(http.ResponseWriter, *http.Request)
}

type BlockInfo struct {
	Hash      string
	BlockData *ledger.Blockdata
}

type TxInfo struct {
	Hash string
	Hex  string
	Tx   *tx.Transaction
}

type TxoutInfo struct {
	High  uint32
	Low   uint32
	Txout tx.TxOutput
}

type NodeInfo struct {
	State    uint   // node status
	Port     uint16 // The nodes's port
	ID       uint64 // The nodes's id
	Time     int64
	Version  uint32 // The network protocol the node used
	Services uint64 // The services the node supplied
	Relay    bool   // The relay capability of the node (merge into capbility flag)
	Height   uint64 // The node latest block height
}

type ConsensusInfo struct {
	// TODO
}

func RegistRpcNode(n Noder) {
	if node == nil {
		node = n
	}
}

func RegistDbftService(d *dbft.DbftService) {
	if dBFT == nil {
		dBFT = d
	}
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

func responsePacking(result interface{}, id interface{}) map[string]interface{} {
	resp := map[string]interface{}{
		"Jsonrpc": "2.0",
		"Result":  result,
		"Id":      id,
	}
	return resp
}

// Call sends RPC request to server
func Call(address string, method string, id interface{}, params []interface{}) ([]byte, error) {
	data, err := json.Marshal(map[string]interface{}{
		"method": method,
		"id":     id,
		"params": params,
	})
	if err != nil {
		log.Fatalf("Marshal: %v", err)
		return nil, err
	}
	resp, err := http.Post(address, "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Post: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
		return nil, err
	}

	return body, nil
}
