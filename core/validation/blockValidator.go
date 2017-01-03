package validation

import (
	"GoOnchain/core/ledger"
	"errors"
)

func VerifyBlock(block *ledger.Block,ledger *ledger.Ledger,completely bool) error  {

	err := VerifyBlockData(block.Blockdata,ledger)
	if(err != nil) {
		return err
	}

	//verfiy block's transactions
	if(completely){
		for _, Tx := range block.Transcations{
			err := VerifyTransaction(Tx,ledger,nil) //TODO: confirm tx pool
			if(err != nil){
				return err
			}
		}
	}
	return nil
}


func VerifyBlockData(bd *ledger.Blockdata,ledger *ledger.Ledger) error  {

	//TODO: genesis block check

	if ledger.Blockchain.ContainsBlock(bd.Hash()){
		return  nil
	}

	prevHeader := ledger.Blockchain.GetHeader(bd.PrevBlockHash)
	if(prevHeader == nil){
		return  errors.New("Cannnot find previous block.")
	}

	if(prevHeader.Blockdata.Height+1 != bd.Height){
		return  errors.New("block height is incorrect.")
	}

	if(prevHeader.Blockdata.Timestamp >= bd.Timestamp){
		return  errors.New("block timestamp is incorrect.")
	}

	return VerifySignableData(bd)
}