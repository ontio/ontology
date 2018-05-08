package server

import "github.com/ontio/ontology/cmd/server/handlers"

func init(){
	DefCliRpcSvr.RegHandler("sigrawtx",handlers.SigRawTransaction)
	DefCliRpcSvr.RegHandler("sigtransfertx",handlers.SigTransferTransaction)
	DefCliRpcSvr.RegHandler("siginvoketx",handlers.SigNeoVMInvokeTx)
}