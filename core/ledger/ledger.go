package ledger

import (
	. "GoOnchain/common"
	tx "GoOnchain/core/transaction"
	"GoOnchain/crypto"
	. "GoOnchain/errors"
	"errors"
	"GoOnchain/core/asset"
)

var DefaultLedger *Ledger

// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	Blockchain *Blockchain
	State      *State
	Store      ILedgerStore
}

func (l *Ledger) IsDoubleSpend(Tx *tx.Transaction) error {
	//TODO: implement ledger IsDoubleSpend

	return nil
}

func GetDefaultLedger() (*Ledger, error) {
	if DefaultLedger == nil {
		return nil, NewDetailErr(errors.New("[Ledger], GetDefaultLedger failed, DefaultLedger not Exist."), ErrNoCode, "")
	}
	return DefaultLedger, nil
}

func GetMinerAddress(miners []*crypto.PubKey) Uint160 {
	//TODO: GetMinerAddress()
	return Uint160{}
}

func (l *Ledger) GetAsset(assetId Uint256) (*asset.Asset,error) {
	asset, err := l.Store.GetAsset(assetId)
	if err != nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetAsset failed with assetId ="+ assetId.ToString())
	}
	return asset,nil
}

func (l *Ledger) GetBlockWithHeight(height uint32) (*Block, error) {
	temp, err := l.Store.GetBlockHash(height)
	if err != nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with height="+ string(height))
	}
	bk, err := DefaultLedger.Store.GetBlock(temp)
	if err != nil {
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with hash="+ temp.ToString())
	}
	return bk, nil
}

func (l *Ledger) GetBlockWithHash(hash Uint256) (*Block, error) {
	bk, err := l.Store.GetBlock(hash)
	if err != nil {
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with hash="+ hash.ToString())
	}
	return bk, nil
}

func (l *Ledger) GetTransactionWithHash(hash Uint256) (*tx.Transaction, error) {
	tx, err := l.Store.GetTransaction(hash)
	if err != nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetTransactionWithHash failed with hash="+ hash.ToString())
	}
	return tx, nil
}

func (l *Ledger) GetLocalBlockChainHeight() ( uint32, error) {
	height,err := l.Store.GetLocalBlockChainHeight()
	if err !=nil{
		return 0,NewDetailErr(err, ErrNoCode, "[Ledger],GetLocalBlockChainHeight failed")
	}
	return  height,nil
}