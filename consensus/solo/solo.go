package solo

import (
	"time"

	"github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	actorTypes "github.com/Ontology/consensus/actor"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/utils"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/actor"
	"reflect"
)

/*
*Simple consensus for solo node in test environment.
 */
var GenBlockTime = (config.DEFAULTGENBLOCKTIME * time.Second)

const ContextVersion uint32 = 0

type SoloService struct {
	Account     *account.Account
	poolActor   *actorTypes.TxPoolActor
	ledgerActor *actorTypes.LedgerActor
	existCh     chan interface{}

	pid *actor.PID
}

func NewSoloService(bkAccount *account.Account, txpool *actor.PID, ledger *actor.PID) (*SoloService, error) {
	service := &SoloService{
		Account:     bkAccount,
		poolActor:   &actorTypes.TxPoolActor{Pool: txpool},
		ledgerActor: &actorTypes.LedgerActor{Ledger: ledger},
	}

	props := actor.FromProducer(func() actor.Actor {
		return service
	})

	pid, err := actor.SpawnNamed(props, "consensus_solo")
	service.pid = pid
	return service, err
}

func (this *SoloService) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Restarting:
		log.Warn("solo actor restarting")
	case *actor.Stopping:
		log.Warn("solo actor stopping")
	case *actor.Stopped:
		log.Warn("solo actor stopped")
	case *actor.Started:
		log.Warn("solo actor started")
	case *actor.Restart:
		log.Warn("solo actor restart")
	case *actorTypes.StartConsensus:
		if this.existCh != nil {
			log.Warn("consensus have started")
			return
		}
		timer := time.NewTicker(GenBlockTime)
		this.existCh = make(chan interface{})
		go func() {
			defer timer.Stop()
			existCh := this.existCh
			for {
				select {
				case <-timer.C:
					this.pid.Tell(&actorTypes.TimeOut{})
				case <-existCh:
					return
				}
			}
		}()
	case *actorTypes.StopConsensus:
		if this.existCh != nil {
			close(this.existCh)
			this.existCh = nil
		}
	case *actorTypes.TimeOut:
		this.genBlock()
	default:
		log.Info("solo actor: Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (this *SoloService) GetPID() *actor.PID {
	return this.pid
}

func (this *SoloService) Start() error {
	this.pid.Tell(&actorTypes.StartConsensus{})
	return nil
}

func (this *SoloService) Halt() error {
	this.pid.Tell(&actorTypes.StopConsensus{})
	return nil
}

func (this *SoloService) genBlock() {
	block := this.makeBlock()
	if block == nil {
		return
	}
	err := ledger.DefLedger.AddBlock(block)
	if err != nil {
		log.Errorf("Blockchain.AddBlock error:%s", err)
		return
	}
	//err = this.localNet.CleanTransactions(block.Transactions)
	//if err != nil {
	//	log.Errorf("CleanSubmittedTransactions error:%s", err)
	//	return
	//}
}

func (this *SoloService) makeBlock() *types.Block {
	log.Debug()
	owner := this.Account.PublicKey
	nextBookKeeper, err := utils.AddressFromBookKeepers([]*crypto.PubKey{owner})
	if err != nil {
		log.Error("SoloService GetBookKeeperAddress error:%s", err)
		return nil
	}
	prevHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight() + 1

	nonce := GetNonce()
	txs := this.poolActor.GetTxnPool(true, height-1)
	// todo : fix feesum calcuation
	feeSum := Fixed64(0)

	// TODO: increment checking txs

	txBookkeeping := this.createBookkeepingTransaction(nonce, feeSum)

	transactions := make([]*types.Transaction, 0, len(txs)+1)
	transactions = append(transactions, txBookkeeping)
	for _, txEntry := range txs {
		transactions = append(transactions, txEntry.Tx)
	}

	txHash := []Uint256{}
	for _, t := range transactions {
		txHash = append(txHash, t.Hash())
	}
	txRoot, err := crypto.ComputeRoot(txHash)
	if err != nil {
		log.Errorf("ComputeRoot error:%s", err)
		return nil
	}

	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(txRoot)
	stateRoot, err := ledger.DefLedger.GetCurrentStateRoot()
	if err != nil {
		log.Errorf("GetCurrentStateRoot error %s", err)
		return nil
	}
	header := &types.Header{
		Version:          ContextVersion,
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		StateRoot:        stateRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           height,
		ConsensusData:    nonce,
		NextBookKeeper:   nextBookKeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: transactions,
	}

	signature, err := crypto.Sign(this.Account.PrivKey(), block.GetMessage())
	if err != nil {
		log.Error("[Signature],Sign failed.")
		return nil
	}

	block.Header.BookKeepers = []*crypto.PubKey{owner}
	block.Header.SigData = [][]byte{signature}
	return block
}

func (this *SoloService) createBookkeepingTransaction(nonce uint64, fee Fixed64) *types.Transaction {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}

	return &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
}
