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

func (e *PubKey) DeSerialize(r io.Reader) error {
	//TODO
	return nil
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

func DecodePoint(encoded []byte) *PubKey{
	//TODO: DecodePoint
	return nil
}

type PubKeySlice []*PubKey

func (p PubKeySlice) Len() int           { return len(p) }
func (p PubKeySlice) Less(i, j int) bool {
	//TODO:PubKeySlice Less
	return false
}
func (p PubKeySlice) Swap(i, j int) {
	//TODO:PubKeySlice Swap
}
