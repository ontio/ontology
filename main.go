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
	"DNA/net/protocol"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	DefaultBookKeeperCount = 4
	DefaultMultiCoreNum  = 4
)

func init() {
	var path string = "./Log/"
	log.CreatePrintLog(path)

	var coreNum int
	if (config.Parameters.MultiCoreNum > DefaultMultiCoreNum) {
		coreNum = int(config.Parameters.MultiCoreNum)
	} else {
		coreNum = DefaultMultiCoreNum
	}
	log.Debug("The Core number is ", coreNum)
	runtime.GOMAXPROCS(coreNum)
}

func fileExisted(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func main() {
	log.Trace("Node version: ", config.Version)
	log.Info("0. Loading the Ledger")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store = ChainStore.NewLedgerStore()
	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	transaction.TxStore = ledger.DefaultLedger.Store
	crypto.SetAlg(config.Parameters.EncryptAlg)

	log.Info("1. Open the account")
	localClient, nodeType := OpenClientAndGetAccount()
	if localClient == nil {
		log.Error("Can't get local account.")
		os.Exit(1)
	}
	issuer, err := localClient.GetDefaultAccount()
	if err != nil {
		log.Error(err)
	}

	log.Info("2. Set the BookKeeper")
	bookKeeper := []*crypto.PubKey{}
	var i uint32
	for i = 0; i < DefaultBookKeeperCount; i++ {
		bookKeeper = append(bookKeeper, getBookKeeper(i+1).PublicKey)
	}
	ledger.StandbyBookKeepers = bookKeeper
	log.Debug("bookKeeper1.PublicKey", issuer.PublicKey)

	log.Info("3. BlockChain init")
	sampleBlockchain := InitBlockChain()
	ledger.DefaultLedger.Blockchain = &sampleBlockchain

	log.Info("4. Start the P2P networks")
	time.Sleep(2 * time.Second)
	neter, noder := net.StartProtocol(issuer.PublicKey, nodeType)
	httpjsonrpc.RegistRpcNode(noder)
	time.Sleep(20 * time.Second)
	noder.SyncNodeHeight()
	noder.WaitForFourPeersStart()
	if protocol.IsNodeTypeVerify(nodeType) {
		log.Info("5. Start DBFT Services")
		dbftServices := dbft.NewDbftService(localClient, "logdbft", neter)
		httpjsonrpc.RegistDbftService(dbftServices)
		go dbftServices.Start()
		time.Sleep(5 * time.Second)
	}

	log.Info("--Start the RPC interface")
	go httpjsonrpc.StartRPCServer()
	go httpjsonrpc.StartLocalServer()

	time.Sleep(2 * time.Second)
	for {
		log.Trace("BlockHeight = ", ledger.DefaultLedger.Blockchain.BlockHeight)
		time.Sleep(dbft.GenBlockTime)
	}
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
		log.Error(err, "  BlockChain generate failed")
	}
	return *blockchain
}

func clientIsDefaultBookKeeper(clientName string) bool {
	var i uint32
	for i = 0; i < DefaultBookKeeperCount; i++ {
		c := fmt.Sprintf("c%d", i+1)
		if strings.Compare(clientName, c) == 0 {
			return true
		}
	}
	return false
}

func OpenClientAndGetAccount() (clt Client, nodeType int) {
	if "service" == config.Parameters.NodeType {
		var sc Client
		walletFile := "wallet.txt"
		if fileExisted(walletFile) {
			sc = OpenClient(walletFile, []byte("\x12\x34\x56"))
		} else {
			sc = CreateClient(walletFile, []byte("\x12\x34\x56"))
		}
		return sc, protocol.GetServiceFlag()
	}

	clientName := config.Parameters.BookKeeperName
	log.Info("The BookKeeper name is ", clientName)
	if clientName == "" {
		log.Error("BookKeeper name not be set at config.json")
		return nil, protocol.GetVerifyFlag()
	}
	isDefaultBookKeeper := clientIsDefaultBookKeeper(clientName)
	var c []Client
	if isDefaultBookKeeper == true {
		c = make([]Client, DefaultBookKeeperCount)
	} else {
		c = make([]Client, DefaultBookKeeperCount+1)
	}
	var i uint32
	for i = 1; i <= DefaultBookKeeperCount; i++ {
		w := fmt.Sprintf("wallet%d.txt", i)
		if fileExisted(w) {
			c[i-1] = OpenClient(w, []byte("\x12\x34\x56"))
		} else {
			c[i-1] = CreateClient(w, []byte("\x12\x34\x56"))
		}
	}
	var n uint32
	fmt.Sscanf(clientName, "c%d", &n)
	if isDefaultBookKeeper == true {
		return c[n-1], protocol.GetVerifyFlag()
	}
	if isDefaultBookKeeper == false {
		w := fmt.Sprintf("wallet%d.txt", n)
		if fileExisted(w) {
			c[DefaultBookKeeperCount] = OpenClient(w, []byte("\x12\x34\x56"))
		} else {
			c[DefaultBookKeeperCount] = CreateClient(w, []byte("\x12\x34\x56"))
		}
	}
	return c[DefaultBookKeeperCount], protocol.GetVerifyFlag()
}

func getBookKeeper(n uint32) *Account {
	w := fmt.Sprintf("wallet%d.txt", n)
	c := OpenClient(w, []byte("\x12\x34\x56"))
	account, err := c.GetDefaultAccount()
	if err != nil {
		log.Error("GetDefaultAccount failed.")
	}
	return account
}
