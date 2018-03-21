package solo

import (
	"time"

	"fmt"
	"github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	actorTypes "github.com/Ontology/consensus/actor"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
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
		err := this.genBlock()
		if err != nil {
			log.Errorf("Solo genBlock error %s", err)
		}
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

func (this *SoloService) genBlock()error {
	block, err := this.makeBlock()
	if err != nil {
		return fmt.Errorf("makeBlock error %s", err)
	}
	err = ledger.DefLedger.AddBlock(block)
	if err != nil {
		return fmt.Errorf("Blockchain.AddBlock error:%s", err)
	}
	//err = this.localNet.CleanTransactions(block.Transactions)
	//if err != nil {
	//	log.Errorf("CleanSubmittedTransactions error:%s", err)
	//	return
	//}
	return nil
}

func (this *SoloService) makeBlock() (*types.Block, error) {
	log.Debug()
	owner := this.Account.PublicKey
	nextBookKeeper, err := types.AddressFromBookKeepers([]*crypto.PubKey{owner})
	if err != nil {
		return nil, fmt.Errorf("GetBookKeeperAddress error:%s", err)
	}
	prevHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight() + 1

	nonce := GetNonce()
	txs := this.poolActor.GetTxnPool(true, height-1)
	// todo : fix feesum calcuation
	feeSum := Fixed64(0)

	// TODO: increment checking txs

	txBookkeeping, err := this.createBookkeepingTransaction(nonce, feeSum)
	if err != nil {
		return nil, fmt.Errorf("createBookkeepingTransaction error:%s", err)
	}

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
		return nil, fmt.Errorf("ComputeRoot error:%s", err)
	}

	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(txRoot)
	header := &types.Header{
		Version:          ContextVersion,
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           height,
		ConsensusData:    nonce,
		NextBookKeeper:   nextBookKeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: transactions,
	}

	blockHash := block.Hash()

	signature, err := crypto.Sign(this.Account.PrivKey(), blockHash[:])
	if err != nil {
		return nil, fmt.Errorf("[Signature],Sign error:%s.", err)
	}

	block.Header.BookKeepers = []*crypto.PubKey{owner}
	block.Header.SigData = [][]byte{signature}
	return block, nil
}

func (this *SoloService) createBookkeepingTransaction(nonce uint64, fee Fixed64) (*types.Transaction, error) {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	tx := &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
	txHash := tx.Hash()
	acc := this.Account
	signature, err := crypto.Sign(acc.PrivateKey, txHash[:])
	if err != nil {
		return nil, fmt.Errorf("crypto.Sign error %s", err)
	}
	sig := &types.Sig{
		PubKeys: []*crypto.PubKey{acc.PublicKey},
		M:       1,
		SigData: [][]byte{signature},
	}
	tx.Sigs = []*types.Sig{sig}
	return tx, nil
}
