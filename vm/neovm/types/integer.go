package types

import (
	"math/big"
	"github.com/Ontology/vm/neovm/interfaces"
	"github.com/Ontology/common"
)

type Integer struct {
	value *big.Int
}

func NewInteger(value *big.Int) *Integer {
	var i Integer
	i.value = value
	return &i
}

func (i *Integer) Equals(other StackItemInterface) bool {
	if _, ok := other.(*Integer); !ok {
		return false
	}
	if i.value.Cmp(other.GetBigInteger()) != 0 {
		return false
	}
	return true
}

func (i *Integer) GetBigInteger() *big.Int {
	return i.value
}

func (i *Integer) GetBoolean() bool {
	if i.value.Cmp(big.NewInt(0)) == 0 {
		return false
	}
	return true
}

func (i *Integer) GetByteArray() []byte {
	return common.BytesReverse(append([]byte{}, i.value.Bytes()...))
}

func (i *Integer) GetInterface() interfaces.IInteropInterface {
	return nil
}

func (i *Integer) GetArray() []StackItemInterface {
	return []StackItemInterface{i}
}
