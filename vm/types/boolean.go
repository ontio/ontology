package types

import (
	"math/big"
)

type Boolean struct {
	value bool
}

func NewBoolean(value bool) *Boolean{
	var b Boolean
	b.value = value
	return &b
}

func (b *Boolean) Equals(other StackItem) bool{
	if _, ok := other.(*Boolean); !ok {
		return false
	}
	if b.value != other.GetBoolean() {
		return false
	}
	return true
}

func (b *Boolean) GetBigInteger() *big.Int{
	if b.value {
		return big.NewInt(1)
	}
	return big.NewInt(0)
}

func (b *Boolean) GetBoolean() bool{
	return b.value
}

func (b *Boolean) GetByteArray() []byte{
	if b.value {
		return []byte{1}
	}
	return []byte{0}
}

func (b *Boolean) GetInterface(){

}

func (b *Boolean) GetArray() []StackItem {
	return []StackItem{b}
}
