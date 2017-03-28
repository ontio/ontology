package validation

import (
	"DNA/core/ledger"
	tx "DNA/core/transaction"
	. "DNA/errors"
	"errors"
	"fmt"
)

func VerifyBlock(block *ledger.Block, ld *ledger.Ledger, completely bool) error {
	if block.Blockdata.Height == 0 {
		return nil
	}
	err := VerifyBlockData(block.Blockdata, ld)
	if err != nil {
		return err
	}

	flag, err := VerifySignableData(block)
	if flag && err == nil {
		return nil
	} else {
		return err
	}

	if block.Transcations == nil {
		return errors.New(fmt.Sprintf("No Transcations Exist in Block."))
	}
	if block.Transcations[0].TxType != tx.BookKeeping {
		return errors.New(fmt.Sprintf("Blockdata Verify failed first Transacion in block is not BookKeeping type."))
	}
	for index, v := range block.Transcations {
		if v.TxType == tx.BookKeeping && index != 0 {
			return errors.New(fmt.Sprintf("This Block Has BookKeeping transaction after first transaction in block."))
		}
	}

	//verfiy block's transactions
	if completely {
	/*
		mineraddress, err := ledger.GetMinerAddress(ld.Blockchain.GetMinersByTXs(block.Transcations))
		if err != nil {
			return errors.New(fmt.Sprintf("GetMinerAddress Failed."))
		}
		if block.Blockdata.NextMiner != mineraddress {
			return errors.New(fmt.Sprintf("Miner is not validate."))
		}
	*/
		for _, txVerify := range block.Transcations {
			transpool := []*tx.Transaction{}
			for _, tx := range block.Transcations {
				if tx.Hash() != txVerify.Hash() {
					transpool = append(transpool, tx)
				}
			}
			err := VerifyTransaction(txVerify, ld, transpool)
			if err != nil {
				return errors.New(fmt.Sprintf("The Input is exist in serval transaction in one block."))
			}
		}
	}

	return nil
}

func VerifyBlockData(bd *ledger.Blockdata, ledger *ledger.Ledger) error {
	if bd.Height == 0 {
		return nil
	}

	prevHeader, err := ledger.Blockchain.GetHeader(bd.PrevBlockHash)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BlockValidator], Cannnot find prevHeader..")
	}
	if prevHeader == nil {
		return NewDetailErr(err, ErrNoCode, "[BlockValidator], Cannnot find previous block.")
	}

	if prevHeader.Blockdata.Height+1 != bd.Height {
		return NewDetailErr(err, ErrNoCode, "[BlockValidator], block height is incorrect.")
	}

	if prevHeader.Blockdata.Timestamp >= bd.Timestamp {
		return NewDetailErr(err, ErrNoCode, "[BlockValidator], block timestamp is incorrect.")
	}

	return nil
}
