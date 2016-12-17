package contract

import (
	"GoOnchain/common"
)

//Contract address is the hash of contract program .
//which be used to control asset or indicate the smart contract address �?


//Contract include the program codes with parameters which can be executed on specific evnrioment
type Contract struct {

	//the contract program code,which will be run on VM or specific envrionment
	Code []byte

	//the Contract Parameter type list
	// describe the number of contract program parameters and the parameter type
	Parameters []ContractParameterType

	//The program hash as contract address
	ProgramHash common.Uint160

	//owner's pubkey hash indicate the owner of contract
	OwnerPubkeyHash common.Uint160

}


