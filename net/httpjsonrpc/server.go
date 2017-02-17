package httpjsonrpc

import (
	. "GoOnchain/common"
	"GoOnchain/core/ledger"
	tx "GoOnchain/core/transaction"
	"GoOnchain/net/protocol"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

//multiplexer that keeps track of every function to be called on specific rpc call
type ServeMux struct {
	m               map[string]func(*http.Request, map[string]interface{}) map[string]interface{}
	defaultFunction func(http.ResponseWriter, *http.Request)
}

type BlockInfo struct {
	Hash  string
	Block *ledger.Block
}

type NoderInfo struct {
	Noder protocol.JsonNoder
}

type TxInfo struct {
	Hash string
	Hex  string
	Tx   *tx.Transaction
}

var nodeInfo NoderInfo

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

func InitNoderInfo(jsonNode protocol.JsonNoder) {
	//TODO
	//return NodeInfo
	if nodeInfo.Noder == nil {
		nodeInfo.Noder = jsonNode
	}
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

func responsePacking(result interface{}, id uint) map[string]interface{} {
	resp := map[string]interface{}{
		"Jsonrpc": "2.0",
		"Result":  result,
		"Id":      id,
	}
	return resp
}

func getBestBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	hash := ledger.DefaultLedger.Blockchain.CurrentBlockHash()
	response := responsePacking(ToHexString(hash.ToArray()), id)
	return response
}

func getBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	params := cmd["params"]
	var block *ledger.Block
	var err error
	var b BlockInfo
	switch (params.([]interface{})[0]).(type) {
	case int:
		index := params.([]interface{})[0].(uint32)
		hash := ledger.DefaultLedger.Store.GetBlockHash(index)
		block, err = ledger.DefaultLedger.Store.GetBlock(hash)
		b = BlockInfo{
			Hash:  ToHexString(hash.ToArray()),
			Block: block,
		}
	case string:
		hash := params.([]interface{})[0].(string)
		hashslice, _ := hex.DecodeString(hash)
		var hasharr Uint256
		hasharr.Deserialize(bytes.NewReader(hashslice[0:32]))
		block, err = ledger.DefaultLedger.Store.GetBlock(hasharr)
		b = BlockInfo{
			Hash:  hash,
			Block: block,
		}
	}

	if err != nil {
		var erro []interface{} = []interface{}{-100, "Unknown block"}
		response := responsePacking(erro, id)
		return response
	}

	raw, _ := json.Marshal(&b)
	response := responsePacking(string(raw), id)
	return response
}

func getBlockCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	count := ledger.DefaultLedger.Blockchain.BlockHeight + 1
	response := responsePacking(count, id)
	return response
}

func getBlockHash(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	index := cmd["params"]
	var hash Uint256
	height, ok := index.(uint32)
	if ok == true {
		hash = ledger.DefaultLedger.Store.GetBlockHash(height)
	}
	hashhex := fmt.Sprintf("%016x", hash)
	response := responsePacking(hashhex, id)
	return response
}

func getConnectionCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	count := nodeInfo.Noder.GetConnectionCnt()
	response := responsePacking(count, id)
	return response
}

func getRawMemPool(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	mempoollist := nodeInfo.Noder.GetTxnPool()
	raw, _ := json.Marshal(mempoollist)
	response := responsePacking(string(raw), id)
	return response
}

func getRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	params := cmd["params"]
	txid := params.([]interface{})[0].(string)
	txidSlice, _ := hex.DecodeString(txid)
	var txidArr Uint256
	txidArr.Deserialize(bytes.NewReader(txidSlice[0:32]))
	verbose := params.([]interface{})[1].(bool)
	tx := nodeInfo.Noder.GetTransaction(txidArr)
	txBuffer := bytes.NewBuffer([]byte{})
	tx.Serialize(txBuffer)
	if verbose == true {
		t := TxInfo{
			Hash: txid,
			Hex:  hex.EncodeToString(txBuffer.Bytes()),
			Tx:   tx,
		}
		raw, _ := json.Marshal(&t)
		response := responsePacking(string(raw), id)
		return response
	}

	response := responsePacking(txBuffer.Bytes(), id)
	return response
}

type TxoutInfo struct {
	High  uint32
	Low   uint32
	Txout tx.TxOutput
}

func getTxout(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	//params := cmd["params"]
	//txid := params.([]interface{})[0].(string)
	//var n int = params.([]interface{})[1].(int)
	var txout tx.TxOutput // := tx.GetTxOut() //TODO
	high := uint32(txout.Value >> 32)
	low := uint32(txout.Value)
	to := TxoutInfo{
		High:  high,
		Low:   low,
		Txout: txout,
	}
	raw, _ := json.Marshal(&to)
	response := responsePacking(string(raw), id)
	return response

}

func sendRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	hexValue := cmd["params"].(string)
	hexSlice, _ := hex.DecodeString(hexValue)
	var txTransaction tx.Transaction
	txTransaction.Deserialize(bytes.NewReader(hexSlice[:]))
	err := nodeInfo.Noder.Xmit(&txTransaction)
	response := responsePacking(err, id)
	return response
}

func submitBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	id := cmd["id"].(uint)
	hexValue := cmd["params"].(string)
	hexSlice, _ := hex.DecodeString(hexValue)
	var txTransaction tx.Transaction
	txTransaction.Deserialize(bytes.NewReader(hexSlice[:]))
	err := nodeInfo.Noder.Xmit(&txTransaction)
	response := responsePacking(err, id)
	return response
}

func StartServer() {
	Trace()
	InitServeMux()
	http.HandleFunc("/", Handle)

	HandleFunc("getbestblock", getBestBlock)
	HandleFunc("getblock", getBlock)
	HandleFunc("getblockcount", getBlockCount)
	HandleFunc("getblockhash", getBlockHash)
	HandleFunc("getconnectioncount", getConnectionCount)
	HandleFunc("getrawmempool", getRawMemPool)
	HandleFunc("getrawtransaction", getRawTransaction)
	HandleFunc("submitblock", submitBlock)

	err := http.ListenAndServe("localhost:20332", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
