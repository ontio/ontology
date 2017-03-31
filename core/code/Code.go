package code

import (
	."github.com/DNAProject/DNA/common"
	."github.com/DNAProject/DNA/core/contract"
)
//ICode is the abstract interface of smart contract code.
type ICode interface {

	GetCode() []byte

	GetParameterTypes() []ContractParameterType

	GetReturnTypes() []ContractParameterType

	CodeHash() Uint160

}

