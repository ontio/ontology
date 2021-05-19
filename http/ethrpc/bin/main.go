package main

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ontio/ontology/http/ethrpc"
)

func main() {

	startEthRpc()
}

func Ensure(err error) {
	if err != nil {
		panic(err)
	}
}

func startEthRpc() {
	calculator := new(ethrpc.EthereumAPI)
	server := rpc.NewServer()
	err := server.RegisterName("eth", calculator)
	Ensure(err)
	netRpcService := new(ethrpc.PublicNetAPI)
	err = server.RegisterName("net", netRpcService)
	Ensure(err)
	fmt.Printf("listen on 8545")
	err = http.ListenAndServe("0.0.0.0:8545", server)
	Ensure(err)
}
