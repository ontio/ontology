package ledger

import (
	"errors"
	. "github.com/Ontology/common"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

var DefaultLedger *Ledger
var StandbyBookKeepers []*crypto.PubKey

// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	Blockchain *Blockchain
	Store      ILedgerStore
}

//check weather the transaction contains the doubleSpend.
func (l *Ledger) IsDoubleSpend(Tx *types.Transaction) bool {
	return DefaultLedger.Store.IsDoubleSpend(Tx)
}

//Get the DefaultLedger.
//Note: the later version will support the mutiLedger.So this func mybe expired later.
func GetDefaultLedger() (*Ledger, error) {
	if DefaultLedger == nil {
		return nil, NewDetailErr(errors.New("[Ledger], GetDefaultLedger failed, DefaultLedger not Exist."), ErrNoCode, "")
	}
	return DefaultLedger, nil
}

//Get the Asset from store.
func (l *Ledger) GetAsset(assetId Uint256) (*states.AssetState, error) {
	asset, err := l.Store.GetAsset(assetId)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ledger],GetAsset failed with assetId ="+assetId.ToString())
	}
	return asset, nil
}

//Get Block With Height.
func (l *Ledger) GetBlockWithHeight(height uint32) (*types.Block, error) {
	temp, err := l.Store.GetBlockHash(height)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with height="+string(height))
	}
	bk, err := DefaultLedger.Store.GetBlock(temp)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with hash="+temp.ToString())
	}
	return bk, nil
}

//Get block with block hash.
func (l *Ledger) GetBlockWithHash(hash Uint256) (*types.Block, error) {
	bk, err := l.Store.GetBlock(hash)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with hash="+hash.ToString())
	}
	return bk, nil
}

//BlockInLedger checks if the block existed in ledger
func (l *Ledger) BlockInLedger(hash Uint256) bool {
	return l.Store.IsBlockInStore(hash)
}

//Get transaction with hash.
func (l *Ledger) GetTransactionWithHash(hash Uint256) (*types.Transaction, error) {
	tx, err := l.Store.GetTransaction(hash)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ledger],GetTransactionWithHash failed with hash="+hash.ToString())
	}
	return tx, nil
}

//Get local block chain height.
func (l *Ledger) GetLocalBlockChainHeight() uint32 {
	return l.Blockchain.BlockHeight
}
