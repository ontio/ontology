package client

import (
	. "DNA/common"
	"DNA/crypto"
	. "DNA/errors"
	"errors"
)

type Account struct {
	PrivateKey    []byte
	PublicKey     *crypto.PubKey
	PublicKeyHash Uint160
}

func NewAccount() (*Account, error) {
	priKey, pubKey, _ := crypto.GenKeyPair()
	temp, err := pubKey.EncodePoint(true)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	hash, err := ToCodeHash(temp)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	return &Account{
		PrivateKey:    priKey,
		PublicKey:     &pubKey,
		PublicKeyHash: hash,
	}, nil
}

func NewAccountWithPrivatekey(privateKey []byte) (*Account, error) {
	privKeyLen := len(privateKey)

	if privKeyLen != 32 && privKeyLen != 96 && privKeyLen != 104 {
		return nil, errors.New("Invalid private Key.")
	}

	// set public key
	pubKey := crypto.NewPubKey(privateKey)
	//priKey,pubKey,_ := crypto.GenKeyPair()
	temp, err := pubKey.EncodePoint(true)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	hash, err := ToCodeHash(temp)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "[Contract],CreateSignatureContract failed.")
	}
	return &Account{
		PrivateKey:    privateKey,
		PublicKey:     pubKey,
		PublicKeyHash: hash,
	}, nil
}

//get signer's private key
func (ac *Account) PrivKey() []byte {
	return ac.PrivateKey
}

//get signer's public key
func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}
