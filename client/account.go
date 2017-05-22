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

	pubKey := crypto.NewPubKey(privateKey)
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

func (ac *Account) PrivKey() []byte {
	return ac.PrivateKey
}

func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}
