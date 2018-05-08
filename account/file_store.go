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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
)

/** AccountData - for wallet read and save, no crypto object included **/
type AccountData struct {
	keypair.ProtectedKey

	Label     string `json:"label"`
	PubKey    string `json:"publicKey"`
	SigSch    string `json:"signatureScheme"`
	IsDefault bool   `json:"isDefault"`
	Lock      bool   `json:"lock"`
	PassHash  string `json:"passwordHash"`
}

func (this *AccountData) SetKeyPair(keyinfo *keypair.ProtectedKey) {
	this.Address = keyinfo.Address
	this.EncAlg = keyinfo.EncAlg
	this.Alg = keyinfo.Alg
	this.Hash = keyinfo.Hash
	this.Key = keyinfo.Key
	this.Param = keyinfo.Param
}
func (this *AccountData) GetKeyPair() *keypair.ProtectedKey {
	var keyinfo = new(keypair.ProtectedKey)
	keyinfo.Address = this.Address
	keyinfo.EncAlg = this.EncAlg
	keyinfo.Alg = this.Alg
	keyinfo.Hash = this.Hash
	keyinfo.Key = this.Key
	keyinfo.Param = this.Param
	return keyinfo
}
func (this *AccountData) VerifyPassword(pwd []byte) bool {
	passwordHash := sha256.Sum256(pwd)
	if this.PassHash != hex.EncodeToString(passwordHash[:]) {
		return false
	}
	return true
}

func (this *AccountData) SetLabel(label string) {
	this.Label = label
}

func CreateAccount(TypeCode keypair.KeyType, CurveCode byte, SchemeName string, password *[]byte) *AccountData {
	prvkey, pubkey, _ := keypair.GenerateKeyPair(TypeCode, CurveCode)
	ta := types.AddressFromPubKey(pubkey)
	address := ta.ToBase58()

	prvSecret, _ := keypair.EncryptPrivateKey(prvkey, address, *password)
	h := sha256.Sum256(*password)

	var acc = new(AccountData)
	acc.SetKeyPair(prvSecret)
	acc.SigSch = SchemeName
	acc.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
	acc.PassHash = hex.EncodeToString(h[:])

	return acc
}

type WalletData struct {
	Name       string               `json:"name"`
	Version    string               `json:"version"`
	Scrypt     *keypair.ScryptParam `json:"scrypt"`
	Identities []Identity           `json:"identities"`
	Accounts   []*AccountData       `json:"accounts"`
	Extra      string               `json:"extra"`
}

//TODO:: for temporary use, these params should be set by user?
func (this *WalletData) Inititalize() {
	this.Name = "MyWallet"
	this.Version = "1.1"
	this.Scrypt = keypair.GetScryptParameters()
	this.Identities = nil
	this.Extra = "null"
	this.Accounts = make([]*AccountData, 0, 0)
}

func (this *WalletData) AddAccount(acc *AccountData) {
	if len(this.Accounts) == 0 {
		acc.IsDefault = true
	}
	this.Accounts = append(this.Accounts, acc)
}

func (this *WalletData) DelAccount(index int) string {
	addr := this.Accounts[index-1].Address
	this.Accounts = append(this.Accounts[:index-1], this.Accounts[index:]...)
	return addr
}

func (this *WalletData) GetDefaultAccount() *AccountData {
	for _, i := range this.Accounts {
		if i.IsDefault {
			return i
		}
	}
	return nil
}

func (this *WalletData) Save(path string) error {
	data, err := json.Marshal(this)
	if err != nil {
		return err
	}
	if common.FileExisted(path) {
		filename := path + "~"
		err := ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}
		return os.Rename(filename, path)
	} else {
		return ioutil.WriteFile(path, data, 0644)
	}
}

func (this *WalletData) Load(path string) error {
	msh, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(msh, this)
}

//TODO:: determine identity structure
type Identity struct{}
