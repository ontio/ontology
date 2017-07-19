package account

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	. "DNA/common"
	"DNA/common/config"
	"DNA/common/log"
	"DNA/common/password"
	"DNA/core/contract"
	ct "DNA/core/contract"
	"DNA/core/ledger"
	sig "DNA/core/signature"
	"DNA/crypto"
	. "DNA/errors"
	"DNA/net/protocol"
)

const (
	DefaultBookKeeperCount = 4
	WalletFileName         = "wallet.dat"
)

type Client interface {
	Sign(context *ct.ContractContext) bool
	ContainsAccount(pubKey *crypto.PubKey) bool
	GetAccount(pubKey *crypto.PubKey) (*Account, error)
	GetDefaultAccount() (*Account, error)
}

type ClientImpl struct {
	mu sync.Mutex

	path      string
	iv        []byte
	masterKey []byte

	accounts  map[Uint160]*Account
	contracts map[Uint160]*ct.Contract

	watchOnly     []Uint160
	currentHeight uint32

	FileStore
	isrunning bool
}

//TODO: adjust contract folder structure
func Create(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, true)

	_, err := cl.CreateAccount()
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
	cl.contracts = cl.LoadContracts()
	if cl.contracts == nil {
		log.Error("Load contracts failure")
	}
	return cl
}

func NewClient(path string, password []byte, create bool) *ClientImpl {
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[Uint160]*Account{},
		contracts: map[Uint160]*ct.Contract{},
		FileStore: FileStore{path: path},
		isrunning: true,
	}

	passwordKey := crypto.ToAesKey(password)
	if create {
		//create new client
		newClient.iv = make([]byte, 16)
		newClient.masterKey = make([]byte, 32)
		newClient.watchOnly = []Uint160{}
		newClient.currentHeight = 0

		//generate random number for iv/masterkey
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := 0; i < 16; i++ {
			newClient.iv[i] = byte(r.Intn(256))
		}
		for i := 0; i < 32; i++ {
			newClient.masterKey[i] = byte(r.Intn(256))
		}

		//new client store (build DB)
		newClient.BuildDatabase(path)

		// SaveStoredData
		pwdhash := sha256.Sum256(passwordKey)
		err := newClient.SaveStoredData("PasswordHash", pwdhash[:])
		if err != nil {
			log.Error(err)
			return nil
		}
		err = newClient.SaveStoredData("IV", newClient.iv[:])
		if err != nil {
			log.Error(err)
			return nil
		}

		aesmk, err := crypto.AesEncrypt(newClient.masterKey[:], passwordKey, newClient.iv)
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
	ClearBytes(passwordKey, len(passwordKey))
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
	cl.masterKey, err = crypto.AesDecrypt(encryptedMasterKey, passwordKey, cl.iv)
	if err != nil {
		fmt.Println("error: failed to decrypt master key")
		return err
	}
	return nil
}

func (cl *ClientImpl) GetDefaultAccount() (*Account, error) {
	for programHash, _ := range cl.accounts {
		return cl.GetAccountByProgramHash(programHash), nil
	}

	return nil, NewDetailErr(errors.New("Can't load default account."), ErrNoCode, "")
}

func (cl *ClientImpl) GetAccount(pubKey *crypto.PubKey) (*Account, error) {
	signatureRedeemScript, err := contract.CreateSignatureRedeemScript(pubKey)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "CreateSignatureRedeemScript failed")
	}
	programHash, err := ToCodeHash(signatureRedeemScript)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "ToCodeHash failed")
	}
	return cl.GetAccountByProgramHash(programHash), nil
}

func (cl *ClientImpl) GetAccountByProgramHash(programHash Uint160) *Account {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if account, ok := cl.accounts[programHash]; ok {
		return account
	}
	return nil
}

func (cl *ClientImpl) GetContract(programHash Uint160) *ct.Contract {
	log.Debug()
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if contract, ok := cl.contracts[programHash]; ok {
		return contract
	}
	return nil
}

func (cl *ClientImpl) ChangePassword(oldPassword []byte, newPassword []byte) bool {
	// check password
	oldPasswordKey := crypto.ToAesKey(oldPassword)
	if !cl.verifyPasswordKey(oldPasswordKey) {
		fmt.Println("error: password verification failed")
		return false
	}
	if err := cl.loadClient(oldPasswordKey); err != nil {
		fmt.Println("error: load wallet info failed")
		return false
	}

	// encrypt master key with new password
	newPasswordKey := crypto.ToAesKey(newPassword)
	newMasterKey, err := crypto.AesEncrypt(cl.masterKey, newPasswordKey, cl.iv)
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
	ClearBytes(newPasswordKey, len(newPasswordKey))
	ClearBytes(cl.masterKey, len(cl.masterKey))

	return true
}

func (cl *ClientImpl) ContainsAccount(pubKey *crypto.PubKey) bool {
	signatureRedeemScript, err := contract.CreateSignatureRedeemScript(pubKey)
	if err != nil {
		return false
	}
	programHash, err := ToCodeHash(signatureRedeemScript)
	if err != nil {
		return false
	}
	if cl.GetAccountByProgramHash(programHash) != nil {
		return true
	} else {
		return false
	}
}

func (cl *ClientImpl) CreateAccount() (*Account, error) {
	ac, err := NewAccount()
	if err != nil {
		return nil, err
	}

	cl.mu.Lock()
	cl.accounts[ac.ProgramHash] = ac
	cl.mu.Unlock()

	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}

	ct, err := contract.CreateSignatureContract(ac.PublicKey)
	if err == nil {
		cl.AddContract(ct)
		address, _ := ct.ProgramHash.ToAddress()
		log.Info("[CreateContract] Address: ", address)
	}

	log.Info("Create account Success")
	return ac, nil
}

func (cl *ClientImpl) CreateAccountByPrivateKey(privateKey []byte) (*Account, error) {
	ac, err := NewAccountWithPrivatekey(privateKey)
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if err != nil {
		return nil, err
	}

	cl.accounts[ac.ProgramHash] = ac
	err = cl.SaveAccount(ac)
	if err != nil {
		return nil, err
	}
	return ac, nil
}

func (cl *ClientImpl) ProcessBlocks() {
	for {
		if !cl.isrunning {
			break
		}

		for {
			if ledger.DefaultLedger.Blockchain == nil {
				break
			}
			if cl.currentHeight > ledger.DefaultLedger.Blockchain.BlockHeight {
				break
			}

			cl.mu.Lock()

			block, _ := ledger.DefaultLedger.GetBlockWithHeight(cl.currentHeight)
			if block != nil {
				cl.ProcessNewBlock(block)
			}

			cl.mu.Unlock()
		}

		for i := 0; i < 20; i++ {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (cl *ClientImpl) ProcessNewBlock(block *ledger.Block) {
	//TODO: ProcessNewBlock

}

func (cl *ClientImpl) Sign(context *ct.ContractContext) bool {
	log.Debug()
	fSuccess := false
	for i, hash := range context.ProgramHashes {
		contract := cl.GetContract(hash)
		if contract == nil {
			continue
		}
		account := cl.GetAccountByProgramHash(hash)
		if account == nil {
			continue
		}

		signature, err := sig.SignBySigner(context.Data, account)
		if err != nil {
			return fSuccess
		}
		err = context.AddContract(contract, account.PublicKey, signature)

		if err != nil {
			fSuccess = false
		} else {
			if i == 0 {
				fSuccess = true
			}
		}
	}
	return fSuccess
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
	///ClearBytes(passwordKey, len(passwordKey))
	if !IsEqualBytes(savedPasswordHash, passwordHash[:]) {
		fmt.Println("error: password wrong")
		return false
	}
	return true
}

func (cl *ClientImpl) EncryptPrivateKey(prikey []byte) ([]byte, error) {
	enc, err := crypto.AesEncrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return enc, nil
}

func (cl *ClientImpl) DecryptPrivateKey(prikey []byte) ([]byte, error) {
	if prikey == nil {
		return nil, NewDetailErr(errors.New("The PriKey is nil"), ErrNoCode, "")
	}
	if len(prikey) != 96 {
		return nil, NewDetailErr(errors.New("The len of PriKeyEnc is not 96bytes"), ErrNoCode, "")
	}

	dec, err := crypto.AesDecrypt(prikey, cl.masterKey, cl.iv)
	if err != nil {
		return nil, err
	}

	return dec, nil
}

func (cl *ClientImpl) SaveAccount(ac *Account) error {
	decryptedPrivateKey := make([]byte, 96)
	temp, err := ac.PublicKey.EncodePoint(false)
	if err != nil {
		return err
	}
	for i := 1; i <= 64; i++ {
		decryptedPrivateKey[i-1] = temp[i]
	}

	for i := len(ac.PrivateKey) - 1; i >= 0; i-- {
		decryptedPrivateKey[96+i-len(ac.PrivateKey)] = ac.PrivateKey[i]
	}

	encryptedPrivateKey, err := cl.EncryptPrivateKey(decryptedPrivateKey)
	if err != nil {
		return err
	}

	ClearBytes(decryptedPrivateKey, 96)

	err = cl.SaveAccountData(ac.ProgramHash.ToArray(), encryptedPrivateKey)
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientImpl) LoadAccount() map[Uint160]*Account {
	i := 0
	accounts := map[Uint160]*Account{}
	for true {
		_, prikeyenc, err := cl.LoadAccountData(i)
		if err != nil {
			// TODO: report the error
		}

		decryptedPrivateKey, err := cl.DecryptPrivateKey(prikeyenc)
		if err != nil {
			log.Error(err)
		}

		prikey := decryptedPrivateKey[64:96]
		ac, err := NewAccountWithPrivatekey(prikey)
		accounts[ac.ProgramHash] = ac
		i++
		break
	}

	return accounts
}

func (cl *ClientImpl) LoadContracts() map[Uint160]*ct.Contract {
	i := 0
	contracts := map[Uint160]*ct.Contract{}

	for true {
		ph, _, rd, err := cl.LoadContractData(i)
		if err != nil {
			fmt.Println(err)
			break
		}

		rdreader := bytes.NewReader(rd)
		ct := new(ct.Contract)
		ct.Deserialize(rdreader)

		programhash, err := Uint160ParseFromBytes(ph)
		ct.ProgramHash = programhash
		contracts[ct.ProgramHash] = ct
		i++
		break
	}

	return contracts
}

func (cl *ClientImpl) AddContract(ct *contract.Contract) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.accounts[ct.ProgramHash] == nil {
		return NewDetailErr(errors.New("AddContract(): contract.OwnerPubkeyHash not in []accounts"), ErrNoCode, "")
	}

	cl.contracts[ct.ProgramHash] = ct

	err := cl.SaveContractData(ct)
	return err
}

func clientIsDefaultBookKeeper(publicKey string) bool {
	for _, bookKeeper := range config.Parameters.BookKeepers {
		if strings.Compare(bookKeeper, publicKey) == 0 {
			return true
		}
	}
	return false
}

func nodeType(typeName string) int {
	if "service" == config.Parameters.NodeType {
		return protocol.SERVICENODE
	} else {
		return protocol.VERIFYNODE
	}
}

func GetClient() Client {
	if !FileExisted(WalletFileName) {
		log.Fatal(fmt.Sprintf("No %s detected, please create a wallet by using command line.", WalletFileName))
		os.Exit(1)
	}
	passwd, err := password.GetAccountPassword()
	if err != nil {
		log.Fatal("Get password error.")
		os.Exit(1)
	}
	c := Open(WalletFileName, passwd)
	if c == nil {
		return nil
	}
	return c
}

func GetBookKeepers() []*crypto.PubKey {
	var pubKeys = []*crypto.PubKey{}
	sort.Strings(config.Parameters.BookKeepers)
	for _, key := range config.Parameters.BookKeepers {
		pubKey := []byte(key)
		pubKey, err := hex.DecodeString(key)
		// TODO Convert the key string to byte
		k, err := crypto.DecodePoint(pubKey)
		if err != nil {
			log.Error("Incorrectly book keepers key")
			return nil
		}
		pubKeys = append(pubKeys, k)
	}

	return pubKeys
}
