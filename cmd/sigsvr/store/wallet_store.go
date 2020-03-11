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
package store

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/types"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type WalletStore struct {
	WalletName       string
	WalletVersion    string
	WalletScrypt     *keypair.ScryptParam
	WalletExtra      string
	path             string
	db               *leveldb.DB
	nextAccountIndex uint32
	lock             sync.RWMutex
}

func NewWalletStore(path string) (*WalletStore, error) {
	lvlOpts := &opt.Options{
		NoSync: false,
		Filter: filter.NewBloomFilter(10),
	}
	db, err := leveldb.OpenFile(path, lvlOpts)
	if err != nil {
		return nil, err
	}
	walletStore := &WalletStore{
		path: path,
		db:   db,
	}

	init, err := walletStore.isInit()
	if err != nil {
		return nil, err
	}
	if !init {
		walletStore.WalletName = DEFAULT_WALLET_NAME
		err = walletStore.setWalletName(walletStore.WalletName)
		if err != nil {
			return nil, fmt.Errorf("setWalletName error:%s", err)
		}
		walletStore.WalletVersion = WALLET_VERSION
		err = walletStore.setWalletVersion(walletStore.WalletVersion)
		if err != nil {
			return nil, fmt.Errorf("setWalletVersion error:%s", err)
		}
		walletStore.WalletScrypt = keypair.GetScryptParameters()
		err = walletStore.setWalletScrypt(walletStore.WalletScrypt)
		if err != nil {
			return nil, fmt.Errorf("setWalletScrypt error:%s", err)
		}
		walletStore.WalletExtra = ""
		err = walletStore.setWalletExtra(walletStore.WalletExtra)
		if err != nil {
			return nil, fmt.Errorf("setWalletExtra error:%s", err)
		}
		err = walletStore.init()
		if err != nil {
			return nil, fmt.Errorf("init error:%s", err)
		}
		return walletStore, nil
	}
	nextAccountIndex, err := walletStore.getNextAccountIndex()
	if err != nil {
		return nil, fmt.Errorf("getNextAccountIndex error:%s", err)
	}
	walletName, err := walletStore.getWalletName()
	if err != nil {
		return nil, fmt.Errorf("getWalletName error:%s", err)
	}
	walletVersion, err := walletStore.getWalletVersion()
	if err != nil {
		return nil, fmt.Errorf("getWalletVersion error:%s", err)
	}
	walletScrypt, err := walletStore.getWalletScrypt()
	if err != nil {
		return nil, fmt.Errorf("getWalletScrypt error:%s", err)
	}
	walletExtra, err := walletStore.getWalletExtra()
	if err != nil {
		return nil, fmt.Errorf("getWalletExtra error: %v", err)
	}
	walletStore.nextAccountIndex = nextAccountIndex
	walletStore.WalletScrypt = walletScrypt
	walletStore.WalletName = walletName
	walletStore.WalletVersion = walletVersion
	walletStore.WalletExtra = walletExtra
	return walletStore, nil
}

func (this *WalletStore) isInit() (bool, error) {
	data, err := this.db.Get(GetWalletInitKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	if string(data) != WALLET_INIT_DATA {
		return false, fmt.Errorf("init not success")
	}
	return true, nil
}

func (this *WalletStore) init() error {
	return this.db.Put(GetWalletInitKey(), []byte(WALLET_INIT_DATA), nil)
}

func (this *WalletStore) setWalletVersion(version string) error {
	return this.db.Put(GetWalletVersionKey(), []byte(version), nil)
}

func (this *WalletStore) getWalletVersion() (string, error) {
	data, err := this.db.Get(GetWalletVersionKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *WalletStore) setWalletName(name string) error {
	return this.db.Put(GetWalletNameKey(), []byte(name), nil)
}

func (this *WalletStore) getWalletName() (string, error) {
	data, err := this.db.Get(GetWalletNameKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *WalletStore) setWalletScrypt(scrypt *keypair.ScryptParam) error {
	data, err := json.Marshal(scrypt)
	if err != nil {
		return err
	}
	return this.db.Put(GetWalletScryptKey(), data, nil)
}

func (this *WalletStore) getWalletScrypt() (*keypair.ScryptParam, error) {
	data, err := this.db.Get(GetWalletScryptKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	scypt := &keypair.ScryptParam{}
	err = json.Unmarshal(data, scypt)
	if err != nil {
		return nil, err
	}
	return scypt, nil
}

func (this *WalletStore) setWalletExtra(extra string) error {
	return this.db.Put(GetWalletExtraKey(), []byte(extra), nil)
}

func (this *WalletStore) getWalletExtra() (string, error) {
	data, err := this.db.Get(GetWalletExtraKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *WalletStore) GetNextAccountIndex() uint32 {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.nextAccountIndex
}

func (this *WalletStore) GetAccountByAddress(address string, passwd []byte) (*account.Account, error) {
	accData, err := this.GetAccountDataByAddress(address)
	if err != nil {
		return nil, err
	}
	if accData == nil {
		return nil, nil
	}
	privateKey, err := keypair.DecryptWithCustomScrypt(&accData.ProtectedKey, passwd, this.WalletScrypt)
	if err != nil {
		return nil, fmt.Errorf("decrypt PrivateKey error:%s", err)
	}
	publicKey := privateKey.Public()
	addr := types.AddressFromPubKey(publicKey)
	scheme, err := s.GetScheme(accData.SigSch)
	if err != nil {
		return nil, fmt.Errorf("signature scheme error:%s", err)
	}
	return &account.Account{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    addr,
		SigScheme:  scheme,
	}, nil
}

func (this *WalletStore) NewAccountData(typeCode keypair.KeyType, curveCode byte, sigScheme s.SignatureScheme, passwd []byte) (*account.AccountData, error) {
	if len(passwd) == 0 {
		return nil, fmt.Errorf("password cannot empty")
	}
	prvkey, pubkey, err := keypair.GenerateKeyPair(typeCode, curveCode)
	if err != nil {
		return nil, fmt.Errorf("generateKeyPair error:%s", err)
	}
	address := types.AddressFromPubKey(pubkey)
	addressBase58 := address.ToBase58()
	prvSecret, err := keypair.EncryptWithCustomScrypt(prvkey, addressBase58, passwd, this.WalletScrypt)
	if err != nil {
		return nil, fmt.Errorf("encryptPrivateKey error:%s", err)
	}
	accData := &account.AccountData{}
	accData.SetKeyPair(prvSecret)
	accData.SigSch = sigScheme.Name()
	accData.PubKey = hex.EncodeToString(keypair.SerializePublicKey(pubkey))

	return accData, nil
}

func (this *WalletStore) AddAccountData(accData *account.AccountData) (bool, error) {
	isExist, err := this.IsAccountExist(accData.Address)
	if err != nil {
		return false, err
	}

	this.lock.Lock()
	defer this.lock.Unlock()

	accountNum, err := this.GetAccountNumber()
	if err != nil {
		return false, fmt.Errorf("GetAccountNumber error:%s", err)
	}
	if accountNum == 0 {
		accData.IsDefault = true
	} else {
		accData.IsDefault = false
	}
	data, err := json.Marshal(accData)
	if err != nil {
		return false, err
	}

	batch := &leveldb.Batch{}
	//Put account
	batch.Put(GetAccountKey(accData.Address), data)

	nextIndex := this.nextAccountIndex
	if !isExist {
		accountIndex := nextIndex
		//Put account index
		batch.Put(GetAccountIndexKey(accountIndex), []byte(accData.Address))

		nextIndex++
		data = make([]byte, 4, 4)
		binary.LittleEndian.PutUint32(data, nextIndex)

		//Put next account index
		batch.Put(GetNextAccountIndexKey(), data)

		accountNum++
		binary.LittleEndian.PutUint32(data, accountNum)

		//Put account number
		batch.Put(GetWalletAccountNumberKey(), data)
	}

	err = this.db.Write(batch, nil)
	if err != nil {
		return false, err
	}
	this.nextAccountIndex = nextIndex

	isAdd := !isExist
	return isAdd, nil
}

func (this *WalletStore) getNextAccountIndex() (uint32, error) {
	data, err := this.db.Get(GetNextAccountIndexKey(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}

func (this *WalletStore) GetAccountDataByAddress(address string) (*account.AccountData, error) {
	data, err := this.db.Get(GetAccountKey(address), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	accData := &account.AccountData{}
	err = json.Unmarshal(data, accData)
	if err != nil {
		return nil, err
	}
	return accData, nil
}

func (this *WalletStore) IsAccountExist(address string) (bool, error) {
	data, err := this.db.Get(GetAccountKey(address), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return len(data) != 0, nil
}

func (this *WalletStore) GetAccountDataByIndex(index uint32) (*account.AccountData, error) {
	address, err := this.GetAccountAddress(index)
	if err != nil {
		return nil, err
	}
	if address == "" {
		return nil, nil
	}
	return this.GetAccountDataByAddress(address)
}

func (this *WalletStore) GetAccountAddress(index uint32) (string, error) {
	data, err := this.db.Get(GetAccountIndexKey(index), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func (this *WalletStore) setAccountNumber(number uint32) error {
	data := make([]byte, 4, 4)
	binary.LittleEndian.PutUint32(data, number)
	return this.db.Put(GetWalletAccountNumberKey(), data, nil)
}

func (this *WalletStore) GetAccountNumber() (uint32, error) {
	data, err := this.db.Get(GetWalletAccountNumberKey(), nil)
	if err == nil {
		return binary.LittleEndian.Uint32(data), nil
	}
	if err != leveldb.ErrNotFound {
		return 0, err
	}
	//Keep downward compatible
	nextIndex, err := this.getNextAccountIndex()
	if err != nil {
		return 0, fmt.Errorf("getNextAccountIndex error:%s", err)
	}
	if nextIndex == 0 {
		return 0, nil
	}
	addresses := make(map[string]string, 0)
	for i := uint32(0); i < nextIndex; i++ {
		address, err := this.GetAccountAddress(i)
		if err != nil {
			return 0, fmt.Errorf("GetAccountAddress Index:%d error:%s", i, err)
		}
		if address == "" {
			continue
		}
		addresses[address] = ""
	}
	accNum := uint32(len(addresses))
	err = this.setAccountNumber(accNum)
	if err != nil {
		return 0, fmt.Errorf("setAccountNumber error")
	}
	return accNum, nil
}
