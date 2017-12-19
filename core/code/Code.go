package code

import (
	. "github.com/Ontology/common"
	. "github.com/Ontology/core/contract"
)

//ICode is the abstract interface of smart contract code.
type ICode interface {
	GetCode() []byte

	GetParameterTypes() []ContractParameterType

	GetReturnTypes() []ContractParameterType

	CodeHash() Uint160
}
