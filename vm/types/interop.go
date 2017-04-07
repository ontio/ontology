package types

import (
	"math/big"
	"DNA/vm/interfaces"
)

type InteropInterface struct {
	_object interfaces.IInteropInterface
}

func NewInteropInterface(value interfaces.IInteropInterface) *InteropInterface {
	var ii InteropInterface
	ii._object = value
	return &ii
}

func (ii *InteropInterface) Equals() bool {
	return false
}

func (ii *InteropInterface) GetBigInteger() big.Int {
	return big.Int{}
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

func (ii *InteropInterface) GetInterface() {
}

func (ii *InteropInterface) GetArray() []StackItem {
	return []StackItem{}
}
