package jsonrpc

import (
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/base/rpc"
	"net/http"
	"strconv"
)




func StartRPCServer() {
	log.Debug()
	http.HandleFunc("/", Handle)

	HandleFunc("getbestblockhash", GetBestBlockHash)
	HandleFunc("getblock", GetBlock)
	HandleFunc("getblockcount", GetBlockCount)
	HandleFunc("getblockhash", GetBlockHash)
	HandleFunc("getconnectioncount", GetConnectionCount)
	HandleFunc("getrawmempool", GetRawMemPool)

	HandleFunc("getrawtransaction", GetRawTransaction)
	HandleFunc("sendrawtransaction", SendRawTransaction)
	HandleFunc("getstorage", GetStorage)
	HandleFunc("getversion", GetNodeVersion)

	HandleFunc("getblocksysfee", GetSystemFee)
	HandleFunc("getcontractstate", GetContractState)
	HandleFunc("getmempooltxstate", GetMemPoolTxState)
	HandleFunc("getsmartcodeevent", GetSmartCodeEvent)
	HandleFunc("getblockheightbytxhash", GetBlockHeightByTxHash)

	HandleFunc("getbalance", GetBalance)

	err := http.ListenAndServe(":"+strconv.Itoa(Parameters.HttpJsonPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
