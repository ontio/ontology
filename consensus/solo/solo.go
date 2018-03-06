package solo

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	cl "github.com/Ontology/account"
	. "github.com/Ontology/common"
	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/core/contract/program"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/transaction/utxo"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/net"
)

/*
*Simple consensus for solo node in test environment.
 */
var GenBlockTime = (config.DEFAULTGENBLOCKTIME * time.Second)

const ContextVersion uint32 = 0

type SoloService struct {
	Client   cl.Client
	localNet net.Neter
	existCh  chan interface{}
}

func NewSoloService(client cl.Client, localNet net.Neter) *SoloService {
	return &SoloService{
		Client:   client,
		localNet: localNet,
	}
}

func (this *SoloService) Start() error {
	timer := time.NewTicker(GenBlockTime)
	go func() {
		defer timer.Stop()
		for {
			select {
			case <-timer.C:
				this.genBlock()
			case <-this.existCh:
				return
			}
		}
	}()
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
	ac, _ := this.Client.GetDefaultAccount()
	owner := ac.PublicKey
	nextBookKeeper, err := core.AddressFromBookKeepers([]*crypto.PubKey{owner})
	if err != nil {
		log.Error("SoloService GetBookKeeperAddress error:%s", err)
		return nil
	}
	nonce := GetNonce()
	//TODO Need remove after merge
	transactionsPool := make([]*types.Transaction, 0)
	feeSum := Fixed64(0)
	//transactionsPool, feeSum := this.localNet.GetTxnPool(true)
	txBookkeeping := this.createBookkeepingTransaction(nonce, feeSum)

	transactions := make([]*types.Transaction, 0, len(transactionsPool)+1)
	transactions = append(transactions, txBookkeeping)
	for _, transaction := range transactionsPool {
		transactions = append(transactions, transaction)
	}

	prevHash := ledger.DefLedger.GetCurrentBlockHash()
	height := ledger.DefLedger.GetCurrentBlockHeight()+1

	txHash := []Uint256{}
	for _, t := range transactions {
		txHash = append(txHash, t.Hash())
	}
	txRoot, err := crypto.ComputeRoot(txHash)
	if err != nil {
		log.Errorf("ComputeRoot error:%s", err)
		return nil
	}

	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(&txRoot)
	stateRoot, err := ledger.DefLedger.GetCurrentStateRoot()
	if err != nil {
		log.Errorf("GetCurrentStateRoot error %s", err)
		return nil
	}
	header := &types.Header{
		Version:          ContextVersion,
		PrevBlockHash:    *prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        *blockRoot,
		StateRoot:        *stateRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           height,
		ConsensusData:    nonce,
		NextBookKeeper:   nextBookKeeper,
	}
	block := &types.Block{
		Header:       header,
		Transactions: transactions,
	}

	prog, err := this.getBlockProgram(block, owner)
	if err != nil {
		log.Errorf("getBlockPrograms error:%s", err)
		return nil
	}
	if prog == nil {
		log.Errorf("getBlockPrograms programs is nil")
		return nil
	}

	block.Header.Program = prog
	return block
}

func (this *SoloService) getBlockProgram(block *types.Block, owner *crypto.PubKey) (*program.Program, error) {
	buf := new(bytes.Buffer)
	block.SerializeUnsigned(buf)
	account, _ := this.Client.GetAccount(owner)

	signature, err := crypto.Sign(account.PrivKey(), buf.Bytes())
	if err != nil {
		return nil, errors.New("[Signature],Sign failed.")
	}

	sc, err := contract.CreateSignatureContract(owner)
	if err != nil {
		return nil, fmt.Errorf("CreateSignatureContract error:%s", err)
	}

	sb := program.NewProgramBuilder()
	sb.PushData(signature)

	return &program.Program{
		Code:      sc.Code,
		Parameter: sb.ToArray(),
	}, nil

}

func (this *SoloService) createBookkeepingTransaction(nonce uint64, fee Fixed64) *types.Transaction {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	ac, _ := this.Client.GetDefaultAccount()
	owner := ac.PublicKey
	signatureRedeemScript, err := contract.CreateSignatureRedeemScript(owner)
	if err != nil {
		return nil
	}
	signatureRedeemScriptHashToCodeHash := ToCodeHash(signatureRedeemScript)
	if err != nil {
		return nil
	}
	outputs := []*utxo.TxOutput{}
	if fee > 0 {
		feeOutput := &utxo.TxOutput{
			AssetID:     genesis.ONGTokenID,
			Value:       fee,
			ProgramHash: signatureRedeemScriptHashToCodeHash,
		}
		outputs = append(outputs, feeOutput)
	}
	return &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
}

func (this *SoloService) Halt() error {
	close(this.existCh)
	return nil
}
