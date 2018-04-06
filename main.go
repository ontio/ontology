package main

import (
	"github.com/Ontology/account"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/consensus"
	"github.com/Ontology/core/ledger"
	ldgactor "github.com/Ontology/core/ledger/actor"
	"github.com/Ontology/crypto"
	"github.com/Ontology/events"
	hserver "github.com/Ontology/http/base/actor"
	"github.com/Ontology/http/jsonrpc"
	"github.com/Ontology/http/localrpc"
	//"github.com/Ontology/http/nodeinfo"
	"github.com/Ontology/http/restful"
	"github.com/Ontology/http/websocket"
	"github.com/Ontology/p2pserver"
	netreqactor "github.com/Ontology/p2pserver/actor/req"
	p2pactor "github.com/Ontology/p2pserver/actor/server"
	nettypes "github.com/Ontology/p2pserver/common"
	"github.com/Ontology/txnpool"
	tc "github.com/Ontology/txnpool/common"
	"github.com/Ontology/validator/statefull"
	"github.com/Ontology/validator/stateless"
	"os"
	"os/signal"
	"runtime"
	"sort"
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
	var err error
	log.Trace("Node version: ", config.Version)

	if len(config.Parameters.BookKeepers) < account.DefaultBookKeeperCount {
		log.Fatal("At least ", account.DefaultBookKeeperCount, " BookKeepers should be set at config.json")
		os.Exit(1)
	}
	crypto.SetAlg(config.Parameters.EncryptAlg)

	log.Info("0. Open the account")
	client := account.GetClient()
	if client == nil {
		log.Fatal("Can't get local account.")
		os.Exit(1)
	}
	acct, err = client.GetDefaultAccount()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	log.Debug("The Node's PublicKey ", acct.PublicKey)
	defBookKeepers, err := client.GetBookKeepers()
	sort.Sort(crypto.PubKeySlice(defBookKeepers))
	if err != nil {
		log.Fatalf("GetBookKeepers error:%s", err)
		os.Exit(1)
	}

	//Init event hub
	events.Init()

	log.Info("1. Loading the Ledger")
	ledger.DefLedger, err = ledger.NewLedger()
	if err != nil {
		log.Fatalf("NewLedger error %s", err)
		os.Exit(1)
	}
	err = ledger.DefLedger.Init(defBookKeepers)
	if err != nil {
		log.Fatalf("DefLedger.Init error %s", err)
		os.Exit(1)
	}
	ldgerActor := ldgactor.NewLedgerActor()
	ledgerPID := ldgerActor.Start()

	log.Info("3. Start the transaction pool server")
	// Start the transaction pool server
	txPoolServer := txnpool.StartTxnPoolServer()
	if txPoolServer == nil {
		log.Fatalf("failed to start txn pool server")
		os.Exit(1)
	}

	stlValidator, _ := stateless.NewValidator("stateless_validator")
	stlValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))

	stfValidator, _ := statefull.NewValidator("statefull_validator")
	stfValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))

	log.Info("4. Start the P2P networks")

	p2p, err := p2pserver.NewServer(acct)
	if err != nil {
		log.Fatalf("p2pserver NewServer error %s", err)
		os.Exit(1)
	}
	err = p2p.Start(true)
	if err != nil {
		log.Fatalf("p2p sevice start error %s", err)
		os.Exit(1)
	}
	p2pActor := p2pactor.NewP2PActor()
	p2pPid, err := p2pActor.Start(p2p)
	if err != nil {
		log.Fatalf("p2pActor init error %s", err)
		os.Exit(1)
	}
	netreqactor.SetLedgerPid(ledgerPID)
	netreqactor.SetTxnPoolPid(txPoolServer.GetPID(tc.TxPoolActor))

	hserver.SetP2pPid(p2pPid)
	hserver.SetLedgerPid(ledgerPID)
	hserver.SetTxnPoolPid(txPoolServer.GetPID(tc.TxPoolActor))
	hserver.SetTxPid(txPoolServer.GetPID(tc.TxActor))
	go restful.StartServer()

	p2p.WaitForPeersStart()
	p2p.WaitForSyncBlkFinish()

	if nettypes.SERVICE_NODE_NAME != config.Parameters.NodeType {
		log.Info("5. Start Consensus Services")
		pool := txPoolServer.GetPID(tc.TxPoolActor)
		consensusService, _ := consensus.NewConsensusService(acct, pool, nil, p2pPid)
		netreqactor.SetConsensusPid(consensusService.GetPID())
		go consensusService.Start()
		time.Sleep(5 * time.Second)
		hserver.SetConsensusPid(consensusService.GetPID())
		go localrpc.StartLocalServer()
	}

	log.Info("--Start the RPC interface")
	go jsonrpc.StartRPCServer()
	go websocket.StartServer()
	// if config.Parameters.HttpInfoStart {
	// 	go nodeinfo.StartServer(noder)
	// }

	go logCurrBlockHeight()

	//等待退出信号
	waitToExit()
}

func logCurrBlockHeight() {
	ticker := time.NewTicker(config.DEFAULTGENBLOCKTIME * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Infof("BlockHeight = %d", ledger.DefLedger.GetCurrentBlockHeight())
			isNeedNewFile := log.CheckIfNeedNewFile()
			if isNeedNewFile {
				log.ClosePrintLog()
				log.Init(log.Path, os.Stdout)
			}
		}
	}
}

func waitToExit() {
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
}
