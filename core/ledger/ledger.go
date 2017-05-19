package ledger

import (
	"DNA/common"
	. "DNA/common"
	"DNA/core/asset"
	"DNA/core/contract"
	tx "DNA/core/transaction"
	"DNA/crypto"
	. "DNA/errors"
	"errors"
)

var DefaultLedger *Ledger
var StandbyBookKeepers []*crypto.PubKey

// Ledger - the struct for onchainDNA ledger
type Ledger struct {
	Blockchain *Blockchain
	State      *State
	Store      ILedgerStore
}

//check weather the transaction contains the doubleSpend.
func (l *Ledger) IsDoubleSpend(Tx *tx.Transaction) bool {
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

//Calc the BookKeepers address by bookKeepers pubkey.
func GetBookKeeperAddress(bookKeepers []*crypto.PubKey) (Uint160, error) {
	//TODO: GetBookKeeperAddress()
	//return Uint160{}
	//CreateSignatureRedeemScript
	if len(bookKeepers) < 1 {
		return Uint160{}, NewDetailErr(errors.New("[Ledger] , GetBookKeeperAddress with no bookKeeper"), ErrNoCode, "")
	}
	var temp []byte
	var err error
	if len(bookKeepers) > 1 {
		temp, err = contract.CreateMultiSigRedeemScript(len(bookKeepers)-(len(bookKeepers)-1)/3, bookKeepers)
		if err != nil {
			return Uint160{}, NewDetailErr(err, ErrNoCode, "[Ledger],GetBookKeeperAddress failed with CreateMultiSigRedeemScript.")
		}
	} else {
		temp, err = contract.CreateSignatureRedeemScript(bookKeepers[0])
		if err != nil {
			return Uint160{}, NewDetailErr(err, ErrNoCode, "[Ledger],GetBookKeeperAddress failed with CreateMultiSigRedeemScript.")
		}
	}
	codehash, err := common.ToCodeHash(temp)
	if err != nil {
		return Uint160{}, NewDetailErr(err, ErrNoCode, "[Ledger],GetBookKeeperAddress failed with ToCodeHash.")
	}
	return codehash, nil
}

//Get the Asset from store.
func (l *Ledger) GetAsset(assetId Uint256) (*asset.Asset, error) {
	asset, err := l.Store.GetAsset(assetId)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ledger],GetAsset failed with assetId ="+assetId.ToString())
	}
	return asset, nil
}

//Get Block With Height.
func (l *Ledger) GetBlockWithHeight(height uint32) (*Block, error) {
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
func (l *Ledger) GetBlockWithHash(hash Uint256) (*Block, error) {
	bk, err := l.Store.GetBlock(hash)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Ledger],GetBlockWithHeight failed with hash="+hash.ToString())
	}
	return bk, nil
}

//BlockInLedger checks if the block existed in ledger
func (l *Ledger) BlockInLedger(hash Uint256) bool {
	bk, _ := l.Store.GetBlock(hash)
	if bk == nil {
		return false
	}
	return true
}

//Get transaction with hash.
func (l *Ledger) GetTransactionWithHash(hash Uint256) (*tx.Transaction, error) {
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
