package interfaces

type ICrypto interface {
	Hash160(message []byte) []byte

	Hash256(message []byte) []byte

	VerifySignature(message []byte, signature []byte, pubkey []byte) (bool, error)
}
