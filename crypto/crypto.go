package crypto

import (
	"GoOnchain/crypto/p256r1"
	"GoOnchain/crypto/sm2"
	"GoOnchain/crypto/util"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"
)

const (
	P256R1 = 0
	SM2    = 1
)

// AlgChoice the alg choice, it can be P256R1 or SM2
var AlgChoice int

// Crypto ---
var Crypto util.InterfaceCrypto

// PubKey ---
type PubKey struct {
	X, Y *big.Int
}

func init() {
	AlgChoice = 0
}

func SetAlg(algChoice int) {
	AlgChoice = algChoice
	Crypto.EccParamA = new(big.Int)
	if SM2 == algChoice {
		sm2.Init(&Crypto)
	} else {
		p256r1.Init(&Crypto)
	}
	return
}

//GenKeyPair FIXME, does the privkey need base58 encoding?
func GenKeyPair() ([]byte, PubKey, error) {
	mPubKey := new(PubKey)
	var privD []byte
	var X *big.Int
	var Y *big.Int
	var err error

	if SM2 == AlgChoice {
		privD, X, Y, err = sm2.GenKeyPair(&Crypto)
	} else {
		privD, X, Y, err = p256r1.GenKeyPair(&Crypto)
	}

	if nil != err {
		return nil, *mPubKey, err
	}

	mPubKey.X = new(big.Int).Set(X)
	mPubKey.Y = new(big.Int).Set(Y)
	return privD, *mPubKey, nil
}

// Sign @prikey, the private key for sign, the length should be 32 bytes currently
func Sign(prikey []byte, data []byte) ([]byte, error) {
	var r *big.Int
	var s *big.Int
	var err error

	if SM2 == AlgChoice {
		r, s, err = sm2.Sign(&Crypto, prikey, data)
	} else {
		r, s, err = p256r1.Sign(&Crypto, prikey, data)
	}
	if err != nil {
		return nil, err
	}

	signature := make([]byte, util.SIGNATURELEN)

	lenR := len(r.Bytes())
	lenS := len(s.Bytes())
	copy(signature[util.SIGNRLEN-lenR:], r.Bytes())
	copy(signature[util.SIGNATURELEN-lenS:], s.Bytes())
	return signature, nil
}

// Verify Fixme: the signature length TBD
func Verify(pubkey PubKey, data []byte, signature []byte) (bool, error) {
	len := len(signature)
	if len != util.SIGNATURELEN {
		fmt.Printf("Unknown signature length %d\n", len)
		return false, errors.New("Unknown signature length")
	}

	r := new(big.Int).SetBytes(signature[:len/2])
	s := new(big.Int).SetBytes(signature[len/2:])

	if SM2 == AlgChoice {
		return sm2.Verify(&Crypto, pubkey.X, pubkey.Y, data, r, s)
	}
	return p256r1.Verify(&Crypto, pubkey.X, pubkey.Y, data, r, s)
}

// Serialize ---
func (e *PubKey) Serialize(w io.Writer) {
	//TODO: implement PubKey.serialize
}

// DeSerialize ---
func (e *PubKey) DeSerialize(r io.Reader) error {
	//TODO
	return nil
}

type PubKeySlice []*PubKey

func (p PubKeySlice) Len() int { return len(p) }
func (p PubKeySlice) Less(i, j int) bool {
	//TODO:PubKeySlice Less
	return false
}
func (p PubKeySlice) Swap(i, j int) {
	//TODO:PubKeySlice Swap
}

func Sha256(value []byte) []byte {
	data := make([]byte, 32)
	digest := sha256.Sum256(value)
	copy(data, digest[0:32])
	return data
}
