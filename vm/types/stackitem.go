package types

import (
	"math/big"
)

type StackItem interface {
	Equals(other StackItem) bool
	GetBigInteger() *big.Int
	GetBoolean() bool
	GetByteArray() []byte
	GetInterface()
	GetArray() []StackItem
}
