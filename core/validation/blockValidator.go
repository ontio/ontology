package validation

import (
	"errors"
	"fmt"
	"github.com/Ontology/core/ledger"
	tx "github.com/Ontology/core/transaction"
	"github.com/Ontology/core/transaction/payload"
	"github.com/Ontology/core/transaction/utxo"
	. "github.com/Ontology/errors"
)

func VerifyBlock(block *ledger.Block, ld *ledger.Ledger, completely bool) error {
	if block.Header.Height == 0 {
		return nil
	}
	err := VerifyBlockData(block.Header, ld)
	if err != nil {
		return err
	}

	err = VerifySignableDataSignature(block)
	if err != nil {
		return err
	}

	err = VerifySignableDataProgramHashes(block)
	if err != nil {
		return err
	}

	if block.Transactions == nil {
		return errors.New(fmt.Sprintf("No Transactions Exist in Block."))
	}
	if block.Transactions[0].TxType != tx.BookKeeping {
		return errors.New(fmt.Sprintf("Header Verify failed first Transacion in block is not BookKeeping type."))
	}

	claimTransactions := make([]*tx.Transaction, 0)
	for index, v := range block.Transactions {
		if v.TxType == tx.BookKeeping && index != 0 {
			return errors.New(fmt.Sprintf("This Block Has BookKeeping transaction after first transaction in block."))
		}
		if v.TxType == tx.Claim {
			claimTransactions = append(claimTransactions, v)
		}
	}

	for i := 0; i < len(claimTransactions)-1; i++ {
		if isIntersectClaim(claimTransactions[i].Payload.(*payload.Claim).Claims, claimTransactions[i+1].Payload.(*payload.Claim).Claims) {
			return errors.New("[VerifyBlock], Invalid intersect claim")
		}
	}

	//verfiy block's transactions
	if completely {
		/*
			//TODO: NextBookKeeper Check.
			bookKeeperaddress, err := ledger.GetBookKeeperAddress(ld.Blockchain.GetBookKeepersByTXs(block.Transactions))
			if err != nil {
				return errors.New(fmt.Sprintf("GetBookKeeperAddress Failed."))
			}
			if block.Header.NextBookKeeper != bookKeeperaddress {
				return errors.New(fmt.Sprintf("BookKeeper is not validate."))
			}
		*/
		for _, txVerify := range block.Transactions {
			if errCode := VerifyTransaction(txVerify); errCode != ErrNoError {
				return errors.New(fmt.Sprintf("VerifyTransaction failed when verifiy block"))
			}
			if errCode := VerifyTransactionWithLedger(txVerify, ledger.DefaultLedger); errCode != ErrNoError {
				return errors.New(fmt.Sprintf("VerifyTransactionWithLedger failed when verifiy block"))
			}
		}
		if err := VerifyTransactionWithBlock(block.Transactions); err != nil {
			return errors.New(fmt.Sprintf("VerifyTransactionWithBlock failed when verifiy block"))
		}
	}

	return nil
}

func VerifyHeader(bd *ledger.Header, ledger *ledger.Ledger) error {
	return VerifyBlockData(bd, ledger)
}

func VerifyBlockData(bd *ledger.Header, ledger *ledger.Ledger) error {
	if bd.Height == 0 {
		return nil
	}

	prevHeader, err := ledger.Blockchain.GetHeader(bd.PrevBlockHash)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BlockValidator], Cannnot find prevHeader..")
	}
	if prevHeader == nil {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], Cannnot find previous block.")
	}

	if prevHeader.Height+1 != bd.Height {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], block height is incorrect.")
	}

	if prevHeader.Timestamp >= bd.Timestamp {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], block timestamp is incorrect.")
	}

	return nil
}

func isIntersectClaim(c1, c2 []*utxo.UTXOTxInput) bool {
	for _, v := range c1 {
		for _, n := range c2 {
			if v == n {
				return true
			}
		}
	}
	return false
}
