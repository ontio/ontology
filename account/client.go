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
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/ontio/ontology-crypto/aes"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/password"
	"github.com/ontio/ontology/core/types"
	ontErrors "github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/p2pserver/protocol"
)

const (
	DEFAULT_BOOKKEEPER_COUNT = 4
	WALLET_FILENAME          = "wallet.dat"
)

type Client interface {
	ContainsAccount(pubKey keypair.PublicKey) bool
	GetAccount(pubKey keypair.PublicKey) *Account
	GetDefaultAccount() *Account
	GetBookkeepers() ([]keypair.PublicKey, error)
}

type ClientImpl struct {
	mu sync.Mutex

	path      string
	iv        []byte
	masterKey []byte

	accounts map[common.Address]*Account

	watchOnly     []common.Address
	currentHeight uint32

	FileStore
	isrunning bool
}

//TODO: adjust contract folder structure
func Create(path string, encrypt string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, true)

	_, err := cl.CreateAccount(encrypt)
	if err != nil {
		fmt.Println(err)
	}

	return cl
}

func Open(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, false)
	if cl == nil {
		log.Error("Alloc new client failure")
		return nil
	}

	cl.accounts = cl.LoadAccount()
	if cl.accounts == nil {
		log.Error("Load accounts failure")
	}
	return cl
}

func NewClient(path string, password []byte, create bool) *ClientImpl {
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[common.Address]*Account{},
		FileStore: FileStore{path: path},
		isrunning: true,
	}

	passwordKey := doubleHash(password)
	if create {
		//create new client
		newClient.iv = make([]byte, 16)
		newClient.masterKey = make([]byte, 32)
		newClient.watchOnly = []common.Address{}
		newClient.currentHeight = 0

		//generate random number for iv/masterkey
		_, err := rand.Read(newClient.iv)
		if err != nil {
			log.Error(err)
			return nil
		}
		_, err = rand.Read(newClient.masterKey)
		if err != nil {
			log.Error(err)
			return nil
		}

		//new client store (build DB)
		newClient.BuildDatabase(path)

		// SaveStoredData
		pwdhash := sha256.Sum256(passwordKey)
		err = newClient.SaveStoredData("PasswordHash", pwdhash[:])
		if err != nil {
			log.Error(err)
			return nil
		}
		err = newClient.SaveStoredData("IV", newClient.iv[:])
		if err != nil {
			log.Error(err)
			return nil
		}

		aesmk, err := aes.AesEncrypt(newClient.masterKey[:], passwordKey, newClient.iv)
		if err == nil {
			err = newClient.SaveStoredData("MasterKey", aesmk)
			if err != nil {
				log.Error(err)
				return nil
			}
		} else {
			log.Error(err)
			return nil
		}
	} else {
		if b := newClient.verifyPasswordKey(passwordKey); b == false {
			return nil
		}
		if err := newClient.loadClient(passwordKey); err != nil {
			return nil
		}
	}
	clearBytes(passwordKey, len(passwordKey))
	return newClient
}

func (cl *ClientImpl) loadClient(passwordKey []byte) error {
	var err error
	cl.iv, err = cl.LoadStoredData("IV")
	if err != nil {
		fmt.Println("error: failed to load iv")
		return err
	}
	encryptedMasterKey, err := cl.LoadStoredData("MasterKey")
	if err != nil {
		fmt.Println("error: failed to load master key")
		return err
	}
	cl.masterKey, err = aes.AesDecrypt(encryptedMasterKey, passwordKey, cl.iv)
	if err != nil {
		fmt.Println("error: failed to decrypt master key")
		return err
	}
	return nil
}

func (cl *ClientImpl) GetDefaultAccount() *Account {
	// todo the iteration of map is not ordered
	for programHash := range cl.accounts {
		return cl.GetAccountByAddress(programHash)
	}

	return nil
}

func (cl *ClientImpl) GetAccount(pubKey keypair.PublicKey) *Account {
	address := types.AddressFromPubKey(pubKey)
	return cl.GetAccountByAddress(address)
}

func (cl *ClientImpl) GetAccountByAddress(address common.Address) *Account {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if account, ok := cl.accounts[address]; ok {
		return account
	}
	return nil
}

func (cl *ClientImpl) ChangePassword(oldPassword []byte, newPassword []byte) bool {
	// check password
	oldPasswordKey := doubleHash(oldPassword)
	if !cl.verifyPasswordKey(oldPasswordKey) {
		fmt.Println("error: password verification failed")
		return false
	}
	if err := cl.loadClient(oldPasswordKey); err != nil {
		fmt.Println("error: load wallet info failed")
		return false
	}

	// encrypt master key with new password
	newPasswordKey := doubleHash(newPassword)
	newMasterKey, err := aes.AesEncrypt(cl.masterKey, newPasswordKey, cl.iv)
	if err != nil {
		fmt.Println("error: set new password failed")
		return false
	}

	// update wallet file
	newPasswordHash := sha256.Sum256(newPasswordKey)
	if err := cl.SaveStoredData("PasswordHash", newPasswordHash[:]); err != nil {
		fmt.Println("error: wallet update failed(password hash)")
		return false
	}
	if err := cl.SaveStoredData("MasterKey", newMasterKey); err != nil {
		fmt.Println("error: wallet update failed (encrypted master key)")
		return false
	}
	clearBytes(newPasswordKey, len(newPasswordKey))
	clearBytes(cl.masterKey, len(cl.masterKey))

	return true
}

func (cl *ClientImpl) ContainsAccount(pubKey keypair.PublicKey) bool {
	addr := types.AddressFromPubKey(pubKey)
	return cl.GetAccountByAddress(addr) != nil
}

func (cl *ClientImpl) CreateAccount(encrypt string) (*Account, error) {
	ac := NewAccount(encrypt)

	cl.mu.Lock()
	cl.accounts[ac.Address] = ac
	cl.mu.Unlock()

	err := cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}

	return ac, nil
}

func (cl *ClientImpl) CreateAccountByPrivateKey(privateKey []byte) (*Account, error) {
	ac, err := NewAccountWithPrivatekey(privateKey)
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if err != nil {
		return nil, err
	}

	cl.accounts[ac.Address] = ac
	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

func (cl *ClientImpl) verifyPasswordKey(passwordKey []byte) bool {
	savedPasswordHash, err := cl.LoadStoredData("PasswordHash")
	if err != nil {
		fmt.Println("error: failed to load password hash")
		return false
	}
	if savedPasswordHash == nil {
		fmt.Println("error: saved password hash is nil")
		return false
	}
	passwordHash := sha256.Sum256(passwordKey)
	///clearBytes(passwordKey, len(passwordKey))
	if !bytes.Equal(savedPasswordHash, passwordHash[:]) {
		fmt.Println("error: password wrong")
		return false
	}
	return true
}

func (cl *ClientImpl) EncryptPrivateKey(prikey []byte) ([]byte, error) {
	enc, err := aes.AesEncrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return enc, nil
}

func (cl *ClientImpl) DecryptPrivateKey(prikey []byte) ([]byte, error) {
	if prikey == nil {
		return nil, ontErrors.NewDetailErr(errors.New("The PriKey is nil"), ontErrors.ErrNoCode, "")
	}

	dec, err := aes.AesDecrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

func (cl *ClientImpl) SaveAccount(ac *Account) error {
	buf := keypair.SerializePrivateKey(ac.PrivateKey)
	encryptedPrivateKey, err := cl.EncryptPrivateKey(buf)
	if err != nil {
		return err
	}

	clearBytes(buf, len(buf))
	encryptedPrivateKey = append(encryptedPrivateKey, byte(ac.SigScheme))

	err = cl.SaveAccountData(ac.Address[:], encryptedPrivateKey)
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientImpl) LoadAccount() map[common.Address]*Account {
	i := 0
	accounts := map[common.Address]*Account{}
	for true {
		_, prikeyenc, err := cl.LoadAccountData(i)
		if err != nil {
			log.Error(err)
			break
		}

		length := len(prikeyenc)
		scheme := prikeyenc[length-1]
		prikeyenc = prikeyenc[:length-1]
		buf, err := cl.DecryptPrivateKey(prikeyenc)
		if err != nil {
			log.Error(err)
			break
		}

		ac, err := NewAccountWithPrivatekey(buf)
		if err != nil {
			log.Error(err)
			break
		}
		ac.SigScheme = s.SignatureScheme(scheme)
		accounts[ac.Address] = ac
		i++
		break
	}

	return accounts
}

func (cl *ClientImpl) GetBookkeepers() ([]keypair.PublicKey, error) {
	var pubKeys = []keypair.PublicKey{}
	consensusType := config.Parameters.ConsensusType
	if consensusType == "solo" {
		ac := cl.GetDefaultAccount()
		if ac == nil {
			return nil, fmt.Errorf("GetDefaultAccount error")
		}
		pubKeys = append(pubKeys, ac.PublicKey)
		return pubKeys, nil
	}

	sort.Strings(config.Parameters.Bookkeepers)
	for _, key := range config.Parameters.Bookkeepers {
		//pubKey := []byte(key)
		pubKey, err := hex.DecodeString(key)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			log.Error("Incorrectly book keepers key")
			return nil, fmt.Errorf("Incorrectly book keepers key:%s", key)
		}
		pubKeys = append(pubKeys, k)
	}

	return pubKeys, nil
}

func clientIsDefaultBookkeeper(publicKey string) bool {
	for _, bookkeeper := range config.Parameters.Bookkeepers {
		if strings.Compare(bookkeeper, publicKey) == 0 {
			return true
		}
	}
	return false
}

func nodeType(typeName string) int {
	if "service" == config.Parameters.NodeType {
		return protocol.SERVICE_NODE
	} else {
		return protocol.VERIFY_NODE
	}
}

func GetClient() Client {
	if !common.FileExisted(WALLET_FILENAME) {
		log.Fatal(fmt.Sprintf("No %s detected, please create a wallet by using command line.", WALLET_FILENAME))
		os.Exit(1)
	}
	passwd, err := password.GetAccountPassword()
	if err != nil {
		log.Fatal("Get password error.")
		os.Exit(1)
	}
	c := Open(WALLET_FILENAME, passwd)
	if c == nil {
		return nil
	}
	return c
}

func GetBookkeepers() []keypair.PublicKey {
	var pubKeys = []keypair.PublicKey{}
	sort.Strings(config.Parameters.Bookkeepers)
	for _, key := range config.Parameters.Bookkeepers {
		pubKey := []byte(key)
		pubKey, err := hex.DecodeString(key)
		// TODO Convert the key string to byte
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			log.Error("Incorrectly book keepers key")
			return nil
		}
		pubKeys = append(pubKeys, k)
	}

	return pubKeys
}

func doubleHash(pwd []byte) []byte {
	pwdhash := sha256.Sum256(pwd)
	pwdhash2 := sha256.Sum256(pwdhash[:])

	// Fixme clean the password buffer
	// clearBytes(pwd,len(pwd))
	clearBytes(pwdhash[:], 32)

	return pwdhash2[:]
}

func clearBytes(arr []byte, len int) {
	for i := 0; i < len; i++ {
		arr[i] = 0
	}
}
