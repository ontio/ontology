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
package account

import (
	"encoding/hex"
	"os"
	"sort"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
)

func genAccountData() (*AccountData, *keypair.ProtectedKey) {
	var acc = new(AccountData)
	prvkey, pubkey, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	ta := types.AddressFromPubKey(pubkey)
	address := ta.ToBase58()
	password := []byte("123456")
	prvSectet, _ := keypair.EncryptPrivateKey(prvkey, address, password)
	acc.SetKeyPair(prvSectet)
	acc.SigSch = "SHA256withECDSA"
	acc.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
	return acc, prvSectet
}

func TestAccountData(t *testing.T) {
	acc, prvSectet := genAccountData()
	assert.NotNil(t, acc)
	assert.Equal(t, acc.Address, acc.ProtectedKey.Address)
	assert.Equal(t, prvSectet, acc.GetKeyPair())
}

func TestWalletSave(t *testing.T) {
	walletFile := "w.data"
	defer func() {
		os.Remove(walletFile)
		os.RemoveAll("Log/")
	}()

	wallet := NewWalletData()
	size := 10
	for i := 0; i < size; i++ {
		acc, _ := genAccountData()
		wallet.AddAccount(acc)
		err := wallet.Save(walletFile)
		if err != nil {
			t.Errorf("Save error:%s", err)
			return
		}
	}

	wallet2 := NewWalletData()
	err := wallet2.Load(walletFile)
	if err != nil {
		t.Errorf("Load error:%s", err)
		return
	}

	assert.Equal(t, len(wallet2.Accounts), len(wallet.Accounts))
}

func TestWalletDel(t *testing.T) {
	wallet := NewWalletData()
	size := 10
	accList := make([]string, 0, size)
	for i := 0; i < size; i++ {
		acc, _ := genAccountData()
		wallet.AddAccount(acc)
		accList = append(accList, acc.Address)
	}
	sort.Strings(accList)
	for _, address := range accList {
		wallet.DelAccount(address)
		_, index := wallet.GetAccountByAddress(address)
		if !assert.Equal(t, -1, index) {
			return
		}
	}
}
