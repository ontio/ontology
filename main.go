package main

import (
	"github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/consensus"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/store/ChainStore"
	"github.com/Ontology/crypto"
	"github.com/Ontology/net"
	"github.com/Ontology/http/jsonrpc"
	"github.com/Ontology/http/nodeinfo"
	"github.com/Ontology/http/restful"
	"github.com/Ontology/http/websocket"
	"github.com/Ontology/http/localrpc"
	"github.com/Ontology/net/protocol"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

const (
	DefaultMultiCoreNum = 4
)

func init() {
	log.Init(log.Path, log.Stdout)
	var coreNum int
	if config.Parameters.MultiCoreNum > DefaultMultiCoreNum {
		coreNum = int(config.Parameters.MultiCoreNum)
	} else {
		coreNum = DefaultMultiCoreNum
	}
	log.Debug("The Core number is ", coreNum)
	runtime.GOMAXPROCS(coreNum)
}

func main() {
	var acct *account.Account
	var blockChain *ledger.Blockchain
	var err error
	var noder protocol.Noder
	log.Trace("Node version: ", config.Version)

	if len(config.Parameters.BookKeepers) < account.DefaultBookKeeperCount {
		log.Fatal("At least ", account.DefaultBookKeeperCount, " BookKeepers should be set at config.json")
		os.Exit(1)
	}

	log.Info("0. Loading the Ledger")
	ledger.DefaultLedger = new(ledger.Ledger)
	ledger.DefaultLedger.Store, err = ChainStore.NewLedgerStore()
	if err != nil {
		log.Fatal("open LedgerStore err:", err)
		os.Exit(1)
	}
	defer ledger.DefaultLedger.Store.Close()

	ledger.DefaultLedger.Store.InitLedgerStore(ledger.DefaultLedger)
	crypto.SetAlg(config.Parameters.EncryptAlg)

	log.Info("1. Open the account")
	client := account.GetClient()
	if client == nil {
		log.Fatal("Can't get local account.")
		goto ERROR
	}
	acct, err = client.GetDefaultAccount()
	if err != nil {
		log.Fatal(err)
		goto ERROR
	}
	log.Debug("The Node's PublicKey ", acct.PublicKey)
	ledger.StandbyBookKeepers, err = client.GetBookKeepers()
	if err != nil {
		log.Fatalf("GetBookKeepers error:%s", err)
		goto ERROR
	}

	log.Info("3. BlockChain init")
	blockChain, err = ledger.NewBlockchainWithGenesisBlock(ledger.StandbyBookKeepers)
	if err != nil {
		log.Fatal(err, "  BlockChain generate failed")
		goto ERROR
	}
	ledger.DefaultLedger.Blockchain = blockChain

	log.Info("4. Start the P2P networks")
	// Don't need two return value.
	noder = net.StartProtocol(acct.PublicKey)
	go restful.StartServer(noder)
	jsonrpc.RegistRpcNode(noder)

	noder.SyncNodeHeight()
	noder.WaitForPeersStart()
	noder.WaitForSyncBlkFinish()
	if protocol.SERVICENODENAME != config.Parameters.NodeType {
		log.Info("5. Start Consensus Services")
		consensusSrv := consensus.ConsensusMgr.NewConsensusService(client, noder)
		jsonrpc.RegistConsensusService(consensusSrv)
		go consensusSrv.Start()
		time.Sleep(5 * time.Second)
	}

	log.Info("--Start the RPC interface")
	go jsonrpc.StartRPCServer()
	go localrpc.StartLocalServer()
	go websocket.StartServer(noder)
	if config.Parameters.HttpInfoStart {
		go nodeinfo.StartServer(noder)
	}

	log.Info("--Loading Event Store--")
	//ChainStore.NewEventStore()

	go func() {
		ticker := time.NewTicker(config.DEFAULTGENBLOCKTIME * time.Second)
		for {
			select {
			case <-ticker.C:
				log.Trace("BlockHeight = ", ledger.DefaultLedger.Blockchain.BlockHeight)
				isNeedNewFile := log.CheckIfNeedNewFile()
				if isNeedNewFile == true {
					log.ClosePrintLog()
					log.Init(log.Path, os.Stdout)
				}
			}
		}
	}()

	func() {
		//等待退出信号
		exit := make(chan bool, 0)
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		go func() {
			for sig := range sc {
				log.Infof("Ontology received exit signal:%v.", sig.String())
				close(exit)
				break
			}
		}()
		<-exit
	}()

ERROR:
	os.Exit(1)
}
