package client

import (
	"GoOnchain/crypto"
	ct "GoOnchain/core/contract"
	. "GoOnchain/common"
	"sync"
	sig "GoOnchain/core/signature"
	. "GoOnchain/errors"
	"errors"
	"GoOnchain/core/ledger"
	"time"
	"fmt"
	"crypto/sha256"
	"math/rand"
	"bytes"
	"GoOnchain/common/serialization"
	"GoOnchain/core/contract"
)

type ClientVersion struct {
	Major uint32
	Minor uint32
	Build uint32
	Revision uint32
}

func (v *ClientVersion) ToArray() []byte {
	vbuf := bytes.NewBuffer(nil)
	serialization.WriteUint32( vbuf, v.Major )
	serialization.WriteUint32( vbuf, v.Minor )
	serialization.WriteUint32( vbuf, v.Build )
	serialization.WriteUint32( vbuf, v.Revision )

	//fmt.Printf("ToArray: %x\n", vbuf.Bytes())

	return vbuf.Bytes()
}

type Client struct {
	mu		sync.Mutex

	path 		string
	iv 		[]byte
	masterKey 	[]byte

	accounts 	map[Uint160]*Account
	contracts 	map[Uint160]*ct.Contract

	watchOnly 	[]Uint160
	currentHeight 	uint32

	store 		ClientStore
	isrunning 	bool
}

//TODO: adjust contract folder structure

func CreateClient( path string, passwordKey []byte ) *Client {
	cl := NewClient( path, passwordKey, true )

	_,err := cl.CreateAccount()
	if err != nil {
		fmt.Println( err )
	}

	return cl
}

func OpenClient( path string, passwordKey []byte ) *Client {
	cl := NewClient( path, passwordKey, false )

	cl.accounts = cl.LoadAccount()
	cl.contracts = cl.LoadContracts()

	return cl
}

func NewClient( path string, passwordKey []byte, create bool ) *Client {

	newClient := &Client{
		path: path,
		accounts:map[Uint160]*Account{},
		contracts:map[Uint160]*ct.Contract{},
		store: ClientStore{path: path},
		isrunning: true,
	}

	// passwordkey to AESKey first
	passwordKey = crypto.ToAesKey(passwordKey)

	if create {
		//create new client
		newClient.iv = make([]byte,16)
		newClient.masterKey = make([]byte,32)
		newClient.watchOnly = []Uint160{}
		newClient.currentHeight = 0

		//generate random number for iv/masterkey
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i:=0; i<16; i++ {
			newClient.iv[i] = byte(r.Intn(256))
		}
		for i:=0; i<32; i++ {
			newClient.masterKey[i] = byte(r.Intn(256))
		}

		//new client store (build DB)
		newClient.store.BuildDatabase( path )

		// SaveStoredData
		pwdhash := sha256.Sum256(passwordKey)
		err := newClient.store.SaveStoredData("PasswordHash",pwdhash[:])
		if err != nil {
			fmt.Println( err )
			return nil
		}
		err = newClient.store.SaveStoredData("IV",newClient.iv[:])
		if err != nil {
			fmt.Println( err )
			return nil
		}

		aesmk,err := crypto.AesEncrypt( newClient.masterKey[:], passwordKey, newClient.iv  )
		if err == nil {
			err = newClient.store.SaveStoredData("MasterKey",aesmk)
			if err != nil {
				fmt.Println( err )
				return nil
			}
		} else {
			fmt.Println( err )
			return nil
		}

		v := ClientVersion{0,0,0,1}
		err = newClient.store.SaveStoredData("ClientVersion",v.ToArray())
		if err != nil {
			fmt.Println( err )
			return nil
		}
		err = newClient.store.SaveStoredData("Height",IntToBytes(int(newClient.currentHeight)))
		if err != nil {
			fmt.Println( err )
			return nil
		}

	} else {

		//load client
		passwordHash,err := newClient.store.LoadStoredData("PasswordHash")
		if err != nil {
			fmt.Println( err )
			return nil
		}

		pwdhash := sha256.Sum256(passwordKey)
		if passwordHash != nil && !IsEqualBytes(passwordHash,pwdhash[:]){
			//TODO: add panic
			fmt.Println( "passwordHash = nil or password wrong!" )
			return nil
		}

		fmt.Println( "[OpenClient] Password Verify." )

		newClient.iv,err = newClient.store.LoadStoredData("IV")
		if err != nil {
			fmt.Println( err )
			return nil
		}

		masterKey, err := newClient.store.LoadStoredData("MasterKey")
		if err != nil {
			fmt.Println( err )
			return nil
		}

		newClient.masterKey,err = crypto.AesDecrypt( masterKey, passwordKey, newClient.iv )
		if err != nil {
			fmt.Println( err )
			return nil
		}

		/*
		newClient.accounts = newClient.LoadAccount()
		newClient.contracts = newClient.LoadContracts()

		//TODO: watch only
		go newClient.ProcessBlocks()
		*/

	}

	ClearBytes(passwordKey,len(passwordKey))

	return newClient
}

func (cl *Client) GetDefaultAccount() (*Account,error){
	for k, _ := range cl.accounts {
		return cl.GetAccountByKeyHash(k),nil
	}

	return nil,NewDetailErr(errors.New("Can't load default account."), ErrNoCode, "")
}

func (cl *Client) GetAccount(pubKey *crypto.PubKey) (*Account,error){
	temp,err := pubKey.EncodePoint(true)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	hash,err :=ToCodeHash(temp)
	if err !=nil{
		return nil,NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	return cl.GetAccountByKeyHash(hash),nil
}

func (cl *Client) GetAccountByKeyHash(publicKeyHash Uint160) *Account{
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if account,ok := cl.accounts[publicKeyHash]; ok{
		return account
	}
	return nil
}

func (cl *Client) GetAccountByProgramHash(programHash Uint160) *Account{
	Trace()
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if contract,ok := cl.contracts[programHash]; ok{
		return cl.accounts[contract.OwnerPubkeyHash]
	}
	return nil
}

func (cl *Client) GetContract(codeHash Uint160) *ct.Contract{
	Trace()
	cl.mu.Lock()
	defer cl.mu.Unlock()
	fmt.Println("codeHash",codeHash)
	    for _, v := range cl.contracts {
	            fmt.Println("cl.contracts = ",v)
	        }

	if contract,ok := cl.contracts[codeHash]; ok{
		return contract
	}
	fmt.Println("contract",cl.contracts[codeHash])
	return nil
}

func (cl *Client) ChangePassword(oldPassword string,newPassword string) bool{
	if !cl.VerifyPassword(oldPassword) {
		return  false
	}

	//TODO: ChangePassword

	return false
}

func (cl *Client) ContainsAccount(pubKey *crypto.PubKey) bool{

	acpubkey,err := pubKey.EncodePoint(true)
	if err == nil {
		Publickey, err := ToCodeHash( acpubkey )
		if err == nil {
			if cl.GetAccountByKeyHash(Publickey) != nil {
				return true
			} else {
				return false
			}

		} else {
			fmt.Println( err )
			return false
		}
	} else {
		fmt.Println( err )
		return false
	}
}

func (cl *Client) CreateAccount() (*Account,error){
	ac,err := NewAccount()

	if err == nil {
		cl.mu.Lock()
		cl.accounts[ac.PublicKeyHash] = ac
		cl.mu.Unlock()

		err := cl.SaveAccount(ac)
		if err != nil {
			return nil,err
		}

		//fmt.Printf("[CreateAccount] PrivateKey: %x\n", ac.PrivateKey)
		//fmt.Printf("[CreateAccount] PublicKeyHash: %x\n", ac.PublicKeyHash)
		//fmt.Printf("[CreateAccount] PublicKeyAddress: %s\n", ac.PublicKeyHash.ToAddress())

		//cl.AddContract( contract.CreateSignatureContract( ac.PublicKey ) )
		ct,err := contract.CreateSignatureContract( ac.PublicKey )
		if err == nil {
			cl.AddContract( ct )
		}

		return ac,nil
	} else {
		return nil,err
	}
}

func (cl *Client) CreateAccountByPrivateKey(privateKey []byte) (*Account, error) {
	ac,err := NewAccountWithPrivatekey(privateKey)
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if err == nil {
		cl.accounts[ac.PublicKeyHash] = ac
		err := cl.SaveAccount(ac)
		if err != nil {
			return nil,err
		}

		//fmt.Printf("[CreateAccountByPrivateKey] PrivateKey: %x\n", ac.PrivateKey)
		//fmt.Printf("[CreateAccountByPrivateKey] PublicKeyHash: %x\n", ac.PublicKeyHash)
		//fmt.Printf("[CreateAccountByPrivateKey] PublicKeyAddress: %s\n", ac.PublicKeyHash.ToAddress())

		return ac,nil
	} else {
		return nil,err
	}
}

func (cl *Client) ProcessBlocks() {
	for  {
		if !cl.isrunning { break}

		for{
			if ledger.DefaultLedger.Blockchain == nil {break}
			if cl.currentHeight > ledger.DefaultLedger.Blockchain.BlockHeight {break}

			cl.mu.Lock()

			block ,_:= ledger.DefaultLedger.GetBlockWithHeight(cl.currentHeight)
			if block != nil{
				cl.ProcessNewBlock(block)
			}

			cl.mu.Unlock()
		}

		for i:=0;i < 20 ;i++ {
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (cl *Client) ProcessNewBlock(block *ledger.Block) {
	//TODO: ProcessNewBlock

}

func (cl *Client) Sign(context *ct.ContractContext) bool{
	Trace()
	fSuccess := false
	for i,hash := range context.ProgramHashes{
		fmt.Println("Sign hash=",hash)
		contract := cl.GetContract(hash)
		if contract == nil {continue}
		fmt.Println("cl.GetContract(hash)=",cl.GetContract(hash))
		account := cl.GetAccountByProgramHash(hash)
		fmt.Println("account",account)
		if account == nil {continue}

		signature,err:= sig.SignBySigner(context.Data,account)
		if err != nil {
			return fSuccess
		}
		err = context.AddContract(contract,account.PublicKey,signature)

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

func (cl *Client) VerifyPassword(password string) bool{
	//TODO: VerifyPassword
	return true
}

func (cl *Client) EncryptPrivateKey( prikey []byte) ([]byte,error) {
	enc,err := crypto.AesEncrypt( prikey, cl.masterKey, cl.iv  )
	if err != nil {
		return nil,err
	}

	return enc,nil
}

func (cl *Client) DecryptPrivateKey( prikey []byte) ([]byte,error) {
	if prikey == nil {
		return nil, NewDetailErr(errors.New("The PriKey is nil"), ErrNoCode, "")
	}
	if len(prikey) != 96  {
		return nil, NewDetailErr(errors.New("The len of PriKeyEnc is not 96bytes"), ErrNoCode, "")
	}

	dec,err := crypto.AesDecrypt( prikey, cl.masterKey, cl.iv  )
	if err != nil {
		return nil,err
	}

	return dec,nil
}

func (cl *Client) SaveAccount(ac *Account) error {

	decryptedPrivateKey := make([]byte, 96)
	temp,err := ac.PublicKey.EncodePoint(false)
	if err != nil {
		return  err
	}
	for i:=1; i<=64; i++ {
		decryptedPrivateKey[i-1] = temp[i]
	}

	for i:=0; i<32; i++ {
		decryptedPrivateKey[64+i] = ac.PrivateKey[i]
	}

	//fmt.Printf("decryptedPrivateKey: %x\n", decryptedPrivateKey)
	encryptedPrivateKey,err := cl.EncryptPrivateKey(decryptedPrivateKey)
	if err != nil {
		return err
	}

	ClearBytes(decryptedPrivateKey, 96)

	err = cl.store.SaveAccountData( ac.PublicKeyHash.ToArray(), encryptedPrivateKey )
	if err != nil {
		return err
	}

	return nil
}

func (cl *Client) LoadAccount()  map[Uint160]*Account {

	i := 0
	accounts := map[Uint160]*Account{}
	for true {
		pubkeyhash,prikeyenc,err := cl.store.LoadAccountData(i)
		if err != nil {
			//fmt.Println( err )
			//return nil
			break
		}

		decryptedPrivateKey,err := cl.DecryptPrivateKey(prikeyenc)
		if err != nil {
			fmt.Println( err )
		}
		//fmt.Printf("decryptedPrivateKey: %x\n", decryptedPrivateKey)

		prikey := decryptedPrivateKey[64:96]
		ac,err := NewAccountWithPrivatekey(prikey)

		//ClearBytes( decryptedPrivateKey, 96 )
		//ClearBytes( prikey, 32 )

		//fmt.Printf("[LoadAccount] PrivateKey: %x\n", ac.PrivateKey)
		//fmt.Printf("[LoadAccount] PublicKeyHash: %x\n", ac.PublicKeyHash.ToArray())
		//fmt.Printf("[LoadAccount] PublicKeyAddress: %s\n", ac.PublicKeyHash.ToAddress())

		pkhash,_ := Uint160ParseFromBytes(pubkeyhash)
		accounts[pkhash] = ac

		i ++
	}

	return accounts
}

func (cl *Client) LoadContracts()  map[Uint160]*ct.Contract{

	i := 0
	contracts := map[Uint160]*ct.Contract{}

	for true {
		ph,_,rd,err := cl.store.LoadContractData(i)
		if err != nil {
			//fmt.Println( err )
			break
		}

		rdreader := bytes.NewReader(rd)
		ct := new(ct.Contract)
		ct.Deserialize(rdreader)

		programhash,err := Uint160ParseFromBytes(ph)
		ct.ProgramHash = programhash

		contracts[ct.ProgramHash] = ct

		//fmt.Printf("[LoadContracts] ScriptHash: %x\n", ct.ProgramHash)
		//fmt.Printf("[LoadContracts] PublicKeyHash: %x\n", ct.OwnerPubkeyHash.ToArray())
		//fmt.Printf("[LoadContracts] Code: %x\n", ct.Code)
		//fmt.Printf("[LoadContracts] Parameters: %x\n", ct.Parameters)

		i ++
	}

	return contracts
}
func (cl *Client) AddContract(ct * contract.Contract) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if cl.accounts[ct.OwnerPubkeyHash] != nil {
		cl.contracts[ct.ProgramHash] = ct;
		// TODO; watchonly

	} else {
		return NewDetailErr(errors.New("AddContract(): contract.OwnerPubkeyHash not in []accounts"), ErrNoCode, "")
	}

	err := cl.store.SaveContractData(ct)
	if err != nil {
		return err
	}

	return nil
}
