package httpjsonrpc

import (
	"DNA/common/log"
	. "DNA/common/config"
	"net/http"
	"strconv"
)

const (
	localHost string = "127.0.0.1"
	LocalDir  string = "/local"
)

func StartLocalServer() {
	log.Debug()
	http.HandleFunc(LocalDir, Handle)

	HandleFunc("getneighbor", getNeighbor)
	HandleFunc("getnodestate", getNodeState)
	HandleFunc("startconsensus", startConsensus)
	HandleFunc("stopconsensus", stopConsensus)
	HandleFunc("sendsampletransaction", sendSampleTransaction)
	HandleFunc("setdebuginfo", setDebugInfo)

	// TODO: only listen to local host
	err := http.ListenAndServe(":"+strconv.Itoa(Parameters.HttpLocalPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
