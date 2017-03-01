package main

import (
	"GoOnchain/common/log"
	"GoOnchain/core/ledger"
	"GoOnchain/core/store"
	"GoOnchain/core/transaction"
	"GoOnchain/crypto"
	"GoOnchain/net"
	"GoOnchain/net/httpjsonrpc"
	"fmt"
	"runtime"
	"time"
	//"GoOnchain/consensus/dbft"
)

const (
	// The number of the CPU cores for parallel optimization,TODO set from config file
	NCPU = 4
)

var Version string

func init() {
	runtime.GOMAXPROCS(NCPU)
	var path string = "./Log/"
	log.CreatePrintLog(path)
}

func main() {
	fmt.Printf("Node version: %s\n", Version)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 0. Client Set                                                      ***")
	fmt.Println("//**************************************************************************")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store = store.NewLedgerStore()
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	crypto.SetAlg(crypto.P256R1)
	fmt.Println("  Client set completed. Test Start...")

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 1. BlockChain init                                                 ***")
	fmt.Println("//**************************************************************************")
	//blockchain :=
	fmt.Println("  BlockChain generate completed. Func test Start...")
	ledger.DefaultLedger.Blockchain, _ = ledger.NewBlockchainWithGenesisBlock()

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 2. Generate Account                                                ***")
	fmt.Println("//**************************************************************************")

	time.Sleep(2 * time.Second)
	net.StartProtocol()

	go httpjsonrpc.StartServer()

	time.Sleep(2 * time.Second)
	//httpjsonrpc.StartClient()

	// Modules start sample
	//ledger.Start(net.NetToLedgerCh <-chan *Msg, net.LedgerToNetCh chan<- *Msg)
	//consensus.Start(net.NetToConsensusCh <-chan *Msg, net.ConsensusToNetCh chan<- *Msg)
	//consensus := new(dbft.DbftService)

	for {
		time.Sleep(2 * time.Second)
	}
}
