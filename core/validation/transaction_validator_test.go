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
	"strings"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
)

func signerTestAddress(b byte) common.Address {
	var addr common.Address
	addr[0] = b
	return addr
}

func TestEnsureTransactionVerifiedRejectsUnverified(t *testing.T) {
	tx := &types.Transaction{Payer: signerTestAddress(0x01)}
	if err := EnsureTransactionVerified(tx, false); err == nil {
		t.Fatal("expected an unverified transaction to be rejected")
	}
}

func TestRejectExtraSignersBeforeEnabled(t *testing.T) {
	extra := signerTestAddress(0x02)

	tx := &types.Transaction{ExtraSigners: []common.Address{extra}}
	err := checkTransactionSignatures(tx, false)
	if err == nil || !strings.Contains(err.Error(), "not enabled") {
		t.Fatalf("expected disabled extra signers to be rejected, got %v", err)
	}

	// The same gate must apply when SignedAddr is already populated.
	payer := signerTestAddress(0x01)
	verified := &types.Transaction{
		Payer:        payer,
		SignedAddr:   []common.Address{payer, extra},
		ExtraSigners: []common.Address{extra},
	}
	if err := EnsureTransactionVerified(verified, false); err == nil || !strings.Contains(err.Error(), "not enabled") {
		t.Fatalf("expected verified transaction with disabled extra signers to be rejected, got %v", err)
	}
}

// GetSignatureAddresses is also used by pre-execution paths. Its fallback must
// not populate SignedAddr, otherwise a forged signature could later be treated
// as already verified.
func TestGetSignatureAddressesDoesNotMarkVerified(t *testing.T) {
	acc := account.NewAccount("")
	tx := &types.Transaction{
		Payer:   acc.Address,
		Payload: &payload.InvokeCode{Code: []byte{}},
		Sigs: []types.Sig{{
			PubKeys: []keypair.PublicKey{acc.PublicKey},
			M:       1,
			SigData: [][]byte{[]byte("invalid-signature")},
		}},
	}

	claimed := tx.GetSignatureAddresses()
	if len(claimed) != 1 || claimed[0] != acc.Address {
		t.Fatalf("unexpected claimed signers: %v", claimed)
	}
	if verified := tx.GetVerifiedSignerAddresses(); len(verified) != 0 {
		t.Fatalf("GetSignatureAddresses must not mark the transaction verified, got %v", verified)
	}
	if err := EnsureTransactionVerified(tx, false); err == nil {
		t.Fatal("expected the invalid signature to be rejected")
	}
}
