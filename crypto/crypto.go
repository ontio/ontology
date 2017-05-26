package crypto

import (
	"DNA/common/serialization"
	"DNA/crypto/p256r1"
	"DNA/crypto/sm2"
	"DNA/crypto/util"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
)

const (
	P256R1 = 0
	SM2    = 1
)

//It can be P256R1 or SM2
var AlgChoice int

var algSet util.CryptoAlgSet

type PubKey struct {
	X, Y *big.Int
}

func init() {
	AlgChoice = 0
}

func SetAlg(algChoice string) {
	if strings.Compare("SM2", algChoice) == 0 {
		AlgChoice = SM2
		sm2.Init(&algSet)
	} else {
		AlgChoice = P256R1
		p256r1.Init(&algSet)
	}
	return
}

func GenKeyPair() ([]byte, PubKey, error) {
	mPubKey := new(PubKey)
	var privateD []byte
	var X *big.Int
	var Y *big.Int
	var err error

	if SM2 == AlgChoice {
		privateD, X, Y, err = sm2.GenKeyPair(&algSet)
	} else {
		privateD, X, Y, err = p256r1.GenKeyPair(&algSet)
	}

	if nil != err {
		return nil, *mPubKey, err
	}

	mPubKey.X = new(big.Int).Set(X)
	mPubKey.Y = new(big.Int).Set(Y)
	return privateD, *mPubKey, nil
}

func Sign(privateKey []byte, data []byte) ([]byte, error) {
	var r *big.Int
	var s *big.Int
	var err error

	if SM2 == AlgChoice {
		r, s, err = sm2.Sign(&algSet, privateKey, data)
	} else {
		r, s, err = p256r1.Sign(&algSet, privateKey, data)
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

func Verify(publicKey PubKey, data []byte, signature []byte) (bool, error) {
	len := len(signature)
	if len != util.SIGNATURELEN {
		fmt.Printf("Unknown signature length %d\n", len)
		return false, errors.New("Unknown signature length")
	}

	r := new(big.Int).SetBytes(signature[:len/2])
	s := new(big.Int).SetBytes(signature[len/2:])

	if SM2 == AlgChoice {
		return sm2.Verify(&algSet, publicKey.X, publicKey.Y, data, r, s)
	}
	return p256r1.Verify(&algSet, publicKey.X, publicKey.Y, data, r, s)
}

func (e *PubKey) Serialize(w io.Writer) error {
	bufX := []byte{}
	if e.X.Sign() == -1 {
		// prefix 0x00 means the big number X is negative
		bufX = append(bufX, 0x00)
	}
	bufX = append(bufX, e.X.Bytes()...)

	if err := serialization.WriteVarBytes(w, bufX); err != nil {
		return err
	}

	bufY := []byte{}
	if e.Y.Sign() == -1 {
		// prefix 0x00 means the big number Y is negative
		bufY = append(bufY, 0x00)
	}
	bufY = append(bufY, e.Y.Bytes()...)
	if err := serialization.WriteVarBytes(w, bufY); err != nil {
		return err
	}
	return nil
}

func (e *PubKey) DeSerialize(r io.Reader) error {
	bufX, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	e.X = big.NewInt(0)
	e.X = e.X.SetBytes(bufX)
	if len(bufX) == util.NEGBIGNUMLEN {
		e.X.Neg(e.X)
	}
	bufY, err := serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}
	e.Y = big.NewInt(0)
	e.Y = e.Y.SetBytes(bufY)
	if len(bufY) == util.NEGBIGNUMLEN {
		e.Y.Neg(e.Y)
	}
	return nil
}

type PubKeySlice []*PubKey

func (p PubKeySlice) Len() int { return len(p) }
func (p PubKeySlice) Less(i, j int) bool {
	r := p[i].X.Cmp(p[j].X)
	if r <= 0 {
		return true
	}
	return false
}
func (p PubKeySlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func Sha256(value []byte) []byte {
	data := make([]byte, 32)
	digest := sha256.Sum256(value)
	copy(data, digest[0:32])
	return data
}
