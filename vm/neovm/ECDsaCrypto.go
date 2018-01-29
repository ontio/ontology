package neovm

import (
	"crypto/sha256"
	"errors"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

type ECDsaCrypto struct {
}

func (c *ECDsaCrypto) Hash160(message []byte) []byte {
	temp := common.ToCodeHash(message)
	return temp.ToArray()
}

func (c *ECDsaCrypto) Hash256(message []byte) []byte {
	temp := sha256.Sum256(message)
	f := sha256.Sum256(temp[:])
	return f[:]
}

func (c *ECDsaCrypto) VerifySignature(message []byte, signature []byte, pubkey []byte) (bool, error) {

	log.Debugf("message: %x", message)
	log.Debugf("signature: %x", signature)
	log.Debugf("pubkey: %x", pubkey)

	pk, err := crypto.DecodePoint(pubkey)
	if err != nil {
		return false, NewDetailErr(errors.New("[ECDsaCrypto], crypto.DecodePoint failed."), ErrNoCode, "")
	}

	err = crypto.Verify(*pk, message, signature)
	if err != nil {
		return false, NewDetailErr(errors.New("[ECDsaCrypto], VerifySignature failed."), ErrNoCode, "")
	}

	return true, nil
}
