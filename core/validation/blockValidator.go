package validation

import (
	"errors"
	"fmt"
	"github.com/Ontology/common"
	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	. "github.com/Ontology/errors"
)

func VerifyBlock(block *types.Block, ld *ledger.Ledger, completely bool) error {
	if block.Header.Height == 0 {
		return nil
	}

	err := VerifyHeaderProgram(block.Header)
	if err != nil {
		return err
	}

	prevHeader, err := ld.Blockchain.GetHeader(block.Header.PrevBlockHash)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[BlockValidator], Cannnot find prevHeader..")
	}

	err = VerifyHeader(block.Header, prevHeader)
	if err != nil {
		return err
	}

	if block.Transactions == nil {
		return errors.New(fmt.Sprintf("No Transactions Exist in Block."))
	}
	if block.Transactions[0].TxType != types.BookKeeping {
		return errors.New(fmt.Sprintf("Header Verify failed first Transacion in block is not BookKeeping type."))
	}

	for index, v := range block.Transactions {
		if v.TxType == types.BookKeeping && index != 0 {
			return errors.New(fmt.Sprintf("This Block Has BookKeeping transaction after first transaction in block."))
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

			if errCode := VerifyTransactionWithLedger(txVerify, ld); errCode != ErrNoError {
				return errors.New(fmt.Sprintf("VerifyTransaction failed when verifiy block"))
			}
		}
	}

	return nil
}

func VerifyHeader(header, prevHeader *types.Header) error {
	if header.Height == 0 {
		return nil
	}

	if prevHeader == nil {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], Cannnot find previous block.")
	}

	if prevHeader.Height+1 != header.Height {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], block height is incorrect.")
	}

	if prevHeader.Timestamp >= header.Timestamp {
		return NewDetailErr(errors.New("[BlockValidator] error"), ErrNoCode, "[BlockValidator], block timestamp is incorrect.")
	}

	programhash := common.ToCodeHash(header.Program.Code)
	if prevHeader.NextBookKeeper != types.Address(programhash) {
		return errors.New("wrong bookkeeper address")
	}

	return nil
}
