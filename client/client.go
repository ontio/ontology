package client

import (
	"GoOnchain/crypto"
	ct "GoOnchain/core/contract"
	. "GoOnchain/common"
	"sync"
)


type Client struct {
	mu           sync.Mutex
	path string
	accounts map[Uint160]*Account
	contracts map[Uint160]*ct.Contract
}

//TODO: adjust contract folder structure

func (cl *Client) GetAccount(pubKey *crypto.PubKey) *Account{
	return cl.GetAccountByKeyHash(ToCodeHash(pubKey.EncodePoint(true)))
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
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if contract,ok := cl.contracts[programHash]; ok{
		return cl.accounts[contract.OwnerPubkeyHash]
	}
	return nil
}

func (cl *Client) GetContract(codeHash Uint160) *ct.Contract{
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if contract,ok := cl.contracts[codeHash]; ok{
		return contract
	}
	return nil
}

func (cl *Client) ContainsAccount(pubKey *crypto.PubKey) bool{
	//TODO: ContainsAccount
	return false
}

func (cl *Client) CreateAccount() *Account{
	privateKey := make([]byte,32)

	//TODO: Generate Private Key

	account := cl.CreateAccountByPrivateKey(privateKey)
	ClearBytes(privateKey)

	return account
}

func (cl *Client) CreateAccountByPrivateKey(privateKey []byte) *Account {
	account,_ := NewAccount(privateKey)
	cl.mu.Lock()
	defer cl.mu.Unlock()
	cl.accounts[account.PublicKeyHash] = account

	return account
}

func (cl *Client) Sign(context *ct.ContractContext) bool{
	fSuccess := false
	for i,hash := range context.ProgramHashes{
		contract := cl.GetContract(hash)
		if contract == nil {continue}

		account := cl.GetAccountByProgramHash(hash)
		if account == nil {continue}


		signature := []byte{}//TODO: sign by account

		err := context.AddContract(contract,account.PublicKey,signature)

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

func ClearBytes(bytes []byte){
	for i:=0; i<len(bytes) ;i++  {
		bytes[i] = 0
	}
}