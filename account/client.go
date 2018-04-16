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
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/ontio/ontology-crypto/aes"
	"github.com/ontio/ontology-crypto/keypair"
	s "github.com/ontio/ontology-crypto/signature"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
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

	WalletData
	isrunning bool
}

//TODO need redesign
func Create(path string, encrypt string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, true)
	return cl
}

func Open(path string, passwordKey []byte) *ClientImpl {
	if "" == path {
		path = WALLET_FILENAME
	}
	if !common.FileExisted(path) {
		log.Error(fmt.Sprintf("No %s detected, please create a wallet first.", path))
		return nil
	}

	defer clearBytes(passwordKey, len(passwordKey))

	if len(passwordKey) == 0 {
		fmt.Printf("Please enter password:")
		passwd, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Error("Get password error.")
			return nil
		}
		passwordKey = passwd
	}

	cl := NewClient(path, passwordKey, false)
	if cl == nil {
		log.Error("Alloc new client failure")
		return nil
	}

	return cl
}

//TODO need redesign
func NewClient(path string, password []byte, create bool) *ClientImpl {
	defer clearBytes(password, len(password))
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[common.Address]*Account{},
		isrunning: true,
	}

	passwordKey := password
	if !create {
		if err := newClient.loadClient(passwordKey); err != nil {
			fmt.Println(err)
			return nil
		}
	}
	return newClient
}

func (cl *ClientImpl) loadClient(passwordKey []byte) error {
	var err error
	err = cl.Load(cl.path)
	if err != nil {
		fmt.Println("error: failed load wallet")
		return err
	}

	for i, v := range cl.Accounts {
		if !v.VerifyPassword(passwordKey) {
			fmt.Println("error: incorrect password for account", i)
			continue
		}
		ac := new(Account)
		ac.PrivateKey, err = keypair.DecryptPrivateKey(&v.ProtectedKey, passwordKey)
		if err != nil {
			fmt.Println("error: failed load account", i+1)
			continue
		}

		ac.PublicKey = ac.PrivateKey.Public()
		ac.Address = types.AddressFromPubKey(ac.PublicKey)
		if ac.Address.ToBase58() != v.Address {
			fmt.Println("warning: incorrect address of account", i)
		}

		scheme, err := s.GetScheme(v.SigSch)
		if err != nil {
			fmt.Println("error: invalid signature scheme of account", i)
			continue
		}
		ac.SigScheme = scheme
		cl.accounts[ac.Address] = ac
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

func (cl *ClientImpl) GetDefaultAccount() *Account {
	ac := cl.WalletData.GetDefaultAccount()
	if ac == nil {
		return nil
	}
	addr := ac.Address
	for _, v := range cl.accounts {
		if v.Address.ToBase58() == addr {
			return v
		}
	}
	return nil
}

func (cl *ClientImpl) ContainsAccount(pubKey keypair.PublicKey) bool {
	addr := types.AddressFromPubKey(pubKey)
	return cl.GetAccountByAddress(addr) != nil
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

	clearBytes(pwdhash[:], 32)

	return pwdhash2[:]
}

func clearBytes(arr []byte, len int) {
	for i := 0; i < len; i++ {
		arr[i] = 0
	}
}
