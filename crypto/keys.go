package crypto

import (
	"io"
	"errors"
	//"math/big"
	"crypto/rand"
	"crypto/ecdsa"
)

type PubKey ECPoint

func (e *PubKey) Serialize(w io.Writer) {
	//TODO: implement PubKey.serialize
}

func (ep *PubKey) EncodePoint(commpressed bool) []byte{
	//TODO: EncodePoint
	return nil
}

func NewPubKey(prikey []byte) *PubKey{
       //TODO: NewPubKey
       return nil
}

func GenPrivKey() []byte {
	return nil
}

//FIXME, does the privkey need base58 encoding?
//This generates a public & private key pair
func GenKeyPair() ([]byte, PubKey, error) {
	pubkey := new(PubKey)
	privatekey := new(ecdsa.PrivateKey)
	privatekey, err := ecdsa.GenerateKey(Crypto.curve, rand.Reader)
	if err != nil {
		return nil, *pubkey, errors.New("Generate key pair error")
	}

	privkey, err := privatekey.D.MarshalText()
	pubkey.X = privatekey.PublicKey.X
	pubkey.Y = privatekey.PublicKey.Y
	return privkey, *pubkey, nil
}
