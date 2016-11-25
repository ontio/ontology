package httpjsonrpc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
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
