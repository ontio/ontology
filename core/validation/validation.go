package validation

import (
	"errors"
	. "github.com/Ontology/common"
	sig "github.com/Ontology/core/signature"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
)

func VerifySignableDataProgramHashes(signableData sig.SignableData) error {
	hashes, err := signableData.GetProgramHashes()
	if err != nil {
		return err
	}

	programs := signableData.GetPrograms()
	Length := len(hashes)
	if Length != len(programs) {
		return errors.New("the number of data hashes is different with number of programs")
	}

	for i := 0; i < len(programs); i++ {
		temp := ToCodeHash(programs[i].Code)
		if hashes[i] != temp {
			return errors.New("the data hashes is different with corresponding program code")
		}
	}

	return nil
}

func VerifySignature(signableData sig.SignableData, pubkey *crypto.PubKey, signature []byte) error {
	err := crypto.Verify(*pubkey, sig.GetHashData(signableData), signature)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "[Validation], VerifySignature failed.")
	} else {
		return nil
	}
}
