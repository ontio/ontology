package main

import (
	. "DNA/client"
	"DNA/common/log"
	"DNA/config"
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
	ledger.DefaultLedger.Store = store.NewLedgerStore()
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	crypto.SetAlg(crypto.P256R1)
	fmt.Println("  Client set completed. Test Start...")
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 1. Generate [Account]                                              ***")
	fmt.Println("//**************************************************************************")
	localclient := OpenClientAndGetAccount()
	if localclient == nil {
		fmt.Println("Can't get local client.")
		os.Exit(1)
	}

	issuer, err := localclient.GetDefaultAccount()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("//**************************************************************************")
	fmt.Println("//*** 2. Set Miner                                                     ***")
	fmt.Println("//**************************************************************************")
	miner := []*crypto.PubKey{}
	miner = append(miner, getMiner1().PublicKey)
	miner = append(miner, getMiner4().PublicKey)
	miner = append(miner, getMiner3().PublicKey)
	miner = append(miner, getMiner2().PublicKey)
	ledger.StandbyMiners = miner
	fmt.Println("miner1.PublicKey", issuer.PublicKey)

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
		log.Debug("ledger.DefaultLedger.Blockchain.BlockHeight= ", ledger.DefaultLedger.Blockchain.BlockHeight)
		time.Sleep(dbft.GenBlockTime)
	}
}

func OpenClientAndGetAccount() Client {
	clientName := config.Parameters.MinerName
	fmt.Printf("The Miner name is %s\n", clientName)
	if clientName == "" {
		fmt.Printf("Miner name not be set at config file protocol.json, which schould be c1,c2,c3,c4. Now is %s\n", clientName)
		return nil
	}
	var c1 Client
	var c2 Client
	var c3 Client
	var c4 Client

	if fileExisted("wallet1.txt") {
		c1 = OpenClient("wallet1.txt", []byte("\x12\x34\x56"))
	} else {
		c1 = CreateClient("wallet1.txt", []byte("\x12\x34\x56"))
	}

	if fileExisted("wallet2.txt") {
		c2 = OpenClient("wallet2.txt", []byte("\x12\x34\x56"))
	} else {
		c2 = CreateClient("wallet2.txt", []byte("\x12\x34\x56"))
	}

	if fileExisted("wallet3.txt") {
		c3 = OpenClient("wallet3.txt", []byte("\x12\x34\x56"))
	} else {
		c3 = CreateClient("wallet3.txt", []byte("\x12\x34\x56"))
	}

	if fileExisted("wallet4.txt") {
		c4 = OpenClient("wallet4.txt", []byte("\x12\x34\x56"))
	} else {
		c4 = CreateClient("wallet4.txt", []byte("\x12\x34\x56"))
	}

	var c Client
	if fileExisted("wallet.txt") {
		c = OpenClient("wallet.txt", []byte("\x12\x34\x56"))
	} else {
		c = CreateClient("wallet.txt", []byte("\x12\x34\x56"))
	}

	switch clientName {
	case "c1":
		return c1
	case "c2":
		return c2
	case "c3":
		return c3
	case "c4":
		return c4
	case "c":
		return c
	default:
		fmt.Printf("Please Check your client's ENV SET, if you are standby miners schould be c1,c2,c3,c4.If not, should be c. Now is %s.\n", clientName)
		return nil
	}
}

func getMiner1() *Account {
	c4 := OpenClient("wallet1.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
func getMiner2() *Account {
	c4 := OpenClient("wallet2.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
func getMiner3() *Account {
	c4 := OpenClient("wallet3.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
func getMiner4() *Account {
	c4 := OpenClient("wallet4.txt", []byte("\x12\x34\x56"))
	account, err := c4.GetDefaultAccount()
	if err != nil {
		fmt.Println("GetDefaultAccount failed.")
	}
	return account

}
