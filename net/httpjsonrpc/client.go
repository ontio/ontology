package httpjsonrpc

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"GoOnchain/common"
)

func Call(address string, method string, id interface{}, params []interface{}) (map[string]interface{}, error) {
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
	result := make(map[string]interface{})
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return nil, err
	}
	//log.Println(result)
	return result, nil
}

func StartClient() {
	var res map[string]interface{}
	var err error

	common.Trace()

	// Call the get info
	res, err = Call("http://127.0.0.1:20337", "getinfo", 1, []interface{}{})
	if err != nil {
		log.Fatalf("Err: %v", err)
	}
	log.Println(res)

	// call send to address
	params := []interface{}{"asset_id", "address", 56}
	res, err = Call("http://127.0.0.1:20337", "sendtoaddress", 2, params)
	if err != nil {
		log.Fatalf("Err: %v", err)
	}
	log.Println(res)
}
