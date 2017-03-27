package ledger

import (
	. "DNA/common"
	tx "DNA/core/transaction"
	"DNA/crypto"
	. "DNA/errors"
	"errors"
	"DNA/core/asset"
	"DNA/core/contract"
	"DNA/common"
)

var DefaultLedger *Ledger
var StandbyMiners []*crypto.PubKey

// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	Blockchain *Blockchain
	State      *State
	Store      ILedgerStore
}

//check weather the transaction contains the doubleSpend.
func (l *Ledger) IsDoubleSpend(Tx *tx.Transaction) error {
	//TODO: implement ledger IsDoubleSpend

	return nil
}

//Get the DefaultLedger.
//Note: the later version will support the mutiLedger.So this func mybe expired later.
func GetDefaultLedger() (*Ledger, error) {
	if DefaultLedger == nil {
		return nil, NewDetailErr(errors.New("[Ledger], GetDefaultLedger failed, DefaultLedger not Exist."), ErrNoCode, "")
	}
	return DefaultLedger, nil
}

//Calc the Miners address by miners pubkey.
func GetMinerAddress(miners []*crypto.PubKey) (Uint160,error) {
	//TODO: GetMinerAddress()
	//return Uint160{}
	//CreateSignatureRedeemScript
	if len(miners) < 1 {
		return Uint160{}, NewDetailErr(errors.New("[Ledger] , GetMinerAddress with no miner"), ErrNoCode, "")
	}
	var temp []byte
	var err error
	if len(miners) > 1 {
		temp, err = contract.CreateMultiSigRedeemScript(len(miners) - (len(miners) - 1) / 3, miners)
		if err != nil {
			return Uint160{}, NewDetailErr(err, ErrNoCode, "[Ledger],GetMinerAddress failed with CreateMultiSigRedeemScript.")
		}
	} else {
		temp, err = contract.CreateSignatureRedeemScript(miners[0])
		if err != nil {
			return Uint160{}, NewDetailErr(err, ErrNoCode, "[Ledger],GetMinerAddress failed with CreateMultiSigRedeemScript.")
		}
	}
	codehash ,err:=common.ToCodeHash(temp)
	if err != nil{
		return Uint160{},NewDetailErr(err, ErrNoCode, "[Ledger],GetMinerAddress failed with ToCodeHash.")
	}
	return codehash,nil
}

//Get the Asset from store.
func (l *Ledger) GetAsset(assetId Uint256) (*asset.Asset,error) {
	asset, err := l.Store.GetAsset(assetId)
	if err != nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetAsset failed with assetId ="+ assetId.ToString())
	}
	return asset,nil
}

//Get Block With Height.
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

//Get block with block hash.
func (l *Ledger) GetBlockWithHash(hash Uint256) (*Block, error) {
	bk, err := l.Store.GetBlock(hash)
	if err != nil {
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with hash="+ hash.ToString())
	}
	return bk, nil
}

//Get transaction with hash.
func (l *Ledger) GetTransactionWithHash(hash Uint256) (*tx.Transaction, error) {
	tx, err := l.Store.GetTransaction(hash)
	if err != nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Ledger],GetTransactionWithHash failed with hash="+ hash.ToString())
	}
	return tx, nil
}

//Get local block chain height.
func (l *Ledger) GetLocalBlockChainHeight() uint32{
	return l.Blockchain.BlockHeight
}
