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
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-eventbus/actor"
	alog "github.com/ontio/ontology-eventbus/log"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/cmd"
	cmdcom "github.com/ontio/ontology/cmd/common"
	"github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/consensus"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/events"
	bactor "github.com/ontio/ontology/http/base/actor"
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
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "Ontology CLI"
	app.Action = startOntology
	app.Version = config.Version
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Commands = []cli.Command{
		cmd.AccountCommand,
		cmd.InfoCommand,
		cmd.AssetCommand,
		cmd.ContractCommand,
		cmd.ImportCommand,
		cmd.ExportCommand,
		cmd.TxCommond,
		cmd.SigTxCommand,
		cmd.MultiSigAddrCommand,
		cmd.MultiSigTxCommand,
		cmd.SendTxCommand,
		cmd.ShowTxCommand,
	}
	app.Flags = []cli.Flag{
		//common setting
		utils.ConfigFlag,
		utils.LogLevelFlag,
		utils.DisableEventLogFlag,
		utils.DataDirFlag,
		//account setting
		utils.WalletFileFlag,
		utils.AccountAddressFlag,
		utils.AccountPassFlag,
		//consensus setting
		utils.EnableConsensusFlag,
		utils.MaxTxInBlockFlag,
		//txpool setting
		utils.GasPriceFlag,
		utils.GasLimitFlag,
		utils.TxpoolPreExecDisableFlag,
		utils.DisableSyncVerifyTxFlag,
		utils.DisableBroadcastNetTxFlag,
		//p2p setting
		utils.ReservedPeersOnlyFlag,
		utils.ReservedPeersFileFlag,
		utils.NetworkIdFlag,
		utils.NodePortFlag,
		utils.ConsensusPortFlag,
		utils.DualPortSupportFlag,
		utils.MaxConnInBoundFlag,
		utils.MaxConnOutBoundFlag,
		utils.MaxConnInBoundForSingleIPFlag,
		//test mode setting
		utils.EnableTestModeFlag,
		utils.TestModeGenBlockTimeFlag,
		//rpc setting
		utils.RPCDisabledFlag,
		utils.RPCPortFlag,
		utils.RPCLocalEnableFlag,
		utils.RPCLocalProtFlag,
		//rest setting
		utils.RestfulEnableFlag,
		utils.RestfulPortFlag,
		utils.RestfulMaxConnsFlag,
		//ws setting
		utils.WsEnabledFlag,
		utils.WsPortFlag,
	}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		cmd.PrintErrorMsg(err.Error())
		os.Exit(1)
	}
}

func startOntology(ctx *cli.Context) {
	initLog(ctx)

	log.Infof("ontology version %s", config.Version)

	cfg, err := initConfig(ctx)
	if err != nil {
		log.Errorf("initConfig error:%s", err)
		return
	}
	acc, err := initAccount(ctx)
	if err != nil {
		log.Errorf("initWallet error:%s", err)
		return
	}
	stateHashHeight := config.GetStateHashCheckHeight(cfg.P2PNode.NetworkId)
	ldg, err := initLedger(ctx, stateHashHeight)
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
	p2pSvr, p2pPid, err := initP2PNode(ctx, txpool)
	if err != nil {
		log.Errorf("initP2PNode error:%s", err)
		return
	}
	_, err = initConsensus(ctx, p2pPid, txpool, acc)
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

func initLog(ctx *cli.Context) {
	//init log module
	logLevel := ctx.GlobalInt(utils.GetFlagName(utils.LogLevelFlag))
	alog.InitLog(log.PATH)
	log.InitLog(logLevel, log.PATH, log.Stdout)
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

func initAccount(ctx *cli.Context) (*account.Account, error) {
	if !config.DefConfig.Consensus.EnableConsensus {
		return nil, nil
	}
	walletFile := ctx.GlobalString(utils.GetFlagName(utils.WalletFileFlag))
	if walletFile == "" {
		return nil, fmt.Errorf("Please config wallet file using --wallet flag")
	}
	if !common.FileExisted(walletFile) {
		return nil, fmt.Errorf("Cannot find wallet file:%s. Please create wallet first", walletFile)
	}

	acc, err := cmdcom.GetAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account error:%s", err)
	}
	log.Infof("Using account:%s", acc.Address.ToBase58())

	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		curPk := hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey))
		config.DefConfig.Genesis.SOLO.Bookkeepers = []string{curPk}
	}

	log.Infof("Account init success")
	return acc, nil
}

func initLedger(ctx *cli.Context, stateHashHeight uint32) (*ledger.Ledger, error) {
	events.Init() //Init event hub

	var err error
	dbDir := utils.GetStoreDirPath(config.DefConfig.Common.DataDir, config.DefConfig.P2PNode.NetworkName)
	ledger.DefLedger, err = ledger.NewLedger(dbDir, stateHashHeight)
	if err != nil {
		return nil, fmt.Errorf("NewLedger error:%s", err)
	}
	bookKeepers, err := config.DefConfig.GetBookkeepers()
	if err != nil {
		return nil, fmt.Errorf("GetBookkeepers error:%s", err)
	}
	genesisConfig := config.DefConfig.Genesis
	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, genesisConfig)
	if err != nil {
		return nil, fmt.Errorf("genesisBlock error %s", err)
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("Init ledger error:%s", err)
	}

	log.Infof("Ledger init success")
	return ledger.DefLedger, nil
}

func initTxPool(ctx *cli.Context) (*proc.TXPoolServer, error) {
	disablePreExec := ctx.GlobalBool(utils.GetFlagName(utils.TxpoolPreExecDisableFlag))
	bactor.DisableSyncVerifyTx = ctx.GlobalBool(utils.GetFlagName(utils.DisableSyncVerifyTxFlag))
	disableBroadcastNetTx := ctx.GlobalBool(utils.GetFlagName(utils.DisableBroadcastNetTxFlag))
	txPoolServer, err := txnpool.StartTxnPoolServer(disablePreExec, disableBroadcastNetTx)
	if err != nil {
		return nil, fmt.Errorf("Init txpool error:%s", err)
	}
	stlValidator, _ := stateless.NewValidator("stateless_validator")
	stlValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))
	stlValidator2, _ := stateless.NewValidator("stateless_validator2")
	stlValidator2.Register(txPoolServer.GetPID(tc.VerifyRspActor))
	stfValidator, _ := stateful.NewValidator("stateful_validator")
	stfValidator.Register(txPoolServer.GetPID(tc.VerifyRspActor))

	hserver.SetTxnPoolPid(txPoolServer.GetPID(tc.TxPoolActor))
	hserver.SetTxPid(txPoolServer.GetPID(tc.TxActor))

	log.Infof("TxPool init success")
	return txPoolServer, nil
}

func initP2PNode(ctx *cli.Context, txpoolSvr *proc.TXPoolServer) (*p2pserver.P2PServer, *actor.PID, error) {
	if config.DefConfig.Genesis.ConsensusType == config.CONSENSUS_TYPE_SOLO {
		return nil, nil, nil
	}
	p2p := p2pserver.NewServer()

	p2pActor := p2pactor.NewP2PActor(p2p)
	p2pPID, err := p2pActor.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("p2pActor init error %s", err)
	}
	p2p.SetPID(p2pPID)
	err = p2p.Start()
	if err != nil {
		return nil, nil, fmt.Errorf("p2p service start error %s", err)
	}
	netreqactor.SetTxnPoolPid(txpoolSvr.GetPID(tc.TxActor))
	txpoolSvr.RegisterActor(tc.NetActor, p2pPID)
	hserver.SetNetServerPID(p2pPID)
	p2p.WaitForPeersStart()
	log.Infof("P2P init success")
	return p2p, p2pPID, nil
}

func initConsensus(ctx *cli.Context, p2pPid *actor.PID, txpoolSvr *proc.TXPoolServer, acc *account.Account) (consensus.ConsensusService, error) {
	if !config.DefConfig.Consensus.EnableConsensus {
		return nil, nil
	}
	pool := txpoolSvr.GetPID(tc.TxPoolActor)

	consensusType := strings.ToLower(config.DefConfig.Genesis.ConsensusType)
	consensusService, err := consensus.NewConsensusService(consensusType, acc, pool, nil, p2pPid)
	if err != nil {
		return nil, fmt.Errorf("NewConsensusService:%s error:%s", consensusType, err)
	}
	consensusService.Start()

	netreqactor.SetConsensusPid(consensusService.GetPID())
	hserver.SetConsensusPid(consensusService.GetPID())

	log.Infof("Consensus init success")
	return consensusService, nil
}

func initRpc(ctx *cli.Context) error {
	if !config.DefConfig.Rpc.EnableHttpJsonRpc {
		return nil
	}
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
	if !ctx.GlobalBool(utils.GetFlagName(utils.RPCLocalEnableFlag)) {
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
	if !config.DefConfig.Restful.EnableHttpRestful {
		return
	}
	go restful.StartServer()

	log.Infof("Restful init success")
}

func initWs(ctx *cli.Context) {
	if !config.DefConfig.Ws.EnableHttpWs {
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
				log.InitLog(int(config.DefConfig.Common.LogLevel), log.PATH, log.Stdout)
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
