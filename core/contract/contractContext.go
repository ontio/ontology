package contract

import (
	"GoOnchain/common"
	sig "GoOnchain/core/signature"
)

type ContractContext struct {
	//TODO: define ContractContext。
	Data sig.SignableData
	ProgramHashes []common.Uint160
	Programs [][]byte
	Parameters [][][]byte
}


func NewContractContext(data sig.SignableData) *ContractContext {
	//TODO: implement NewContractContext
	return nil
}