package types

import (
	"github.com/Ontology/common"
	"github.com/Ontology/vm/neovm/interfaces"
	"math/big"
)

type ByteArray struct {
	value []byte
}

func NewByteArray(value []byte) *ByteArray {
	var ba ByteArray
	ba.value = value
	return &ba
}

func (ba *ByteArray) Equals(other StackItemInterface) bool {
	if _, ok := other.(*ByteArray); !ok {
		return false
	}
	a1 := ba.value
	a2 := other.GetByteArray()
	l1 := len(a1)
	l2 := len(a2)
	if l1 != l2 {
		return false
	}
	for i := 0; i < l1; i++ {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func (ba *ByteArray) GetBigInteger() *big.Int {
	res := big.NewInt(0)
	l := len(ba.value)
	if l == 0 {
		return res
	}

	bytes := make([]byte, 0, l)
	bytes = append(bytes, ba.value...)
	common.BytesReverse(bytes)

	if bytes[0] >> 7 == 1 {
		for i, b := range bytes {
			bytes[i] = ^b
		}

		temp := big.NewInt(0)
		temp.SetBytes(bytes)
		temp2 := big.NewInt(0)
		temp2.Add(temp, big.NewInt(1))
		bytes = temp2.Bytes()
		res.SetBytes(bytes)
		return res.Neg(res)
	}

	res.SetBytes(bytes)
	return res
}

func (ba *ByteArray) GetBoolean() bool {
	for _, b := range ba.value {
		if b != 0 {
			return true
		}
	}
	return false
}

func (ba *ByteArray) GetByteArray() []byte {
	return ba.value
}

func (ba *ByteArray) GetInterface() interfaces.IInteropInterface {
	return nil
}

func (ba *ByteArray) GetArray() []StackItemInterface {
	return []StackItemInterface{ba}
}
