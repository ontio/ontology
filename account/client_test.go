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
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	testWallet     Client
	testWalletPath = "./wallet_test.dat"
	testPasswd     = []byte("123456")
)

func TestMain(t *testing.M) {
	var err error
	testWallet, err = Open(testWalletPath)
	if err != nil {
		fmt.Printf("Open wallet:%s error:%s\n", testWalletPath, err)
		return
	}
	t.Run()
	os.Remove(testWalletPath)
	os.Remove("ActorLog")
}

func TestClientNewAccount(t *testing.T) {
	accountNum := testWallet.GetAccountNum()
	label1 := "t1"
	acc1, err := testWallet.NewAccount(label1, keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount error:%s", err)
		return
	}
	label2 := "t2"
	acc2, err := testWallet.NewAccount(label2, keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount error:%s", err)
		return
	}
	if accountNum+2 != testWallet.GetAccountNum() {
		t.Errorf("TestClientNewAccount account num:%d != %d", testWallet.GetAccountNum(), accountNum+2)
		return
	}
	accTmp, err := testWallet.GetAccountByAddress(acc1.Address.ToBase58(), testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount GetAccountByAddress:%s error:%s", acc1.Address.ToBase58(), err)
		return
	}
	if accTmp.Address.ToBase58() != acc1.Address.ToBase58() {
		t.Errorf("TestClientNewAccount by address address:%s != %s", accTmp.Address.ToBase58(), acc1.Address.ToBase58())
		return
	}
	accTmp, err = testWallet.GetAccountByIndex(accountNum+1, testPasswd)
	if err != nil {
		t.Errorf("")
	}
	if accTmp.Address.ToBase58() != acc1.Address.ToBase58() {
		t.Errorf("TestClientNewAccount by index address:%s != %s", accTmp.Address.ToBase58(), acc1.Address.ToBase58())
		return
	}
	accTmp, err = testWallet.GetAccountByLabel(label1, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount GetAccountByLabel:%s error:%s", label1, err)
		return
	}
	if accTmp.Address.ToBase58() != acc1.Address.ToBase58() {
		t.Errorf("TestClientNewAccount by label address:%s != %s", accTmp.Address.ToBase58(), acc1.Address.ToBase58())
		return
	}

	testWallet2, err := Open(testWalletPath)
	if err != nil {
		t.Errorf("NewAccount Open wallet:%s error:%s", testWalletPath, err)
		return
	}

	if testWallet.GetAccountNum() != testWallet2.GetAccountNum() {
		t.Errorf("TestClientNewAccount  AccountNum:%d != %d", testWallet2.GetAccountNum(), testWallet.GetAccountNum())
		return
	}

	accTmp, err = testWallet2.GetAccountByLabel(label2, testPasswd)
	if err != nil {
		t.Errorf("TestClientNewAccount GetAccountByLabel:%s error:%s", label2, err)
		return
	}

	if accTmp.Address.ToBase58() != acc2.Address.ToBase58() {
		t.Errorf("TestClientNewAccount reopen address:%s != %s", accTmp.Address.ToBase58(), acc2.Address.ToBase58())
		return
	}

	_, err = testWallet.NewAccount(label2, keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err == nil {
		t.Errorf("TestClientNewAccount new account with duplicate label:%s should failed", label2)
		return
	}
}

func TestClientDeleteAccount(t *testing.T) {
	accountNum := testWallet.GetAccountNum()
	accSize := 10
	for i := 0; i < accSize; i++ {
		_, err := testWallet.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
		if err != nil {
			t.Errorf("TestClientDeleteAccount NewAccount error:%s", err)
			return
		}
	}
	delIndex := accountNum + 3
	delAcc, err := testWallet.GetAccountByIndex(delIndex, testPasswd)
	if err != nil {
		t.Errorf("TestClientDeleteAccount GetAccountByIndex:%d error:%s", delIndex, err)
		return
	}
	if delAcc == nil {
		t.Errorf("TestClientDeleteAccount cannot getaccount by index:%d", delIndex)
		return
	}

	accountNum += accSize
	delAccTmp, err := testWallet.DeleteAccount(delAcc.Address.ToBase58(), testPasswd)
	if err != nil {
		t.Errorf("TestClientDeleteAccount DeleteAccount error:%s", err)
		return
	}
	if delAcc.Address.ToBase58() != delAccTmp.Address.ToBase58() {
		t.Errorf("TestClientDeleteAccount Account address %s != %s", delAcc.Address.ToBase58(), delAccTmp.Address.ToBase58())
		return
	}
	if testWallet.GetAccountNum() != accountNum-1 {
		t.Errorf("TestClientDeleteAccount AccountNum:%d != %d", testWallet.GetAccountNum(), accountNum-1)
		return
	}
	accTmp, err := testWallet.GetAccountByAddress(delAcc.Address.ToBase58(), testPasswd)
	if err != nil {
		t.Errorf("TestClientDeleteAccount GetAccountByAddress:%s error:%s", delAcc.Address.ToBase58(), err)
		return
	}
	if accTmp != nil {
		t.Errorf("TestClientDeleteAccount GetAccountByAddress:%s should return nil", delAcc.Address.ToBase58())
		return
	}
}

func TestClientSetLabel(t *testing.T) {
	accountNum := testWallet.GetAccountNum()
	accountSize := 10
	if accountNum < accountSize {
		for i := accountSize - accountNum; i > accountNum; i-- {
			_, err := testWallet.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
			if err != nil {
				t.Errorf("TestClientSetLabel NewAccount error:%s", err)
				return
			}
		}
	}
	testAccIndex := 5
	testAcc := testWallet.GetAccountMetadataByIndex(testAccIndex)
	oldLabel := testAcc.Label
	newLabel := fmt.Sprintf("%s-%d", oldLabel, testAccIndex)

	accountNum = testWallet.GetAccountNum()
	err := testWallet.SetLabel(testAcc.Address, newLabel)
	if err != nil {
		t.Errorf("TestClientSetLabel SetLabel error:%s", err)
		return
	}

	if testWallet.GetAccountNum() != accountNum {
		t.Errorf("TestClientSetLabel account num %d != %d", testWallet.GetAccountNum(), accountNum)
		return
	}

	accTmp, err := testWallet.GetAccountByLabel(newLabel, testPasswd)
	if err != nil {
		t.Errorf("TestClientSetLabel GetAccountByLabel:%s error:%s", newLabel, err)
		return
	}
	if accTmp == nil {
		t.Errorf("TestClientSetLabel cannot get account by label:%s", newLabel)
		return
	}

	if accTmp.Address.ToBase58() != testAcc.Address {
		t.Errorf("TestClientSetLabel address:%s != %s", accTmp.Address.ToBase58(), testAcc.Address)
		return
	}

	accTmp, err = testWallet.GetAccountByLabel(oldLabel, testPasswd)
	if err != nil {
		t.Errorf("TestClientSetLabel GetAccountByLabel:%s error:%s", oldLabel, err)
		return
	}
	if accTmp != nil {
		t.Errorf("TestClientSetLabel GetAccountByLabel:%s should return nil", oldLabel)
		return
	}
}

func TestClientSetDefault(t *testing.T) {
	accountNum := testWallet.GetAccountNum()
	accountSize := 10
	if accountNum < accountSize {
		for i := accountSize - accountNum; i > accountNum; i-- {
			_, err := testWallet.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
			if err != nil {
				t.Errorf("TestClientSetDefault NewAccount error:%s", err)
				return
			}
		}
	}
	testAccIndex := 5
	testAcc, err := testWallet.GetAccountByIndex(testAccIndex, testPasswd)
	if err != nil {
		t.Errorf("TestClientSetDefault GetAccountByIndex:%d error:%s", testAccIndex, err)
		return
	}

	oldDefAcc, err := testWallet.GetDefaultAccount(testPasswd)
	if err != nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount error:%s", err)
		return
	}
	if oldDefAcc == nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount return nil")
		return
	}

	err = testWallet.SetDefaultAccount(testAcc.Address.ToBase58())
	if err != nil {
		t.Errorf("TestClientSetDefault SetDefaultAccount error:%s", err)
		return
	}

	defAcc, err := testWallet.GetDefaultAccount(testPasswd)
	if err != nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount error:%s", err)
		return
	}
	if defAcc == nil {
		t.Errorf("TestClientSetDefault GetDefaultAccount return nil")
		return
	}

	if defAcc.Address.ToBase58() != testAcc.Address.ToBase58() {
		t.Errorf("TestClientSetDefault address %s != %s", defAcc.Address.ToBase58(), testAcc.Address.ToBase58())
		return
	}

	accTmp := testWallet.GetAccountMetadataByAddress(oldDefAcc.Address.ToBase58())
	if accTmp.IsDefault {
		t.Errorf("TestClientSetDefault address:%s should not default account", accTmp.Address)
		return
	}

	accTmp = testWallet.GetAccountMetadataByAddress(testAcc.Address.ToBase58())
	if !accTmp.IsDefault {
		t.Errorf("TestClientSetDefault address:%s should be default account", accTmp.Address)
		return
	}
}

func TestImportAccount(t *testing.T) {
	walletPath2 := "tmp.dat"
	wallet2, err := NewClientImpl(walletPath2)
	if err != nil {
		t.Errorf("TestImportAccount NewClientImpl error:%s", err)
		return
	}
	defer os.Remove(walletPath2)

	acc1, err := wallet2.NewAccount("", keypair.PK_ECDSA, keypair.P256, s.SHA256withECDSA, testPasswd)
	if err != nil {
		t.Errorf("TestImportAccount NewAccount error:%s", err)
		return
	}
	accMetadata := wallet2.GetAccountMetadataByAddress(acc1.Address.ToBase58())
	if accMetadata == nil {
		t.Errorf("TestImportAccount GetAccountMetadataByAddress:%s return nil", acc1.Address.ToBase58())
		return
	}
	err = testWallet.ImportAccount(accMetadata)
	if err != nil {
		t.Errorf("TestImportAccount ImportAccount error:%s", err)
		return
	}

	acc, err := testWallet.GetAccountByAddress(accMetadata.Address, testPasswd)
	if err != nil {
		t.Errorf("TestImportAccount GetAccountByAddress error:%s", err)
		return
	}
	if acc == nil {
		t.Errorf("TestImportAccount failed, GetAccountByAddress return nil after import")
		return
	}
	assert.Equal(t, acc.Address.ToBase58() == acc1.Address.ToBase58(), true)
}

func TestCheckSigScheme(t *testing.T) {
	testClient, _ := NewClientImpl("")

	assert.Equal(t, testClient.checkSigScheme("ECDSA", "SHA224withECDSA"), true)
	assert.Equal(t, testClient.checkSigScheme("ECDSA", "SM3withSM2"), false)
	assert.Equal(t, testClient.checkSigScheme("SM2", "SM3withSM2"), true)
	assert.Equal(t, testClient.checkSigScheme("SM2", "SHA224withECDSA"), false)
	assert.Equal(t, testClient.checkSigScheme("Ed25519", "SHA512withEdDSA"), true)
	assert.Equal(t, testClient.checkSigScheme("Ed25519", "SHA224withECDSA"), false)
}
