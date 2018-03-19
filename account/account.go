package account

import (
	"errors"
	. "github.com/Ontology/common"
	"github.com/Ontology/crypto"
	"github.com/Ontology/core/types"
)

type Account struct {
	PrivateKey []byte
	PublicKey  *crypto.PubKey
	Address    Uint160
}

func NewAccount() *Account {
	priKey, pubKey, _ := crypto.GenKeyPair()
	address := types.AddressFromPubKey(&pubKey)
	return &Account{
		PrivateKey: priKey,
		PublicKey:  &pubKey,
		Address:    address,
	}
}

func NewAccountWithPrivatekey(privateKey []byte) (*Account, error) {
	privKeyLen := len(privateKey)

	if privKeyLen != 32 && privKeyLen != 96 && privKeyLen != 104 {
		return nil, errors.New("Invalid private Key.")
	}

	pubKey := crypto.NewPubKey(privateKey)
	address := types.AddressFromPubKey(pubKey)

	return &Account{
		PrivateKey: privateKey,
		PublicKey:  pubKey,
		Address:    address,
	}, nil
}

func (ac *Account) PrivKey() []byte {
	return ac.PrivateKey
}

func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}
