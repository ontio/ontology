package util

import (
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

const (
	HASHLEN       = 32
	PRIVATEKEYLEN = 32
	PUBLICKEYLEN  = 32
	SIGNRLEN      = 32
	SIGNSLEN      = 32
	SIGNATURELEN  = 64
)

// InterfaceCrypto ---
type InterfaceCrypto struct {
	EccParams elliptic.CurveParams
	EccParamA *big.Int
	Curve     elliptic.Curve
}

// RandomNum Generate the "real" random number which can be used for crypto algorithm
func RandomNum(n int) ([]byte, error) {
	// TODO Get the random number from System urandom
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}
	return b, nil
}

// Hash ---
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

func RIPEMD160(value []byte) []byte {
	//TODO: implement RIPEMD160

	return nil
}
