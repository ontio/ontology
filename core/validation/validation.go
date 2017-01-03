package validation

import (
	"GoOnchain/core/signature"
	"errors"
)


func VerifySignableData(signableData signature.SignableData) error {

	hashes,err := signableData.GetProgramHashes()
	if(err != nil){
		return err
	}

	programs := signableData.GetPrograms()
	Length := len(hashes)
	if(Length != len(programs)){
		return  errors.New("The number of data hashes is different with number of programs.")
	}

	for i := 0; i < Length; i++ {
		if(hashes[i] != programs[i].CodeHash()){
			return errors.New("The data hashes is different with corresponding program code.")
		}

		//TODO: VM integration
		//new scriptEngine
		//engine.ExecuteScript (program.parameter)
		//engine.ExecuteScript (program.code)
	}

	return nil
}
