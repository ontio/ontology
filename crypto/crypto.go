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
	// Get the random number from System urandom
	return rand.GenerateRandomBytes(n)
}
