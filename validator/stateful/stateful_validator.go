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

package stateful

import (
	ethcomm "github.com/ethereum/go-ethereum/common"
	"github.com/gammazero/workerpool"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	vatypes "github.com/ontio/ontology/validator/types"
)

type ValidatorPool struct {
	pool *workerpool.WorkerPool
}

func NewValidatorPool(maxWorkers int) *ValidatorPool {
	return &ValidatorPool{pool: workerpool.New(maxWorkers)}
}

func (self *ValidatorPool) SubmitVerifyTask(tx *types.Transaction, rspCh chan<- *vatypes.CheckResponse) {
	task := func() {
		height := ledger.DefLedger.GetCurrentBlockHeight()

		errCode := errors.ErrNoError
		response := &vatypes.CheckResponse{
			Type:    vatypes.Stateful,
			Hash:    tx.Hash(),
			Tx:      tx,
			Height:  height,
			ErrCode: errCode,
		}
		hash := tx.Hash()

		exist, err := ledger.DefLedger.IsContainTransaction(hash)
		if err != nil {
			response.ErrCode = errors.ErrUnknown
		} else if exist {
			response.ErrCode = errors.ErrDuplicatedTx
		} else if tx.IsEipTx() {
			ethacct, err := ledger.DefLedger.GetEthAccount(ethcomm.Address(tx.Payer))
			if err != nil {
				response.ErrCode = errors.ErrNoAccount
			} else if uint64(tx.Nonce) < ethacct.Nonce {
				response.ErrCode = errors.ErrHigherNonceExist
			} else {
				response.Nonce = ethacct.Nonce
			}
		}
		if IsSenderLimited(tx.GetSignatureAddresses()) {
			response.ErrCode = errors.ErrNoAccount
		}

		rspCh <- response
	}

	self.pool.Submit(task)
}

var senderLimitor = func() map[common.Address]bool {
	limitAddress := []string{
		"AYCYB3rQCVuGUauHFUPU7kawgNYLUZMskP",
		"AM2gvtpUFruGKkV7kFa8FJZJpfzJy3vZdv",
		"AYn9spXyNG8hy2hSNJktLR5LesQY97vXN7",
		"ATZRhpQymY2CZXonv7h3KqptQWAbc9PhXe",
		"AUi6qQQe1R2ka5mG2RdY1EMWty3BnJWtL7",
	}

	limitMap := make(map[common.Address]bool)
	for _, v := range limitAddress {
		addr, err := common.AddressFromBase58(v)
		if err != nil {
			panic(err)
		}
		limitMap[addr] = true
	}
	return limitMap
}()

func IsSenderLimited(senders []common.Address) bool {
	for _, v := range senders {
		if senderLimitor[v] {
			return true
		}
	}
	return false
}
