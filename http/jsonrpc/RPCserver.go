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
	//HandleFunc("getunspendoutput", getUnspendOutput)
	HandleFunc("getconnectioncount", GetConnectionCount)
	HandleFunc("getrawmempool", GetRawMemPool)
	HandleFunc("getmempooltx", GetMemPoolTx)
	HandleFunc("getrawtransaction", GetRawTransaction)
	HandleFunc("sendrawtransaction", SendRawTransaction)
	HandleFunc("getstorage", GetStorage)
	HandleFunc("getbalance", GetBalance)
	HandleFunc("submitblock", SubmitBlock)
	HandleFunc("getversion", GetNodeVersion)
	HandleFunc("getdataile", GetDataFile)
	HandleFunc("catdatarecord", CatDataRecord)
	HandleFunc("regdatafile", RegDataFile)
	HandleFunc("uploadDataFile", UploadDataFile)
	HandleFunc("getsmartcodeevent", GetSmartCodeEvent)

	err := http.ListenAndServe(":"+strconv.Itoa(Parameters.HttpJsonPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
