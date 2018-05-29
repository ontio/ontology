package actor

import (
	ldgactor "github.com/ontio/ontology/core/ledger/actor"
	"testing"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"os"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology-crypto/keypair"
	"sort"
	"github.com/ontio/ontology/core/types"
	netreqactor "github.com/ontio/ontology/p2pserver/actor/req"
	"github.com/stretchr/testify/assert"
	"github.com/ontio/ontology/core/genesis"
	"time"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/payload"
)
var ledgerPID *actor.PID
var err error

var acct *account.Account

var defBookkeepers []keypair.PublicKey

func init(){
	log.Init(log.PATH, log.Stdout)
	client := account.Open("/Users/sss/gopath/src/github.com/ontio/ontology/wallet.dat",[]byte("111111"))
	acct = client.GetDefaultAccount()
	ledgerstore.DBDirEvent = "../../../Chain/ledgerevent"
	ledgerstore.DBDirBlock = "../../../Chain/block"
	ledgerstore.DBDirState = "../../../Chain/states"
	ledgerstore.MerkleTreeStorePath = "../../../Chain/merkle_tree.db"

	defBookkeepers, err = client.GetBookkeepers()
	sort.Sort(keypair.NewPublicList(defBookkeepers))
	log.Info("1. Loading the Ledger")
	ledger.DefLedger, err = ledger.NewLedger()
	if err != nil {
		log.Fatalf("NewLedger error %s", err)
		os.Exit(1)
	}
	err = ledger.DefLedger.Init(defBookkeepers)
	if err != nil {
		log.Fatalf("DefLedger.Init error %s", err)
		os.Exit(1)
	}
	ldgerActor := ldgactor.NewLedgerActor()
	ledgerPID = ldgerActor.Start()
	SetLedgerPid(ledgerPID)
	netreqactor.SetLedgerPid(ledgerPID)
}

func TestGetBlockByHeight(t *testing.T) {
	_, err := GetBlockByHeight(uint32(1))
	assert.Nil(t,err)
}

func TestGetHeaderByHeight(t *testing.T) {
	_, err := GetHeaderByHeight(1);
	assert.Nil(t, err)

}

func TestGetBlockHashFromStore(t *testing.T) {
	_, err := GetBlockHashFromStore(1)
	assert.Nil(t, err)
}

func TestCurrentBlockHash(t *testing.T) {
	_, err := CurrentBlockHash()
	assert.Nil(t, err)
}

func TestGetBlockFromStore(t *testing.T) {
	block, _ := GetHeaderByHeight(uint32(1))
	_, err := GetBlockFromStore(block.Hash())
	assert.Nil(t, err)
}

func TestBlockHeight(t *testing.T) {
	_, err := BlockHeight()
	assert.Nil(t, err)
}

func TestGetTransaction(t *testing.T){
	block, _ := GetBlockByHeight(1)
	_, err := GetTransaction(block.Transactions[0].Hash())
	assert.Nil(t , err)
}

func TestGetStorageItem(t *testing.T){

	_, err := GetStorageItem(genesis.OntContractAddress, acct.Address[:])
	assert.Nil(t, err)
}

func TestGetContractStateFromStore(t *testing.T) {
	_, err := GetContractStateFromStore(genesis.OntContractAddress)
	assert.Nil(t, err)
}

func TestGetTxnWithHeightByTxHash(t *testing.T){
	block, _ := GetBlockByHeight(uint32(1))
	_ , _,err := GetTxnWithHeightByTxHash(block.Transactions[0].Hash())
	assert.Nil(t, err)
}

func TestAddBlock(t *testing.T) {

	feeSum := common.Fixed64(0)

	// TODO: increment checking txs

	nonce := common.GetNonce()
	txBookkeeping := createBookkeepingTransaction(nonce, feeSum)

	transactions := make([]*types.Transaction, 0, 1)
	transactions = append(transactions, txBookkeeping)

	txHash := []common.Uint256{}
	for _, t := range transactions {
		txHash = append(txHash, t.Hash())
	}
	txRoot, _ := common.ComputeMerkleRoot(txHash)


	blockRoot := ledger.DefLedger.GetBlockRootWithNewTxRoot(txRoot)

	prevHash,_ := CurrentBlockHash()
	height, err := BlockHeight()
	addr, _ := types.AddressFromBookkeepers(defBookkeepers)

	header := &types.Header{
		Version:          0,
		PrevBlockHash:    prevHash,
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        uint32(time.Now().Unix()),
		Height:           height + 1,
		ConsensusData:    nonce,
		NextBookkeeper:   addr,
	}
	block := &types.Block{
		Header:       header,
		Transactions: transactions,
	}

	blockHash := block.Hash()

	sig, err := signature.Sign(acct, blockHash[:])


	block.Header.Bookkeepers = []keypair.PublicKey{defBookkeepers[0]}
	block.Header.SigData = [][]byte{sig}

	err = AddBlock(block)
	assert.Nil(t,err)
}

func createBookkeepingTransaction(nonce uint64, fee common.Fixed64) *types.Transaction {
	log.Debug()
	//TODO: sysfee
	bookKeepingPayload := &payload.Bookkeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}
	tx := &types.Transaction{
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
		Attributes: []*types.TxAttribute{},
	}
	txHash := tx.Hash()
	acc := acct
	s, err := signature.Sign(acc, txHash[:])
	if err != nil {
		return nil
	}
	sig := &types.Sig{
		PubKeys: []keypair.PublicKey{acc.PublicKey},
		M:       1,
		SigData: [][]byte{s},
	}
	tx.Sigs = []*types.Sig{sig}
	return tx
}


func TestGetEventNotifyByTxHash(t *testing.T) {
	block, _ := GetBlockByHeight(uint32(1))
	_, err := GetEventNotifyByTxHash(block.Transactions[0].Hash())
	assert.Nil(t, err)
}

func TestGetEventNotifyByHeight(t *testing.T) {
	_, err := GetEventNotifyByHeight(0)
	assert.Nil(t, err)
}

func TestGetMerkleProof(t *testing.T) {
	block, _ := GetBlockByHeight(uint32(1))
	height, _, _ := GetTxnWithHeightByTxHash(block.Transactions[0].Hash())
	curHeight, _ := BlockHeight()
	_, err := GetMerkleProof(uint32(height),uint32(curHeight))
	assert.Nil(t, err)
}
