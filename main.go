package main

import (
	. "DNA/client"
	"DNA/common/config"
	"DNA/common/log"
	"DNA/consensus/dbft"
	"DNA/core/ledger"
	"DNA/core/store/ChainStore"
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
	NCPU                   = 4
	DefaultBookKeeperCount = 4
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

func InitBlockChain() ledger.Blockchain {
	blockchain, err := ledger.NewBlockchainWithGenesisBlock()
	if err != nil {
		fmt.Println(err, "  BlockChain generate failed")
	}
	fmt.Println("  BlockChain generate completed. Func test Start...")
	return *blockchain
}

func main() {
	fmt.Printf("Node version: %s\n", Version)
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 0. Client open                                                     ***")
	fmt.Println("//**************************************************************************")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store = ChainStore.NewLedgerStore()
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	crypto.SetAlg(crypto.P256R1)
	fmt.Println("  Client set completed. Test Start...")
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 1. Generate [Account]                                              ***")
	fmt.Println("//**************************************************************************")
	var bookKeeperCount uint32 = DefaultBookKeeperCount
	if config.Parameters.BookKeeperCount != 0 {
		bookKeeperCount = config.Parameters.BookKeeperCount
	}
	localclient := OpenClientAndGetAccount(bookKeeperCount)
	if localclient == nil {
		fmt.Println("Can't get local client.")
		os.Exit(1)
	}

	issuer, err := localclient.GetDefaultAccount()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 2. Set BookKeeper                                                     ***")
	fmt.Println("//**************************************************************************")
	bookKeeper := []*crypto.PubKey{}
	var i uint32
	for i = 0; i < bookKeeperCount; i++ {
		bookKeeper = append(bookKeeper, getBookKeeper(i+1).PublicKey)
	}
	ledger.StandbyBookKeepers = bookKeeper
	fmt.Println("bookKeeper1.PublicKey", issuer.PublicKey)

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 3. BlockChain init                                                 ***")
	fmt.Println("//**************************************************************************")
	sampleBlockchain := InitBlockChain()
	ledger.DefaultLedger.Blockchain = &sampleBlockchain

	time.Sleep(2 * time.Second)
	neter, noder := net.StartProtocol(issuer.PublicKey)
	httpjsonrpc.RegistRpcNode(noder)
	time.Sleep(20 * time.Second)

	noder.LocalNode().SyncNodeHeight()

	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 5. Start DBFT Services                                             ***")
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
		log.Trace("ledger.DefaultLedger.Blockchain.BlockHeight= ", ledger.DefaultLedger.Blockchain.BlockHeight)
		time.Sleep(dbft.GenBlockTime)
	}
}

func OpenClientAndGetAccount(count uint32) Client {
	clientName := config.Parameters.BookKeeperName
	fmt.Printf("The BookKeeper name is %s\n", clientName)
	if clientName == "" {
		log.Error("BookKeeper name not be set at config.json")
		return nil
	}
	var c []Client
	c = make([]Client, count)
	var i uint32
	for i = 1; i <= count; i++ {
		w := fmt.Sprintf("wallet%d.txt", i)
		if fileExisted(w) {
			c[i-1] = OpenClient(w, []byte("\x12\x34\x56"))
		} else {
			c[i-1] = CreateClient(w, []byte("\x12\x34\x56"))
		}
	}
	var n uint32
	fmt.Sscanf(clientName, "c%d", &n)
	return c[n-1]
}

func getBookKeeper(n uint32) *Account {
	w := fmt.Sprintf("wallet%d.txt", n)
	c := OpenClient(w, []byte("\x12\x34\x56"))
	account, err := c.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account
}
