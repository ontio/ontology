package contract

import (
	"GoOnchain/common"
	"GoOnchain/vm"
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

func (c *Contract) IsStandard() bool {
	if len(c.Code) != 35 {
		return false
	}
	if c.Code[0] != 33 || c.Code[34] != byte(vm.OP_CHECKSIG) {
		return false
	}
	return true
}

func (c *Contract) IsMultiSigContract() bool {
	//TODO: IsMultiSigContract
	return false
}

func (c *Contract) GetType() ContractType{
	if c.IsStandard() {
		return SignatureContract
	}
	if c.IsMultiSigContract() {
		return MultiSigContract
	}
	return CustomContract
}

