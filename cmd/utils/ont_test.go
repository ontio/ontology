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

package utils

import (
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/validation"
	ontErrors "github.com/ontio/ontology/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetSignerAddresses(t *testing.T) {
	payer := account.NewAccount("")
	signer1 := account.NewAccount("")
	signer2 := account.NewAccount("")
	tx := NewInvokeTransaction(0, 0, nil)

	err := SetSignerAddresses(tx, payer.Address, signer1.Address, payer.Address, signer2.Address)
	require.NoError(t, err)
	assert.Equal(t, payer.Address, tx.Payer)
	assert.Equal(t, []common.Address{signer1.Address, signer2.Address}, tx.ExtraSigners)

	err = SignTransaction(payer, tx)
	require.NoError(t, err)
	assert.ErrorContains(t, SetSignerAddresses(tx, signer1.Address), "before signing")
}

func TestSignTransactionDeclaresSignerBeforeSigning(t *testing.T) {
	payer := account.NewAccount("")
	extraSigner := account.NewAccount("")
	tx := NewInvokeTransaction(0, 0, nil)
	tx.Payer = payer.Address

	err := SignTransaction(extraSigner, tx)
	require.NoError(t, err)
	assert.Equal(t, []common.Address{extraSigner.Address}, tx.ExtraSigners)
	assert.Len(t, tx.Sigs, 1)
}

func TestSignTransactionRejectsUndeclaredSignerAfterSigning(t *testing.T) {
	payer := account.NewAccount("")
	extraSigner := account.NewAccount("")
	tx := NewInvokeTransaction(0, 0, nil)
	tx.Payer = payer.Address

	require.NoError(t, SignTransaction(payer, tx))
	err := SignTransaction(extraSigner, tx)
	assert.ErrorContains(t, err, "not declared in ExtraSigners")
}

func TestMultiSignerTransactionVerifiesAfterSigning(t *testing.T) {
	payer := account.NewAccount("")
	signer1 := account.NewAccount("")
	signer2 := account.NewAccount("")
	tx := NewInvokeTransaction(0, 0, nil)
	require.NoError(t, SetSignerAddresses(tx, payer.Address, signer1.Address, signer2.Address))
	require.NoError(t, SignTransaction(payer, tx))
	require.NoError(t, SignTransaction(signer1, tx))
	require.NoError(t, SignTransaction(signer2, tx))

	immutable, err := tx.IntoImmutable()
	require.NoError(t, err)
	assert.Equal(t, ontErrors.ErrNoError, validation.VerifyTransaction(immutable, true))
}

func TestMultiSigTransactionDeclaresMultiAddress(t *testing.T) {
	payer := account.NewAccount("")
	signer1 := account.NewAccount("")
	signer2 := account.NewAccount("")

	tx := NewInvokeTransaction(0, 0, nil)
	tx.Payer = payer.Address
	keys := []keypair.PublicKey{signer1.PublicKey, signer2.PublicKey}
	multiAddr, err := types.AddressFromMultiPubKeys(keys, 2)
	require.NoError(t, err)

	require.NoError(t, MultiSigTransaction(tx, 2, keys, signer1))
	assert.Equal(t, []common.Address{multiAddr}, tx.ExtraSigners)
	require.NoError(t, SignTransaction(payer, tx))
	require.NoError(t, MultiSigTransaction(tx, 2, keys, signer2))

	immutable, err := tx.IntoImmutable()
	require.NoError(t, err)
	assert.Equal(t, ontErrors.ErrNoError, validation.VerifyTransaction(immutable, true))
}
