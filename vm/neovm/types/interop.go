package types

import (
	"math/big"
	"github.com/Ontology/vm/neovm/interfaces"
	"github.com/Ontology/common"
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
	if _, ok := other.(*InteropInterface); !ok {
		return false
	}
	if !common.IsEqualBytes(ii._object.ToArray(), other.GetInterface().ToArray()) {
		return false
	}
	return true
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
	return []StackItemInterface{ii}
}

func (ii *InteropInterface) GetStruct() []StackItemInterface {
	return []StackItemInterface{ii}
}

func (ii *InteropInterface) Clone() StackItemInterface {
	return &InteropInterface{ii._object.Clone()}
}
