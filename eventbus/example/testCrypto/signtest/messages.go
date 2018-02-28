package signtest

import "github.com/Ontology/crypto"

type SignRequest struct {
	Data []byte
	Seq  string
}

type SignResponse struct {
	Signature []byte
	Seq  string
}

type SetPrivKey struct{
	PrivKey []byte
}



type VerifyRequest struct {
	Signature []byte
	Data []byte
	PublicKey crypto.PubKey
	Seq string
}

type VerifyResponse struct {
	Seq string
	Result bool
	ErrorMsg string
}