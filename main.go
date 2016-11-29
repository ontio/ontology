package main

import (
	"GoOnchain/net/httpjsonrpc"
	"GoOnchain/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func rpcCallDeprecated(nodeAddr string, rpcCommand map[string]interface{}) {

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

func rpcCall(nodeAddr string, rpcCommand map[string]interface{}) {
	// sendToAddress := &SendToAddress{
	// 	cmd: "sendtoaddress",
	// 	id: 1,
	// 	params: []interface{} {"asset_id", "address", "value"},
	// }
	//data, err := json.Marshal(sendToAddress)

	//data, err := json.Marshal(map[string]interface{} (getinfo))
	data, err := json.Marshal(rpcCommand)
	if err != nil {
		log.Fatalf("Marshal: %v", err)
	}

	resp, err := http.Post(nodeAddr, "application/json", strings.NewReader(string(data)))
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

func getPeerInfo() {
	seedNodes := config.SeedNodes()

	getinfo := map[string]interface{}{"method": "getinfo", "id": 1, "params": "test"}
	for i, node := range seedNodes {
		fmt.Printf("This seed node %d is %s\n", i, node)
		rpcCall(node, getinfo)

	}
}

func init() {
	go httpjsonrpc.StartServer()
}


func main() {
	time.Sleep(2 * time.Second)

	//getPeerInfo()
	httpjsonrpc.StartClient()
}
