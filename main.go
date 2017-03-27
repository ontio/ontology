package main

import (
	. "DNA/client"
	"DNA/common/log"
	"DNA/consensus/dbft"
	"DNA/core/ledger"
	"DNA/core/store"
	"DNA/core/transaction"
	"DNA/crypto"
	"DNA/net"
	"DNA/net/httpjsonrpc"
	"fmt"
	"os"
	"runtime"
	"time"
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

func fileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func openLocalClient(name string) Client {
	var c Client

	if fileExisted(name) {
		c = OpenClient(name, []byte("\x12\x34\x56"))
	} else {
		c = CreateClient(name, []byte("\x12\x34\x56"))
	}

	return c
}

func main() {
	fmt.Printf("Node version: %s\n", Version)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 0. Client open                                                     ***")
	fmt.Println("//**************************************************************************")
	crypto.SetAlg(crypto.P256R1)
	fmt.Println("  Client set completed. Test Start...")
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 1. Generate [Account]                                              ***")
	fmt.Println("//**************************************************************************")
	localclient := openLocalClient("wallet.txt")
	if localclient == nil {
		fmt.Println("Can't get local client.")
		os.Exit(1)
	}
	account, err := localclient.GetDefaultAccount()
	if err != nil {
		fmt.Println("Can't get default account.")
		os.Exit(1)
	}
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 2. Ledger init                                                     ***")
	fmt.Println("//**************************************************************************")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store = store.NewLedgerStore()
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	ledger.DefaultLedger.Blockchain = ledger.NewBlockchain()

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 3. Start Networking Services                                       ***")
	fmt.Println("//**************************************************************************")
	neter, noder := net.StartProtocol(account.PublicKey)
	httpjsonrpc.RegistRpcNode(noder)
	time.Sleep(20 * time.Second)
	miners, _ := neter.GetMinersAddrs()
	ledger.CreateGenesisBlock(miners)

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 4. Start DBFT Services                                             ***")
	fmt.Println("//**************************************************************************")
	dbftServices := dbft.NewDbftService(localclient, "logdbft", neter)
	httpjsonrpc.RegistDbftService(dbftServices)
	go dbftServices.Start()
	time.Sleep(5 * time.Second)
	fmt.Println("DBFT Services start completed.")
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** Init Complete                                                      ***")
	fmt.Println("//**************************************************************************")
	go httpjsonrpc.StartRPCServer()
	go httpjsonrpc.StartLocalServer()

	time.Sleep(2 * time.Second)

	for {
		log.Debug("ledger.DefaultLedger.Blockchain.BlockHeight= ", ledger.DefaultLedger.Blockchain.BlockHeight)
		time.Sleep(dbft.GenBlockTime)
	}
}
