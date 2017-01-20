package signature

import (
	"GoOnchain/crypto"
)

//Signer is the abstract interface of user's information(Keys) for signing data.
type Signer interface {

	//get signer's private key
	PrivKey() []byte

	//get signer's public key
	PubKey() *crypto.PubKey

}

