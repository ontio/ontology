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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
)

/** AccountData - for wallet read and save, no crypto object included **/
type AccountData struct {
	keypair.ProtectedKey

	Label     string `json:"label"`
	PubKey    string `json:"publicKey"`
	SigSch    string `json:"signatureScheme"`
	IsDefault bool   `json:"isDefault"`
	Lock      bool   `json:"lock"`
}

func (this *AccountData) SetKeyPair(keyinfo *keypair.ProtectedKey) {
	this.Address = keyinfo.Address
	this.EncAlg = keyinfo.EncAlg
	this.Alg = keyinfo.Alg
	this.Hash = keyinfo.Hash
	this.Key = keyinfo.Key
	this.Param = keyinfo.Param
	this.Salt = keyinfo.Salt
}

func (this *AccountData) GetKeyPair() *keypair.ProtectedKey {
	var keyinfo = new(keypair.ProtectedKey)
	keyinfo.Address = this.Address
	keyinfo.EncAlg = this.EncAlg
	keyinfo.Alg = this.Alg
	keyinfo.Hash = this.Hash
	keyinfo.Key = this.Key
	keyinfo.Param = this.Param
	keyinfo.Salt = this.Salt
	return keyinfo
}

func (this *AccountData) SetLabel(label string) {
	this.Label = label
}

type WalletData struct {
	Name       string               `json:"name"`
	Version    string               `json:"version"`
	Scrypt     *keypair.ScryptParam `json:"scrypt"`
	Identities []Identity           `json:"identities,omitempty"`
	Accounts   []*AccountData       `json:"accounts,omitempty"`
	Extra      string               `json:"extra,omitempty"`
}

func NewWalletData() *WalletData {
	return &WalletData{
		Name:       "MyWallet",
		Version:    "1.1",
		Scrypt:     keypair.GetScryptParameters(),
		Identities: nil,
		Extra:      "",
		Accounts:   make([]*AccountData, 0, 0),
	}
}

func (this *WalletData) Clone() *WalletData {
	w := WalletData{}
	w.Name = this.Name
	w.Version = this.Version
	sp := *this.Scrypt
	w.Scrypt = &sp
	w.Accounts = make([]*AccountData, len(this.Accounts))
	for i, v := range this.Accounts {
		ac := *v
		ac.SetKeyPair(v.GetKeyPair())
		w.Accounts[i] = &ac
	}
	w.Identities = this.Identities
	w.Extra = this.Extra
	return &w
}

func (this *WalletData) AddAccount(acc *AccountData) {
	this.Accounts = append(this.Accounts, acc)
}

func (this *WalletData) DelAccount(address string) {
	_, index := this.GetAccountByAddress(address)
	if index < 0 {
		return
	}
	this.Accounts = append(this.Accounts[:index], this.Accounts[index+1:]...)
}

func (this *WalletData) GetAccountByIndex(index int) *AccountData {
	if index < 0 || index >= len(this.Accounts) {
		return nil
	}
	return this.Accounts[index]
}

func (this *WalletData) GetAccountByAddress(address string) (*AccountData, int) {
	index := -1
	var accData *AccountData
	for i, acc := range this.Accounts {
		if acc.Address == address {
			index = i
			accData = acc
			break
		}
	}
	if index == -1 {
		return nil, -1
	}
	return accData, index
}

// defaultWalletFilePerm is the permission used for wallet files. A wallet holds
// encrypted private key material, so it must only be readable by its owner.
const defaultWalletFilePerm os.FileMode = 0600

// walletFilePerm returns the permission to use when writing a wallet to path.
// It defaults to 0600 and keeps the existing file's permission when that is
// strictly more restrictive (for example 0400), so saving never widens access
// to an already locked-down wallet.
func walletFilePerm(path string) os.FileMode {
	if info, err := os.Stat(path); err == nil {
		if existing := info.Mode().Perm(); existing&^defaultWalletFilePerm == 0 {
			return existing
		}
	}
	return defaultWalletFilePerm
}

func (this *WalletData) Save(path string) error {
	data, err := json.Marshal(this)
	if err != nil {
		return err
	}
	perm := walletFilePerm(path)
	if common.FileExisted(path) {
		filename := path + "~"
		err := ioutil.WriteFile(filename, data, perm)
		if err != nil {
			return err
		}
		// WriteFile ignores the permission for an existing file, so a leftover
		// temp file (or a umask) could otherwise leave the renamed wallet with
		// looser permissions than intended.
		if err := os.Chmod(filename, perm); err != nil {
			return err
		}
		return os.Rename(filename, path)
	} else {
		return ioutil.WriteFile(path, data, perm)
	}
}

func (this *WalletData) Load(path string) error {
	msh, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(msh, this)
}

var lowSecurityParam = keypair.ScryptParam{
	N:     4096,
	R:     8,
	P:     8,
	DKLen: 64,
}

func (this *WalletData) ToLowSecurity(passwords [][]byte) error {
	return this.reencrypt(passwords, &lowSecurityParam)
}

func (this *WalletData) ToDefaultSecurity(passwords [][]byte) error {
	return this.reencrypt(passwords, nil)
}

func (this *WalletData) reencrypt(passwords [][]byte, param *keypair.ScryptParam) error {
	if len(passwords) != len(this.Accounts) {
		return errors.New("not enough passwords for the accounts")
	}
	keys := make([]*keypair.ProtectedKey, len(this.Accounts))
	for i, v := range this.Accounts {
		prot, err := keypair.ReencryptPrivateKey(&v.ProtectedKey, passwords[i], passwords[i], this.Scrypt, param)
		if err != nil {
			return fmt.Errorf("re-encrypt account %d failed: %s", i, err)
		}
		keys[i] = prot
	}

	for i, v := range keys {
		this.Accounts[i].SetKeyPair(v)
	}
	if param != nil {
		this.Scrypt = param
	} else {
		// default parameters
		this.Scrypt = keypair.GetScryptParameters()
	}
	return nil
}

func (this *WalletData) AddIdentity(id *Identity) {
	this.Identities = append(this.Identities, *id)
}
