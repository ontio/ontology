package client

import (
	. "DNA/common"
	"DNA/common/log"
	"DNA/common/serialization"
	"DNA/core/contract"
	ct "DNA/core/contract"
	"DNA/core/ledger"
	sig "DNA/core/signature"
	"DNA/crypto"
	. "DNA/errors"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ClientVersion struct {
	Major    uint32
	Minor    uint32
	Build    uint32
	Revision uint32
}

func (v *ClientVersion) ToArray() []byte {
	vbuf := bytes.NewBuffer(nil)
	serialization.WriteUint32(vbuf, v.Major)
	serialization.WriteUint32(vbuf, v.Minor)
	serialization.WriteUint32(vbuf, v.Build)
	serialization.WriteUint32(vbuf, v.Revision)

	return vbuf.Bytes()
}

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

	store     FileStore
	isrunning bool
}

//TODO: adjust contract folder structure

func CreateClient(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, true)

	_, err := cl.CreateAccount()
	if err != nil {
		fmt.Println(err)
	}

	return cl
}

func OpenClient(path string, passwordKey []byte) *ClientImpl {
	cl := NewClient(path, passwordKey, false)

	if cl != nil {
		cl.accounts = cl.LoadAccount()
		cl.contracts = cl.LoadContracts()

		return cl
	}

	return nil
}

func NewClient(path string, passwordKey []byte, create bool) *ClientImpl {
	newClient := &ClientImpl{
		path:      path,
		accounts:  map[Uint160]*Account{},
		contracts: map[Uint160]*ct.Contract{},
		store:     FileStore{path: path},
		isrunning: true,
	}

	passwordKey = crypto.ToAesKey(passwordKey)

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
		newClient.store.BuildDatabase(path)

		// SaveStoredData
		pwdhash := sha256.Sum256(passwordKey)
		err := newClient.store.SaveStoredData("PasswordHash", pwdhash[:])
		if err != nil {
			fmt.Println(err)
			return nil
		}
		err = newClient.store.SaveStoredData("IV", newClient.iv[:])
		if err != nil {
			fmt.Println(err)
			return nil
		}

		aesmk, err := crypto.AesEncrypt(newClient.masterKey[:], passwordKey, newClient.iv)
		if err == nil {
			err = newClient.store.SaveStoredData("MasterKey", aesmk)
			if err != nil {
				log.Error(err)
				return nil
			}
		} else {
			log.Error(err)
			return nil
		}
	} else {
		//load client
		passwordHash, err := newClient.store.LoadStoredData("PasswordHash")
		if err != nil {
			log.Error(err)
			return nil
		}

		pwdhash := sha256.Sum256(passwordKey)
		if passwordHash == nil {
			log.Error("ERROR: passwordHash = nil")
			return nil
		}

		if !IsEqualBytes(passwordHash, pwdhash[:]) {
			log.Error("ERROR: password wrong!")
			return nil
		}

		log.Info("[OpenClient] Password Verify.")

		newClient.iv, err = newClient.store.LoadStoredData("IV")
		if err != nil {
			log.Error(err)
			return nil
		}

		masterKey, err := newClient.store.LoadStoredData("MasterKey")
		if err != nil {
			log.Error(err)
			return nil
		}

		newClient.masterKey, err = crypto.AesDecrypt(masterKey, passwordKey, newClient.iv)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	ClearBytes(passwordKey, len(passwordKey))

	return newClient
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

func (cl *ClientImpl) ChangePassword(oldPassword string, newPassword string) bool {
	if !cl.VerifyPassword(oldPassword) {
		return false
	}

	//TODO: ChangePassword

	return false
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
		log.Info("[CreateContract] Address: %s\n", ct.ProgramHash.ToAddress())
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

func (cl *ClientImpl) VerifyPassword(password string) bool {
	//TODO: VerifyPassword
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

	err = cl.store.SaveAccountData(ac.ProgramHash.ToArray(), encryptedPrivateKey)
	if err != nil {
		return err
	}

	return nil
}

func (cl *ClientImpl) LoadAccount() map[Uint160]*Account {

	i := 0
	accounts := map[Uint160]*Account{}
	for true {
		_, prikeyenc, err := cl.store.LoadAccountData(i)
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
		ph, _, rd, err := cl.store.LoadContractData(i)
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

	err := cl.store.SaveContractData(ct)
	return err
}
