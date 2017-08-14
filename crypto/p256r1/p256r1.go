package p256r1

import (
	"DNA/crypto/util"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
)

func Init(algSet *util.CryptoAlgSet) {
	algSet.Curve = elliptic.P256()
	algSet.EccParams = *(algSet.Curve.Params())
}

func GenKeyPair(algSet *util.CryptoAlgSet) ([]byte, *big.Int, *big.Int, error) {
	privateKey := new(ecdsa.PrivateKey)
	privateKey, err := ecdsa.GenerateKey(algSet.Curve, rand.Reader)
	if err != nil {
		return nil, nil, nil, errors.New("Generate key pair error")
	}

	priKey := privateKey.D.Bytes()
	return priKey, privateKey.PublicKey.X, privateKey.PublicKey.Y, nil
}

func Sign(algSet *util.CryptoAlgSet, priKey []byte, data []byte) (*big.Int, *big.Int, error) {
	digest := util.Hash(data)

	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = algSet.Curve
	privateKey.D = big.NewInt(0)
	privateKey.D.SetBytes(priKey)

	r := big.NewInt(0)
	s := big.NewInt(0)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest[:])
	if err != nil {
		fmt.Printf("Sign error\n")
		return nil, nil, err
	}
	return r, s, nil
}

func Verify(algSet *util.CryptoAlgSet, X *big.Int, Y *big.Int, data []byte, r, s *big.Int) error {
	digest := util.Hash(data)

	pub := new(ecdsa.PublicKey)
	pub.Curve = algSet.Curve

	pub.X = new(big.Int).Set(X)
	pub.Y = new(big.Int).Set(Y)

	if ecdsa.Verify(pub, digest[:], r, s) {
		return nil
	} else {
		return errors.New("[Validation], Verify failed.")
	}
}
