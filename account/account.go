package account

import (
	. "DNA/common"
	"DNA/core/contract"
	"DNA/crypto"
	. "DNA/errors"
	"errors"
)

type Account struct {
	PrivateKey  []byte
	PublicKey   *crypto.PubKey
	ProgramHash Uint160
}

func NewAccount() (*Account, error) {
	priKey, pubKey, _ := crypto.GenKeyPair()
	signatureRedeemScript, err := contract.CreateSignatureRedeemScript(&pubKey)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "CreateSignatureRedeemScript failed")
	}
	programHash, err := ToCodeHash(signatureRedeemScript)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "ToCodeHash failed")
	}
	return &Account{
		PrivateKey:  priKey,
		PublicKey:   &pubKey,
		ProgramHash: programHash,
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
	programHash, err := ToCodeHash(signatureRedeemScript)
	if err != nil {
		return nil, NewDetailErr(err, ErrNoCode, "ToCodeHash failed")
	}
	return &Account{
		PrivateKey:  privateKey,
		PublicKey:   pubKey,
		ProgramHash: programHash,
	}, nil
}

func (ac *Account) PrivKey() []byte {
	return ac.PrivateKey
}

func (ac *Account) PubKey() *crypto.PubKey {
	return ac.PublicKey
}
