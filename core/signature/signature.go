package signature

import (
	"GoOnchain/common"
	"GoOnchain/core/contract/program"
)

//SignableData describe the data need be signed.
type SignableData interface {

	//Get the the SignableData's program hashes
	GetProgramHashes() ([]common.Uint160, error)

	SetPrograms([]*program.Program)
}


func Sign(data SignableData,signer Signer) ([]byte, error){

	//TODO: implement Sign()
	return  []byte{},nil
}

