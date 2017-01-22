package crypto

import (
	"crypto/rand"
)

func Sha256(value []byte) []byte{
	//TODO: implement Sha256

	return nil
}

func RIPEMD160(value []byte) []byte{
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
