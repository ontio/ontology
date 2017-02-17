package crypto

import (
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	//"crypto/subtle"		// normal crypto operation like compare, assign etc avoid timing attack
	"crypto/ecdsa"
	//"crypto/x509"
)

const (
	HASHLEN       = 32
	PRIVATEKEYLEN = 32
	PUBLICKEYLEN  = 32
	SIGNATURELEN  = 64
)

type crypto struct {
	eccParams elliptic.CurveParams
	curve     elliptic.Curve
}

var Crypto crypto

func Sha256(value []byte) []byte {
	//TODO: implement Sha256

	return nil
}

func RIPEMD160(value []byte) []byte {
	//TODO: implement RIPEMD160

	return nil
}

// Generate the "real" random number which can be used for crypto algorithm
func RandomNum(n int) ([]byte, error) {
	// TODO Get the random number from System urandom
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}
	return b, nil
}

func Hash(data []byte) [HASHLEN]byte {
	return sha256.Sum256(data)
}

// CheckMAC reports whether messageMAC is a valid HMAC tag for message.
func CheckMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func Init() {
	// FixMe init the ECC parameters based on curve type, like secp256k1
	Crypto.curve = elliptic.P256()
}

// @prikey, the private key for sign, the length should be 32 bytes currently
func Sign(prikey []byte, data []byte) ([]byte, error) {
	// if (len(prikey) != PRIVATEKEYLEN) {
	// 	fmt.Printf("Unexpected private key length %d\n", len(prikey))
	// 	return nil, errors.New("Unexpected private key length")
	// }

	digest := Hash(data)

	privateKey := new(ecdsa.PrivateKey)
	privateKey.Curve = Crypto.curve
	// TODO check the return value
	privateKey.D = big.NewInt(0)
	privateKey.D.UnmarshalText(prikey)

	r := big.NewInt(0)
	s := big.NewInt(0)
	//ecdsa.Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error)
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest[:])
	if err != nil {
		fmt.Printf("Sign error\n")
		return nil, err
	}

	signature := r.Bytes()
	signature = append(signature, s.Bytes()...)
	fmt.Printf("Signature : %x, len of signature is %d\n", signature, len(signature))

	return signature, nil
}

type ecdsaSignature struct {
	R, S big.Int
}

// Fixme: the signature length TBD
func Verify(pubkey PubKey, data []byte, signature []byte) (bool, error) {
	ecdsaSig := new(ecdsaSignature)

	len := len(signature)
	if len != SIGNATURELEN {
		fmt.Printf("Unknown signature length %d\n", len)
		return false, errors.New("Unknown signature length")
	}
	ecdsaSig.R.SetBytes(signature[:len/2])
	ecdsaSig.S.SetBytes(signature[len/2:])
	//if ecdsaSig.R.Sign() <= 0 || ecdsaSig.S.Sign() <= 0 {
	//	return false, errors.New("ECDSA signature contained zero or negative values")
	//}

	digest := Hash(data)

	pub := new(ecdsa.PublicKey)
	pub.Curve = Crypto.curve
	pub.X = pubkey.X
	pub.Y = pubkey.Y

	//ecdsa.Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool
	return ecdsa.Verify(pub, digest[:], &ecdsaSig.R, &ecdsaSig.S), nil
}
