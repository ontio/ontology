package main

import (
	"GoOnchain/httpjsonrpc"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func trace() {
	pc := make([]uintptr, 10)
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	fmt.Printf("%s:%d %s\n", file, line, f.Name())
}

func HelloServer(w http.ResponseWriter, req *http.Request) {
	trace()
	fmt.Fprintf(w, "Hello,"+req.URL.Path[1:])
}

type GetBalance struct {
	assetId string
}

func getBalance(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getBestBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()

	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getInfo(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()

	//if err := json.Unmarshal(byt, &dat); err != nil {
	//	panic(err)
	//}

	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]

	// for range params
	//	param1 := params[0]
	fmt.Println(params)

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
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getBlockCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type GetBlockHash struct {
	index int //FIXME the maxium index overflow int?
}

func getBlockHash(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getConnectionCount(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func getRawMemPool(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

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
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

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
	trace()
	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type SendRawTransaction struct {
	hex interface{}
}

func sendRawTransaction(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()

	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

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
	trace()

	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

type SubmitBlock struct {
	hex interface{}
}

func submitBlock(req *http.Request, cmd map[string]interface{}) map[string]interface{} {
	trace()

	id := cmd["id"].(float64)
	fmt.Println(id)

	params := cmd["params"]
	fmt.Println(params)
	//param1 := params[0]
	//fmt.Println(param1)

	if cmd == nil {
		cmd = make(map[string]interface{})
	}

	return cmd
}

func start_server() {
	print("hello start sever\n")
	httpjsonrpc.InitServeMux()
	http.HandleFunc("/", httpjsonrpc.Handle)
	httpjsonrpc.HandleFunc("getinfo", getInfo)
	httpjsonrpc.HandleFunc("sendtoaddress", sendToAddress)
	err := http.ListenAndServe("localhost:10332", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}

func start_client() {
	var res map[string]interface{}
	var err error
	print("start client\n")

	// Call the get info
	res, err = httpjsonrpc.Call("http://127.0.0.1:10332", "getinfo", 1, []interface{}{})
	if err != nil {
		log.Fatalf("Err: %v", err)
	}
	log.Println(res)

	// call send to address
	params := []interface{}{"asset_id", "address", 56}
	res, err = httpjsonrpc.Call("http://127.0.0.1:10332", "sendtoaddress", 2, params)
	if err != nil {
		log.Fatalf("Err: %v", err)
	}
	log.Println(res)

}

func init() {
	go start_server()
}

func main() {
	time.Sleep(2 * time.Second)
	start_client()
}

func tmp() {

	getinfo := map[string]interface{}{"method": "getinfo", "id": 1, "params": "test"}

	// sendToAddress := &SendToAddress{
	// 	cmd: "sendtoaddress",
	// 	id: 1,
	// 	params: []interface{} {"asset_id", "address", "value"},
	// }
	//data, err := json.Marshal(sendToAddress)

	//data, err := json.Marshal(map[string]interface{} (getinfo))
	data, err := json.Marshal(getinfo)

	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}
	resp, err := http.Post("http://user:pass@127.0.0.1:10332",
		"application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Post: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ReadAll: %v", err)
	}
	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	log.Println(result)
}
