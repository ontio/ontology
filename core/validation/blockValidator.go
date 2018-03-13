package validation

import (
	"errors"
	"fmt"

	"github.com/Ontology/core/ledger"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/utils"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

func VerifyBlock(block *types.Block, ld *ledger.Ledger, completely bool) error {
	header := block.Header
	if header.Height == 0 {
		return nil
	}

	m := len(header.BookKeepers) - (len(header.BookKeepers)-1)/3
	err := crypto.VerifyMultiSignature(block.Header.GetMessage(), header.BookKeepers, m, header.SigData)
	if err != nil {
		return err
	}

	prevHeader, err := ld.GetHeaderByHash(&block.Header.PrevBlockHash)
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

	address, err := utils.AddressFromBookKeepers(header.BookKeepers)
	if err != nil {
		return err
	}

	if prevHeader.NextBookKeeper != address {
		return fmt.Errorf("bookkeeper address error")
	}

	return nil
}
