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

// P256PARAMA is the param A in p256r1
const P256PARAMA = -3

// Init ---
func Init(Crypto *util.InterfaceCrypto) {
	Crypto.Curve = elliptic.P256()
	Crypto.EccParams = *(Crypto.Curve.Params())
	Crypto.EccParamA.Set(big.NewInt(P256PARAMA))
}

// GenKeyPair FIXME, does the privkey need base58 encoding?
func GenKeyPair(Crypto *util.InterfaceCrypto) ([]byte, *big.Int, *big.Int, error) {
	privatekey := new(ecdsa.PrivateKey)
	privatekey, err := ecdsa.GenerateKey(Crypto.Curve, rand.Reader)
	if err != nil {
		return nil, nil, nil, errors.New("Generate key pair error")
	}

	privkey := privatekey.D.Bytes()
	return privkey, privatekey.PublicKey.X, privatekey.PublicKey.Y, nil
}

// Sign @prikey, the private key for sign, the length should be 32 bytes currently
func Sign(Crypto *util.InterfaceCrypto, priKey []byte, data []byte) (*big.Int, *big.Int, error) {
	// if (len(priKey) != PRIVATEKEYLEN) {
	// 	fmt.Printf("Unexpected private key length %d\n", len(prikey))
	// 	return nil, errors.New("Unexpected private key length")
	// }

	digest := util.Hash(data)

	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = Crypto.Curve
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

// Verify Fixme: the signature length TBD
func Verify(Crypto *util.InterfaceCrypto, X *big.Int, Y *big.Int, data []byte, r, s *big.Int) (bool, error) {
	/*if r.Sign() <= 0 || s.Sign() <= 0 {
		return false, errors.New("ECDSA signature contained zero or negative values")
	}
	*/
	digest := util.Hash(data)

	pub := new(ecdsa.PublicKey)
	pub.Curve = Crypto.Curve

	pub.X = new(big.Int).Set(X)
	pub.Y = new(big.Int).Set(Y)

	return ecdsa.Verify(pub, digest[:], r, s), nil
}
