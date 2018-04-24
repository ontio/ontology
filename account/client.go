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
	"github.com/ontio/ontology/net/protocol"
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
		fmt.Println("")
		passwordKey = passwd
	}

	cl := NewClient(path, passwordKey, false)
	if cl == nil {
		log.Error("Alloc new client failure")
		return nil
	}

	return cl
}

func NewClient(path string, password []byte, create bool) *ClientImpl {
	defer clearBytes(password, len(password))
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[common.Address]*Account{},
		isrunning: true,
	}

	passwordKey := password
	if create { /*
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
			}*/
	} else {
		//if b := newClient.verifyPasswordKey(passwordKey); b == false {
		//	return nil
		//}
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
	addr := cl.WalletData.GetDefaultAccount().Address
	for _, v := range cl.accounts {
		if v.Address.ToBase58() == addr {
			return v
		}
	}
	return nil
}

// Deprecated
func (cl *ClientImpl) ChangePassword(oldPassword []byte, newPassword []byte) bool {

	/*
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
		defer clearBytes(newPasswordKey, len(newPasswordKey))
		defer clearBytes(cl.masterKey, len(cl.masterKey))
	*/

	return true
}

func (cl *ClientImpl) ContainsAccount(pubKey keypair.PublicKey) bool {
	addr := types.AddressFromPubKey(pubKey)
	return cl.GetAccountByAddress(addr) != nil
}

// Deprecated
func (cl *ClientImpl) CreateAccount(encrypt string) (*Account, error) {
	ac := NewAccount(encrypt)

	cl.mu.Lock()
	cl.accounts[ac.Address] = ac
	cl.mu.Unlock()

	//err := cl.SaveAccount(ac)
	//if err != nil {
	//	return nil, err
	//}

	return ac, nil
}

/*
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
}*/

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

/*
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
}*/

/*
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
*/

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
