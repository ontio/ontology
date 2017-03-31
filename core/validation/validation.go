package validation

import (
	. "DNA/common"
	sig "DNA/core/signature"
	"DNA/crypto"
	. "DNA/errors"
	"DNA/vm"
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
		var cryptos vm.ICrypto
		cryptos = new(vm.ECDsaCrypto)
		se := vm.NewExecutionEngine(nil, cryptos, nil, signableData)
		se.LoadScript(programs[i].Code, true)
		se.LoadScript(programs[i].Parameter, false)
		se.ExecuteProgram()

		if se.State != vm.HALT {
			return false,NewDetailErr(errors.New("[VM] Finish State not equal to HALT."), ErrNoCode, "")
		}

		if se.Stack.Count() != 1 {
			return false,NewDetailErr(errors.New("[VM] Execute Engine Stack Count Error."), ErrNoCode, "")
		}

		flag := se.Stack.Pop().GetBool()
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
