package account

import (
	"errors"
	. "github.com/Ontology/common"
	"github.com/Ontology/core/contract"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

type Account struct {
	PrivateKey []byte
	PublicKey  *crypto.PubKey
	Address    Uint160
}

func NewAccount() (*Account, error) {
	priKey, pubKey, _ := crypto.GenKeyPair()
	signatureRedeemScript, err := contract.CreateSignatureRedeemScript(&pubKey)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "CreateSignatureRedeemScript failed")
	}
	programHash := ToCodeHash(signatureRedeemScript)
	return &Account{
		PrivateKey: priKey,
		PublicKey:  &pubKey,
		Address:    programHash,
	}, nil
}

func NewAccountWithPrivatekey(privateKey []byte) (*Account, error) {
	privKeyLen := len(privateKey)

	if privKeyLen != 32 && privKeyLen != 96 && privKeyLen != 104 {
		return nil, errors.New("Invalid private Key.")
	}

	pubKey := crypto.NewPubKey(privateKey)
	signatureRedeemScript, err := contract.CreateSignatureRedeemScript(pubKey)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "CreateSignatureRedeemScript failed")
	}
	programHash := ToCodeHash(signatureRedeemScript)
	return &Account{
		PrivateKey: privateKey,
		PublicKey:  pubKey,
		Address:    programHash,
	}, nil
}

func (ac *Account) PrivKey() []byte {
	return ac.PrivateKey
}

func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}
