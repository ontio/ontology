package main

import (
	"time"
	"runtime"
	"GoOnchain/net"
	"GoOnchain/net/httpjsonrpc"
	"GoOnchain/common/log"
)

const (
	// The number of the CPU cores for parallel optimization,TODO set from config file
	NCPU	 = 4
)

func init() {
	runtime.GOMAXPROCS(NCPU)
	var path string = "./Log/"
	log.CreatePrintLog(path)
}


func main() {
	time.Sleep(2 * time.Second)
	net.StartProtocol()


	go httpjsonrpc.StartServer()

	time.Sleep(2 * time.Second)
	httpjsonrpc.StartClient()

	// Modules start sample
	//ledger.Start(net.NetToLedgerCh <-chan *Msg, net.LedgerToNetCh chan<- *Msg)
	//consensus.Start(net.NetToConsensusCh <-chan *Msg, net.ConsensusToNetCh chan<- *Msg)

	for {
		time.Sleep(2 * time.Second)
	}
}
