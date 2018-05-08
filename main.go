/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/consensus"
	"github.com/ontio/ontology/core/ledger"
	ldgactor "github.com/ontio/ontology/core/ledger/actor"
	"github.com/ontio/ontology/events"
	hserver "github.com/ontio/ontology/http/base/actor"
	"github.com/ontio/ontology/http/jsonrpc"
	"github.com/ontio/ontology/http/localrpc"
	"github.com/ontio/ontology/http/nodeinfo"
	"github.com/ontio/ontology/http/restful"
	"github.com/ontio/ontology/http/websocket"
	"github.com/ontio/ontology/p2pserver"
	netreqactor "github.com/ontio/ontology/p2pserver/actor/req"
	p2pactor "github.com/ontio/ontology/p2pserver/actor/server"
	"github.com/ontio/ontology/txnpool"
	tc "github.com/ontio/ontology/txnpool/common"
	"github.com/ontio/ontology/txnpool/proc"
	"github.com/ontio/ontology/validator/stateful"
	"github.com/ontio/ontology/validator/stateless"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Ontology CLI"
	app.Action = startOntology
	app.Version = "0.7.0"
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Commands = []cli.Command{
		cmd.AccountCommand,
		cmd.InfoCommand,
		cmd.AssetCommand,
		cmd.ContractCommand,
	}
	app.Flags = []cli.Flag{
		//common setting
		utils.ConfigFlag,
		utils.LogLevelFlag,
		utils.WalletFileFlag,
		utils.AccountPassFlag,
		utils.DisableEventLogFlag,
		utils.MaxTxInBlockFlag,
		//p2p setting
		utils.NodePortFlag,
		utils.ConsensusPortFlag,
		utils.DualPortSupportFlag,
		//test mode setting
		utils.EnableTestModeFlag,
		utils.TestModeGenBlockTimeFlag,
		//rpc setting
		utils.RPCPortFlag,
		utils.RPCLocalEnableFlag,
		utils.RPCLocalProtFlag,
		//rest setting
		utils.RestfulEnableFlag,
		utils.RestfulPortFlag,
		//ws setting
		utils.WsEnabledFlag,
		utils.WsPortFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		log.Init(log.PATH, log.Stdout)
		return nil
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startOntology(ctx *cli.Context) {
	_, err := initConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error:%s", err)
		return
	}
	wallet, err := initWallet(ctx)
	if err != nil {
		log.Errorf("initWallet error:%s", err)
		return
	}
	ldg, err := initLedger(ctx)
	if err != nil {
		log.Errorf("%s", err)
		return
	}
	defer ldg.Close()
	txpool, err := initTxPool(ctx)
	if err != nil {
		log.Errorf("initTxPool error:%s", err)
		return
	}
	p2pSvr, p2pPid, err := initP2PNode(ctx, wallet, ldgactor.DefLedgerPid, txpool)
	if err != nil {
		log.Errorf("initP2PNode error:%s", err)
		return
	}
	_, err = initConsensus(ctx, p2pPid, txpool, wallet)
	if err != nil {
		log.Errorf("initConsensus error:%s", err)
		return
	}
	err = initRpc(ctx)
	if err != nil {
		log.Errorf("initRpc error:%s", err)
		return
	}
	err = initLocalRpc(ctx)
	if err != nil {
		log.Errorf("initLocalRpc error:%s", err)
		return
	}
	initRestful(ctx)
	initWs(ctx)
	initNodeInfo(ctx, p2pSvr)

	go logCurrBlockHeight()
	waitToExit()
}

func initConfig(ctx *cli.Context) (*config.OntologyConfig, error) {
	//init ontology config from cli
	cfg, err := cmd.SetOntologyConfig(ctx)
	if err != nil {
		return nil, err
	}
	log.Infof("Config init success")
	return cfg, nil
}

func initWallet(ctx *cli.Context) (*account.ClientImpl, error) {
	walletFile := ctx.GlobalString(utils.WalletFileFlag.Name)
	if walletFile == "" {
		return nil, fmt.Errorf("Please config wallet file using --wallet flag")
	}
	if !common.FileExisted(walletFile) {
		return nil, fmt.Errorf("Cannot find wallet file:%s. Please create wallet first", walletFile)
	}

	var pwd []byte = nil
	var err error
	if ctx.IsSet(utils.AccountPassFlag.Name) {
		pwd = []byte(ctx.GlobalString(utils.AccountPassFlag.Name))
	} else {
		pwd, err = password.GetAccountPassword()
		if err != nil {
			return nil, fmt.Errorf("Password error")
		}
	}
	client := account.Open(walletFile, pwd)
	if client == nil {
		return nil, fmt.Errorf("Cannot open wallet file:%s", walletFile)
	}

	acc := client.GetDefaultAccount()
	if acc == nil {
		return nil, fmt.Errorf("Cannot GetDefaultAccount")
	}

	curPk := hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey))

	switch config.DefConfig.Genesis.ConsensusType {
	case config.CONSENSUS_TYPE_DBFT:
		isBookKeeper := false
		for _, pk := range config.DefConfig.Genesis.DBFT.Bookkeepers {
			if pk == curPk {
				isBookKeeper = true
				break
			}
		}
		if !isBookKeeper {
			config.DefConfig.Common.EnableConsensus = false
		}
	case config.CONSENSUS_TYPE_SOLO:
		config.DefConfig.Genesis.SOLO.Bookkeepers = []string{curPk}
	}

	log.Infof("Wallet init success")
	return client, nil
}

func initLedger(ctx *cli.Context) (*ledger.Ledger, error) {
	events.Init() //Init event hub

	var err error
	ledger.DefLedger, err = ledger.NewLedger()
	if err != nil {
		return nil, fmt.Errorf("NewLedger error:%s", err)
	}
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return nil, fmt.Errorf("GetBookkeepers error:%s", err)
	}
	err = ledger.DefLedger.Init(bookKeepers)
	if err != nil {
		return nil, fmt.Errorf("Init ledger error:%s", err)
	}
	ldgactor.NewLedgerActor().Start()

	hserver.SetLedgerPid(ldgactor.DefLedgerPid)

	log.Infof("Ledger init success")
	return ledger.DefLedger, nil
}

func initTxPool(ctx *cli.Context) (*proc.TXPoolServer, error) {
	txPoolServer, err := txnpool.StartTxnPoolServer()
	if err != nil {
		return nil, fmt.Errorf("Init txpool error:%s", err)
	}
	stlValidator, _ := stateless.NewValidator("stateless_validator")
	stlValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))
	stfValidator, _ := stateful.NewValidator("stateful_validator")
	stfValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))

	hserver.SetTxnPoolPid(txPoolServer.GetPID(tc.TxPoolActor))
	hserver.SetTxPid(txPoolServer.GetPID(tc.TxActor))

	log.Infof("TxPool init success")
	return txPoolServer, nil
}

func initP2PNode(ctx *cli.Context, wallet *account.ClientImpl, ledgerPid *actor.PID, txpoolSvr *proc.TXPoolServer) (*p2pserver.P2PServer, *actor.PID, error) {
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		return nil, nil, nil
	}
	acc := wallet.GetDefaultAccount()
	if acc == nil {
		return nil, nil, fmt.Errorf("Cannot GetDefaultAccount")
	}
	p2p, err := p2pserver.NewServer(acc)
	if err != nil {
		return nil, nil, fmt.Errorf("P2P node NewServer error:%s", err)
	}
	p2pActor := p2pactor.NewP2PActor(p2p)
	p2pPID, err := p2pActor.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("p2pActor init error %s", err)
	}
	p2p.SetPID(p2pPID)
	err = p2p.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("p2p sevice start error %s", err)
	}
	netreqactor.SetLedgerPid(ledgerPid)
	netreqactor.SetTxnPoolPid(txpoolSvr.GetPID(tc.TxActor))
	txpoolSvr.RegisterActor(tc.NetActor, p2pPID)
	hserver.SetNetServerPID(p2pPID)

	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_VBFT {
		return p2p, p2pPID, nil
	}
	p2p.WaitForPeersStart()
	p2p.WaitForSyncBlkFinish()

	log.Infof("P2P node init success")
	return p2p, p2pPID, nil
}

func initConsensus(ctx *cli.Context, p2pPid *actor.PID, txpoolSvr *proc.TXPoolServer, wallet *account.ClientImpl) (consensus.ConsensusService, error) {
	if !config.DefConfig.Common.EnableConsensus {
		return nil, nil
	}
	acc := wallet.GetDefaultAccount()
	if acc == nil {
		return nil, fmt.Errorf("GetDefaultAccount failed")
	}
	pool := txpoolSvr.GetPID(tc.TxPoolActor)

	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	consensusService, err := consensus.NewConsensusService(consensusType, acc, pool, nil, p2pPid)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusService:%s error:%s", consensusType, err)
	}
	netreqactor.SetConsensusPid(consensusService.GetPID())
	hserver.SetConsensusPid(consensusService.GetPID())

	go consensusService.Start()

	log.Infof("Consensus init success")
	return consensusService, nil
}

func initRpc(ctx *cli.Context) error {
	var err error
	exitCh := make(chan interface{}, 0)
	go func() {
		err = jsonrpc.StartRPCServer()
		close(exitCh)
	}()

	flag := false
	select {
	case <-exitCh:
		if !flag {
			return err
		}
	case <-time.After(time.Millisecond * 5):
		flag = true
	}
	log.Infof("Rpc init success")
	return nil
}

func initLocalRpc(ctx *cli.Context) error {
	if !ctx.GlobalBool(utils.RPCLocalEnableFlag.Name) {
		return nil
	}
	var err error
	exitCh := make(chan interface{}, 0)
	go func() {
		err = localrpc.StartLocalServer()
		close(exitCh)
	}()

	flag := false
	select {
	case <-exitCh:
		if !flag {
			return err
		}
	case <-time.After(time.Millisecond * 5):
		flag = true
	}

	log.Infof("Local rpc init success")
	return nil
}

func initRestful(ctx *cli.Context) {
	if !ctx.GlobalBool(utils.RestfulEnableFlag.Name) {
		return
	}
	go restful.StartServer()

	log.Infof("Restful init success")
}

func initWs(ctx *cli.Context) {
	if !ctx.GlobalBool(utils.WsEnabledFlag.Name) {
		return
	}
	websocket.StartServer()

	log.Infof("Ws init success")
}

func initNodeInfo(ctx *cli.Context, p2pSvr *p2pserver.P2PServer) {
	if config.DefConfig.P2PNode.HttpInfoPort == 0 {
		return
	}
	go nodeinfo.StartServer(p2pSvr.GetNetWork())

	log.Infof("Nodeinfo init success")
}

func logCurrBlockHeight() {
	ticker := time.NewTicker(config.DEFAULT_GEN_BLOCK_TIME * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Infof("CurrentBlockHeight = %d", ledger.DefLedger.GetCurrentBlockHeight())
			isNeedNewFile := log.CheckIfNeedNewFile()
			if isNeedNewFile {
				log.ClosePrintLog()
				log.Init(log.PATH, os.Stdout)
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
