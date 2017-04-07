package validation

import (
	. "DNA/common"
	sig "DNA/core/signature"
	"DNA/crypto"
	. "DNA/errors"
	"DNA/vm"
	"DNA/vm/interfaces"
	"errors"
)

func VerifySignableData(signableData sig.SignableData) (bool,error) {

	hashes, err := signableData.GetProgramHashes()
	if err != nil {
		return false,err
	}

	programs := signableData.GetPrograms()
	Length := len(hashes)
	if Length != len(programs) {
		return false,errors.New("The number of data hashes is different with number of programs.")
	}

	programs = signableData.GetPrograms()
	for i := 0; i < len(programs); i++ {
		temp,_ := ToCodeHash(programs[i].Code)
		if hashes[i] != temp {
			return false,errors.New("The data hashes is different with corresponding program code.")
		}
		//execute program on VM
		var cryptos interfaces.ICrypto
		cryptos = new(vm.ECDsaCrypto)
		se := vm.NewExecutionEngine(signableData, cryptos, 1200, nil, nil)
		se.LoadScript(programs[i].Code, false)
		se.LoadScript(programs[i].Parameter, true)
		se.Execute()

		if se.GetState() != vm.HALT {
			return false,NewDetailErr(errors.New("[VM] Finish State not equal to HALT."), ErrNoCode, "")
		}

		if se.GetEvaluationStack().Count() != 1 {
			return false,NewDetailErr(errors.New("[VM] Execute Engine Stack Count Error."), ErrNoCode, "")
		}

		flag := se.GetExecuteResult()
		if !flag {
			return false,NewDetailErr(errors.New("[VM] Check Sig FALSE."), ErrNoCode, "")
		}
	}

	return true,nil
}

func VerifySignature(signableData sig.SignableData, pubkey *crypto.PubKey, signature []byte) (bool,error) {
	ok, err := crypto.Verify(*pubkey, sig.GetHashForSigning(signableData), signature)
	if !ok {
		return false,NewDetailErr(err, ErrNoCode, "[Validation], VerifySignature failed.")
	} else {
		return true,nil
	}
}
