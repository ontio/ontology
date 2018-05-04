/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package validation

import (
	"errors"
	"fmt"

	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
)

// VerifyBlock checks whether the block is valid
func VerifyBlock(block *types.Block, ld *ledger.Ledger, completely bool) error {
	header := block.Header
	if header.Height == 0 {
		return nil
	}

	m := len(header.Bookkeepers) - (len(header.Bookkeepers)-1)/3
	hash := block.Hash()
	err := signature.VerifyMultiSignature(hash[:], header.Bookkeepers, m, header.SigData)
	if err != nil {
		return err
	}

	prevHeader, err := ld.GetHeaderByHash(block.Header.PrevBlockHash)
	if err != nil {
		return fmt.Errorf("[BlockValidator], can not find prevHeader: %s", err)
	}

	err = VerifyHeader(block.Header, prevHeader)
	if err != nil {
		return err
	}

	//verfiy block's transactions
	if completely {
		/*
			//TODO: NextBookkeeper Check.
			bookkeeperaddress, err := ledger.GetBookkeeperAddress(ld.Blockchain.GetBookkeepersByTXs(block.Transactions))
			if err != nil {
				return errors.New(fmt.Sprintf("GetBookkeeperAddress Failed."))
			}
			if block.Header.NextBookkeeper != bookkeeperaddress {
				return errors.New(fmt.Sprintf("Bookkeeper is not validate."))
			}
		*/
		for _, txVerify := range block.Transactions {
			if errCode := VerifyTransaction(txVerify); errCode != ontErrors.ErrNoError {
				return errors.New(fmt.Sprintf("VerifyTransaction failed when verifiy block"))
			}

			if errCode := VerifyTransactionWithLedger(txVerify, ld); errCode != ontErrors.ErrNoError {
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
		return errors.New("[BlockValidator], can not find previous block.")
	}

	if prevHeader.Height+1 != header.Height {
		return errors.New("[BlockValidator], block height is incorrect.")
	}

	if prevHeader.Timestamp >= header.Timestamp {
		return errors.New("[BlockValidator], block timestamp is incorrect.")
	}

	address, err := types.AddressFromBookkeepers(header.Bookkeepers)
	if err != nil {
		return err
	}

	if prevHeader.NextBookkeeper != address {
		return fmt.Errorf("bookkeeper address error")
	}

	return nil
}
