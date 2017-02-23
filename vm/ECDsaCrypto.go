package vm

import (
	"GoOnchain/crypto"
	. "GoOnchain/errors"
	"errors"
	"fmt"
)


type ECDsaCrypto struct  {
}

func (c * ECDsaCrypto) Hash160( message []byte ) []byte {
	return []byte{}
}

func (c * ECDsaCrypto) Hash256( message []byte ) []byte {
	return []byte{}
}

func (c * ECDsaCrypto) VerifySignature(message []byte,signature []byte, pubkey []byte) (bool,error) {

	fmt.Printf( "message: %x \n", message )
	fmt.Printf( "signature: %x \n", signature )
	fmt.Printf( "pubkey: %x \n", pubkey )

	pk,err := crypto.DecodePoint(pubkey)
	if err != nil {
		return false,NewDetailErr(errors.New("[ECDsaCrypto], crypto.DecodePoint failed."), ErrNoCode, "")
	}

	temp ,err := crypto.Verify(*pk, message,signature)
	if !temp {
		return false,NewDetailErr(errors.New("[ECDsaCrypto], VerifySignature failed."), ErrNoCode, "")
	}

	return true,nil
}