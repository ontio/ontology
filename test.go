package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/http/base/rpc"
	"github.com/ontio/ontology/smartcontract/states"
	vmtypes "github.com/ontio/ontology/smartcontract/types"
	"os"
	"time"
)

var (
	code = `51c56b616168144e656f2e436f6e74726163742e44657374726f79610b68656c6c6f2c776f726c646c766b00527ac46203006c766b00c3616c7566`
)

func main() {
	fmt.Println("Start Test Neo Smart Contract")

	b, _ := common.HexToBytes(code)

	tx := utils.NewDeployTransaction(vmtypes.VmCode{Code: b, VmType: vmtypes.NEOVM}, "ONG", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONG Token", true)

	tx.GasLimit = 10000000
	tx.GasPrice = 10000000
	tx.Nonce = uint32(time.Now().Unix())
	RequestPreExecute(tx)

	str := `8063804cf405124dbb06f79143da927aa00602a6`
	bs, _ := common.HexToBytes(str)
	ads, _ := common.AddressParseFromBytes(bs)
	c := states.Contract{
		Address: ads,
		Args:    []byte{},
	}
	bf := new(bytes.Buffer)
	c.Serialize(bf)

	//c1 := new(states.Contract)
	//c1.Deserialize(bf)
	//fmt.Printf("contract:%+v", c1)
	tx = utils.NewInvokeTransaction(vmtypes.VmCode{Code: append([]byte{103}, bf.Bytes()...), VmType: vmtypes.NEOVM})
	tx.GasLimit = 10000000
	tx.GasPrice = 10000000
	tx.Nonce = uint32(time.Now().Unix())

	Request(tx)
}

func RpcAddress() string {
	address := "http://localhost:20336"
	return address
}

func Request(tx *types.Transaction) {
	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		os.Exit(1)
	}
	resp, err := rpc.Call(RpcAddress(), "sendrawtransaction", 0,
		[]interface{}{hex.EncodeToString(txbf.Bytes())})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(resp)
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
	}
	switch r["result"].(type) {
	case map[string]interface{}:
		fmt.Println(r["result"])
	case string:
		fmt.Println(r["result"].(string))
	}
}

func RequestPreExecute(tx *types.Transaction) {
	txbf := new(bytes.Buffer)
	if err := tx.Serialize(txbf); err != nil {
		fmt.Println("Serialize transaction error.")
		os.Exit(1)
	}
	resp, err := rpc.Call(RpcAddress(), "sendrawtransaction", 0,
		[]interface{}{hex.EncodeToString(txbf.Bytes()), 1})

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(resp)
	r := make(map[string]interface{})
	err = json.Unmarshal(resp, &r)
	if err != nil {
		fmt.Println("Unmarshal JSON failed")
	}
	fmt.Println(r["result"])
	switch r["result"].(type) {
	case map[string]interface{}:
		fmt.Println(r["result"])
	case string:
		fmt.Println(r["result"].(string))
	}
}
