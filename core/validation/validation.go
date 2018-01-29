package validation

import (
	"errors"
	. "github.com/Ontology/common"
	sig "github.com/Ontology/core/signature"
	"github.com/Ontology/crypto"
	. "github.com/Ontology/errors"
	"github.com/Ontology/smartcontract/service"
	"github.com/Ontology/smartcontract/types"
	vm "github.com/Ontology/vm/neovm"
	"github.com/Ontology/vm/neovm/interfaces"
)

func VerifySignableData(signableData sig.SignableData) error {

	hashes, err := signableData.GetProgramHashes()
	if err != nil {
		return err
	}

	programs := signableData.GetPrograms()
	Length := len(hashes)
	if Length != len(programs) {
		return errors.New("The number of data hashes is different with number of programs.")
	}

	programs = signableData.GetPrograms()
	for i := 0; i < len(programs); i++ {
		temp := ToCodeHash(programs[i].Code)
		if hashes[i] != temp {
			return errors.New("The data hashes is different with corresponding program code.")
		}
		//execute program on VM
		var cryptos interfaces.ICrypto
		cryptos = new(vm.ECDsaCrypto)
		stateReader := service.NewStateReader(types.Verification)
		se := vm.NewExecutionEngine(signableData, cryptos, nil, stateReader)
		se.LoadCode(programs[i].Code, false)
		se.LoadCode(programs[i].Parameter, true)
		se.Execute()

		if se.GetState() != vm.HALT {
			return NewDetailErr(errors.New("[VM] Finish State not equal to HALT."), ErrNoCode, "")
		}

		if se.GetEvaluationStack().Count() != 1 {
			return NewDetailErr(errors.New("[VM] Execute Engine Stack Count Error."), ErrNoCode, "")
		}

		flag := se.GetExecuteResult()
		if !flag {
			return NewDetailErr(errors.New("[VM] Check Sig FALSE."), ErrNoCode, "")
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
