package types

import (
	"math/big"
	"github.com/Ontology/vm/neovm/interfaces"
)

type InteropInterface struct {
	_object interfaces.IInteropInterface
}

func NewInteropInterface(value interfaces.IInteropInterface) *InteropInterface {
	var ii InteropInterface
	ii._object = value
	return &ii
}

func (ii *InteropInterface) Equals(other StackItemInterface) bool {
	return false
}

func (ii *InteropInterface) GetBigInteger() *big.Int {
	return big.NewInt(0)
}

func (ii *InteropInterface) GetBoolean() bool {
	if ii._object == nil {
		return false
	}
	return true
}

func (ii *InteropInterface) GetByteArray() []byte {
	return ii._object.ToArray()
}

func (ii *InteropInterface) GetInterface() interfaces.IInteropInterface {
	return ii._object
}

func (ii *InteropInterface) GetArray() []StackItemInterface {
	return []StackItemInterface{}
}
