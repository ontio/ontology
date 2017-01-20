package contract

import (
	"GoOnchain/common"
	sig "GoOnchain/core/signature"
	"GoOnchain/crypto"
	"GoOnchain/core/contract/program"
)

type ContractContext struct {
	//TODO: define ContractContext。
	Data sig.SignableData
	ProgramHashes []common.Uint160
	Programs [][]byte
	Parameters [][][]byte
}


func NewContractContext(data sig.SignableData) *ContractContext {

	programHashes,_ := data.GetProgramHashes() //TODO: check error
	hashLen := len(programHashes)

	return &ContractContext{
		Data: data,
		ProgramHashes: programHashes,
		Programs: make([][]byte,hashLen),
		Parameters: make([][][]byte,hashLen),
	}
}

func (cxt *ContractContext) AddContract(contract *Contract, pubkey *crypto.PubKey,paramenter []byte ) error {
	//TODO: implement AddContract()

	//TODO: check contract type for diff building
	return  nil
}


func (cxt *ContractContext) GetPrograms() ([]*program.Program) {
	//TODO: implement GetProgram()

	return  []*program.Program{}

}
