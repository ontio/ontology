package client

import (
	"GoOnchain/crypto"
	. "GoOnchain/common"
	_ "errors"
)

type Account struct {
	PrivateKey []byte
	PublicKey *crypto.PubKey
	PublicKeyHash Uint160
}

func NewAccount(privateKey []byte) (*Account, error){
	//privKeyLen := len(privateKey)
	//
	//if privKeyLen != 32 && privKeyLen != 96 && privKeyLen != 104 {
	//	return nil,errors.New("Invalid private Key.")
	//}
	//
	//priKey := make([]byte,32)
	//pubKey := &crypto.PubKey{}

	//TODO: copy private Key

	//TODO: set public key
	priKey,pubKey,_ := crypto.GenKeyPair()
	return &Account{
		PrivateKey: priKey,
		PublicKey: &pubKey,
		PublicKeyHash: ToCodeHash(pubKey.EncodePoint(true)),
	},nil
}

//get signer's private key
func (ac *Account) PrivKey() []byte{
	return ac.PrivateKey
}

//get signer's public key
func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}