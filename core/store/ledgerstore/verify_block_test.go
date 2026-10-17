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

package ledgerstore

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
)

func TestVerifyBlockTransactionsRejectsBadSignature(t *testing.T) {
	acc := account.NewAccount("")
	tx := &types.Transaction{
		Payer:   acc.Address,
		Payload: &payload.InvokeCode{Code: []byte{}},
		Sigs: []types.Sig{{
			PubKeys: []keypair.PublicKey{acc.PublicKey},
			M:       1,
			SigData: [][]byte{[]byte("not-a-signature")},
		}},
	}
	block := &types.Block{
		Header:       &types.Header{Height: 1},
		Transactions: []*types.Transaction{tx},
	}

	if err := (&LedgerStoreImp{}).verifyBlockTransactions(block); err == nil {
		t.Fatal("expected forged signature to be rejected")
	}
}
