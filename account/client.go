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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"sync"
	"time"
)

//Client of wallet
type Client interface {
	//NewAccount create a new account.
	NewAccount(label string, typeCode keypair.KeyType, curveCode byte, sigScheme s.SignatureScheme, passwd []byte) (*Account, error)
	//GetAccountByAddress return account object by address
	GetAccountByAddress(address string, passwd []byte) (*Account, error)
	//GetAccountByLabel return account object by label
	GetAccountByLabel(label string, passwd []byte) (*Account, error)
	//GetAccountByIndex return account object by index. Index start from 1
	GetAccountByIndex(index int, passwd []byte) (*Account, error)
	//GetDefaultAccount return default account
	GetDefaultAccount(passwd []byte) (*Account, error)
	//GetAccountPublicByIndex return account public info by address
	GetAccountPublicByAddress(address string) (*AccountPublic, error)
	//GetAccountPublicByLabel return account public info by label
	GetAccountPublicByLabel(label string) (*AccountPublic, error)
	//GetAccountPublicByIndex return account public info by index. Index start from 1
	GetAccountPublicByIndex(index int) (*AccountPublic, error)
	//GetDefaultAccountPublic return default account public info
	GetDefaultAccountPublic() (*AccountPublic, error)
	//GetAccountNum return total account number
	GetAccountNum() int
	//DeleteAccount delete account
	DeleteAccount(address string, passwd []byte) (*Account, error)
	//UnLockAccount can get account without password in expire time
	UnLockAccount(address string, expiredAt int, passwd []byte) error
	//LockAccount lock unlock account
	LockAccount(address string)
	//GetUnlockAccount return account which was unlock and in expired time
	GetUnlockAccount(address string) *Account
	//Set a new account to default account
	SetDefaultAccount(address string, passwd []byte) error
	//Set a new label to accont
	SetLabel(address, label string, passwd []byte) error
	//Change pasword to account
	ChangePassword(address string, oldPasswd, newPasswd []byte) error
	//Change sig scheme to account
	ChangeSigScheme(address string, sigScheme s.SignatureScheme, passwd []byte) error
}

func Open(path string) (Client, error) {
	return NewClientImpl(path)
}

type unlockAccountInfo struct {
	acc        *Account
	unlockTime time.Time
	expiredAt  int //s
}

func (this *unlockAccountInfo) isAvail() bool {
	return int(time.Now().Sub(this.unlockTime).Seconds()) < this.expiredAt
}

type ClientImpl struct {
	path       string
	accAddrs   map[string]*AccountData //Map Address(base58) => Account
	accLabels  map[string]*AccountData //Map Label => Account
	defaultAcc *AccountData
	walletData *WalletData
	unlockAccs map[string]*unlockAccountInfo //Map Address(base58) => unlockAccountInfo
	lock       sync.RWMutex
}

func NewClientImpl(path string) (*ClientImpl, error) {
	cli := &ClientImpl{
		path:       path,
		accAddrs:   make(map[string]*AccountData),
		accLabels:  make(map[string]*AccountData),
		unlockAccs: make(map[string]*unlockAccountInfo),
		walletData: NewWalletData(),
	}
	if common.FileExisted(path) {
		err := cli.load()
		if err != nil {
			return nil, err
		}
	}
	return cli, nil
}

func (this *ClientImpl) load() error {
	err := this.walletData.Load(this.path)
	if err != nil {
		return fmt.Errorf("load wallet:%s error:%s", this.path, err)
	}
	for _, accData := range this.walletData.Accounts {
		this.accAddrs[accData.Address] = accData
		if accData.Label != "" {
			this.accLabels[accData.Label] = accData
		}
		if accData.IsDefault {
			this.defaultAcc = accData
		}
	}
	return nil
}

func (this *ClientImpl) save() error {
	return this.walletData.Save(this.path)
}

func (this *ClientImpl) NewAccount(label string, typeCode keypair.KeyType, curveCode byte, sigScheme s.SignatureScheme, passwd []byte) (*Account, error) {
	if len(passwd) == 0 {
		return nil, fmt.Errorf("password cannot empty")
	}
	prvkey, pubkey, err := keypair.GenerateKeyPair(typeCode, curveCode)
	if err != nil {
		return nil, fmt.Errorf("generateKeyPair error:%s", err)
	}
	address := types.AddressFromPubKey(pubkey)
	addressBase58 := address.ToBase58()
	prvSecret, err := keypair.EncryptPrivateKey(prvkey, addressBase58, passwd)
	if err != nil {
		return nil, fmt.Errorf("encryptPrivateKey error:%s", err)
	}
	accData := &AccountData{}
	accData.Label = label
	accData.SetKeyPair(prvSecret)
	accData.SigSch = sigScheme.Name()
	accData.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))
	passHash := sha256.Sum256(passwd)
	accData.PassHash = hex.EncodeToString(passHash[:])

	this.lock.Lock()
	defer this.lock.Unlock()
	if label != "" {
		_, ok := this.accLabels[label]
		if ok {
			return nil, fmt.Errorf("duplicate label")
		}
	}
	if len(this.walletData.Accounts) == 0 {
		accData.IsDefault = true
	}
	this.walletData.AddAccount(accData)
	err = this.save()
	if err != nil {
		this.walletData.DelAccount(addressBase58)
		return nil, fmt.Errorf("save error:%s", err)
	}
	this.accAddrs[addressBase58] = accData
	if accData.IsDefault {
		this.defaultAcc = accData
	}
	if label != "" {
		this.accLabels[label] = accData
	}
	return &Account{
		Label:      label,
		KeyType:    accData.Alg,
		Curve:      accData.Param["curve"],
		PrivateKey: prvkey,
		PublicKey:  pubkey,
		Address:    address,
		SigScheme:  sigScheme,
	}, nil
}

func (this *ClientImpl) GetAccountByAddress(address string, passwd []byte) (*Account, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData, ok := this.accAddrs[address]
	if !ok {
		return nil, nil
	}
	return this.getAccount(accData, passwd)
}

func (this *ClientImpl) GetAccountByLabel(label string, passwd []byte) (*Account, error) {
	if len(label) == 0 {
		return nil, nil
	}
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData, ok := this.accLabels[label]
	if !ok {
		return nil, nil
	}
	return this.getAccount(accData, passwd)
}

//Index start from 1
func (this *ClientImpl) GetAccountByIndex(index int, passwd []byte) (*Account, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData := this.walletData.GetAccountByIndex(index - 1)
	if accData == nil {
		return nil, nil
	}
	return this.getAccount(accData, passwd)
}

func (this *ClientImpl) GetDefaultAccount(passwd []byte) (*Account, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if this.defaultAcc == nil {
		return nil, nil
	}
	return this.getAccount(this.defaultAcc, passwd)
}

func (this *ClientImpl) GetAccountPublicByAddress(address string) (*AccountPublic, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData, ok := this.accAddrs[address]
	if !ok {
		return nil, nil
	}
	return this.getAccountPublic(accData)
}

func (this *ClientImpl) GetAccountPublicByLabel(label string) (*AccountPublic, error) {
	if label == "" {
		return nil, nil
	}
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData, ok := this.accLabels[label]
	if !ok {
		return nil, nil
	}
	return this.getAccountPublic(accData)
}

//Index start from 1
func (this *ClientImpl) GetAccountPublicByIndex(index int) (*AccountPublic, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	accData := this.walletData.GetAccountByIndex(index - 1)
	if accData == nil {
		return nil, nil
	}
	return this.getAccountPublic(accData)
}

func (this *ClientImpl) GetDefaultAccountPublic() (*AccountPublic, error) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if this.defaultAcc != nil {
		return this.getAccountPublic(this.defaultAcc)
	}
	return nil, nil
}

func (this *ClientImpl) getAccount(accData *AccountData, passwd []byte) (*Account, error) {
	if !accData.VerifyPassword(passwd) {
		return nil, fmt.Errorf("verify password failed")
	}
	privateKey, err := keypair.DecryptPrivateKey(&accData.ProtectedKey, passwd)
	if err != nil {
		return nil, fmt.Errorf("decrypt PrivateKey error:%s", err)
	}
	publicKey := privateKey.Public()
	addr := types.AddressFromPubKey(publicKey)
	scheme, err := s.GetScheme(accData.SigSch)
	if err != nil {
		return nil, fmt.Errorf("signature scheme error:%s", err)
	}
	isDefault := false
	if this.defaultAcc != nil {
		isDefault = accData.Address == this.defaultAcc.Address
	}
	return &Account{
		Label:      accData.Label,
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    addr,
		SigScheme:  scheme,
		IsDefault:  isDefault,
	}, nil
}

func (this *ClientImpl) getAccountPublic(accData *AccountData) (*AccountPublic, error) {
	scheme, err := s.GetScheme(accData.SigSch)
	if err != nil {
		return nil, fmt.Errorf("signature scheme error:%s", err)
	}
	pubKeyData, err := hex.DecodeString(accData.PubKey)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	pubKey, err := keypair.DeserializePublicKey(pubKeyData)
	if err != nil {
		return nil, fmt.Errorf("keypair.DeserializePublicKey error:%s", err)
	}
	addr, err := common.AddressFromBase58(accData.Address)
	if err != nil {
		return nil, fmt.Errorf("addressFromBase58:%s error:%s", accData.Address, err)
	}
	curve := accData.Param["curve"]
	isDefault := false
	if this.defaultAcc != nil {
		isDefault = accData.Address == this.defaultAcc.Address
	}
	return &AccountPublic{
		Label:     accData.Label,
		KeyType:   accData.Alg,
		Curve:     curve,
		PublicKey: pubKey,
		CipherKey: accData.Key,
		SigScheme: scheme,
		Address:   addr,
		IsDefault: isDefault,
	}, nil
}

func (this *ClientImpl) GetAccountNum() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return len(this.accAddrs)
}

func (this *ClientImpl) DeleteAccount(address string, passwd []byte) (*Account, error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	accData, ok := this.accAddrs[address]
	if !ok {
		return nil, nil
	}
	if accData.IsDefault {
		return nil, fmt.Errorf("cannot delete default account")
	}
	acc, err := this.getAccount(accData, passwd)
	if err != nil {
		return nil, err
	}

	bkAccList := append([]*AccountData{}, this.walletData.Accounts...)
	this.walletData.DelAccount(address)
	err = this.save()
	if err != nil {
		this.walletData.Accounts = bkAccList
		return nil, err
	}
	delete(this.accAddrs, address)
	if accData.Label != "" {
		delete(this.accLabels, accData.Label)
	}
	delete(this.unlockAccs, address)
	return acc, nil
}

func (this *ClientImpl) UnLockAccount(address string, expiredAt int, passwd []byte) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	accData, ok := this.accAddrs[address]
	if !ok {
		return fmt.Errorf("cannot find account by address:%s", address)
	}
	if expiredAt < 0 {
		return fmt.Errorf("invalid expired time")
	}
	acc, err := this.getAccount(accData, passwd)
	if err != nil {
		return err
	}
	this.unlockAccs[address] = &unlockAccountInfo{
		acc:        acc,
		expiredAt:  expiredAt,
		unlockTime: time.Now(),
	}
	return nil
}

func (this *ClientImpl) LockAccount(address string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.unlockAccs, address)
}

func (this *ClientImpl) GetUnlockAccount(address string) *Account {
	this.lock.Lock()
	defer this.lock.Unlock()
	accInfo, ok := this.unlockAccs[address]
	if !ok {
		return nil
	}
	if !accInfo.isAvail() {
		delete(this.unlockAccs, address)
		return nil
	}
	return accInfo.acc
}

func (this *ClientImpl) SetDefaultAccount(address string, passwd []byte) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.defaultAcc != nil && this.defaultAcc.Address == address {
		return nil
	}
	accData, ok := this.accAddrs[address]
	if !ok {
		return fmt.Errorf("cannot find account by address:%s", address)
	}
	if !accData.VerifyPassword(passwd) {
		return fmt.Errorf("verifyPassword failed")
	}
	old := this.defaultAcc
	if old != nil {
		old.IsDefault = false
	}
	this.defaultAcc = accData
	accData.IsDefault = true
	err := this.save()
	if err != nil {
		this.defaultAcc = old
		if old != nil {
			old.IsDefault = true
		}
		accData.IsDefault = false
		return fmt.Errorf("save error:%s", err)
	}
	return nil
}

func (this *ClientImpl) SetLabel(address, label string, passwd []byte) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	_, ok := this.accLabels[label]
	if ok {
		return fmt.Errorf("duplicate label")
	}
	accData, ok := this.accAddrs[address]
	if !ok {
		return fmt.Errorf("cannot find account by address:%s", address)
	}
	if accData.Label == label {
		return nil
	}
	oldLabel := accData.Label
	if !accData.VerifyPassword(passwd) {
		return fmt.Errorf("VerifyPassword failed")
	}
	accData.Label = label
	err := this.save()
	if err != nil {
		accData.Label = oldLabel
		return fmt.Errorf("save error:%s", err)
	}
	delete(this.accLabels, oldLabel)
	this.accLabels[label] = accData
	acc, ok := this.unlockAccs[address]
	if ok {
		acc.acc.Label = label
	}
	return nil
}

func (this *ClientImpl) ChangePassword(address string, oldPasswd, newPasswd []byte) error {
	if bytes.Equal(oldPasswd, newPasswd) {
		return nil
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	accData, ok := this.accAddrs[address]
	if !ok {
		return fmt.Errorf("cannot find account by address:%s", address)
	}
	if !accData.VerifyPassword(oldPasswd) {
		return fmt.Errorf("verifyPassword failed")
	}
	oldPrvSecret := accData.GetKeyPair()
	oldPasswdHash := accData.PassHash
	prv, err := keypair.DecryptPrivateKey(accData.GetKeyPair(), oldPasswd)
	if err != nil {
		return fmt.Errorf("keypair.DecryptPrivateKey error:%s", err)
	}
	newPrvSecret, err := keypair.EncryptPrivateKey(prv, address, newPasswd)
	if err != nil {
		return fmt.Errorf("keypair.EncryptPrivateKey error:%s", err)
	}
	newPassHash := sha256.Sum256(newPasswd)
	accData.PassHash = hex.EncodeToString(newPassHash[:])
	newPasswdHash := hex.EncodeToString(newPassHash[:])
	accData.SetKeyPair(newPrvSecret)
	accData.PassHash = newPasswdHash
	err = this.save()
	if err != nil {
		accData.PassHash = oldPasswdHash
		accData.SetKeyPair(oldPrvSecret)
		return fmt.Errorf("save error:%s", err)
	}
	return nil
}

func (this *ClientImpl) ChangeSigScheme(address string, sigScheme s.SignatureScheme, passwd []byte) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	accData, ok := this.accAddrs[address]
	if !ok {
		return fmt.Errorf("cannot find account by address:%s", address)
	}
	if !accData.VerifyPassword(passwd) {
		return fmt.Errorf("verifyPassword failed")
	}
	oldSigScheme := accData.SigSch
	accData.SigSch = sigScheme.Name()
	err := this.save()
	if err != nil {
		accData.SigSch = oldSigScheme
		return fmt.Errorf("save error:%s", err)
	}
	accInfo, ok := this.unlockAccs[address]
	if ok {
		accInfo.acc.SigScheme = sigScheme
	}
	return nil
}
