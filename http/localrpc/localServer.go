package localrpc

import (
	. "github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	. "github.com/Ontology/http/common/rpc"
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

	HandleFunc("getneighbor", GetNeighbor)
	HandleFunc("getnodestate", GetNodeState)
	HandleFunc("startconsensus", StartConsensus)
	HandleFunc("stopconsensus", StopConsensus)
	HandleFunc("sendsampletransaction", SendSampleTransaction)
	HandleFunc("setdebuginfo", SetDebugInfo)

	// TODO: only listen to local host
	err := http.ListenAndServe(":"+strconv.Itoa(Parameters.HttpLocalPort), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
