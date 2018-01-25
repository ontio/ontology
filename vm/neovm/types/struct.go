package types

import (
	"math/big"
	"github.com/Ontology/vm/neovm/interfaces"
)

type Struct struct {
	_array []StackItemInterface
}

func NewStruct(value []StackItemInterface) *Struct {
	var s Struct
	s._array = value
	return &s
}

func (s *Struct) Equals(other StackItemInterface) bool {
	if _, ok := other.(*Struct); !ok {
		return false
	}
	a1 := s._array
	a2 := other.GetStruct()
	l1 := len(a1)
	l2 := len(a2)
	if l1 != l2 {
		return false
	}
	for i := 0; i < l1; i++ {
		if !a1[i].Equals(a2[i]) {
			return false
		}
	}
	return true
}

func (s *Struct) GetBigInteger() *big.Int {
	if len(s._array) == 0 {
		return big.NewInt(0)
	}
	return s._array[0].GetBigInteger()
}

func (s *Struct) GetBoolean() bool {
	if len(s._array) == 0 {
		return false
	}
	return s._array[0].GetBoolean()
}

func (s *Struct) GetByteArray() []byte {
	if len(s._array) == 0 {
		return []byte{}
	}
	return s._array[0].GetByteArray()
}

func (s *Struct) GetInterface() interfaces.IInteropInterface {
	if len(s._array) == 0 {
		return nil
	}
	return s._array[0].GetInterface()
}

func (s *Struct) GetArray() []StackItemInterface {
	return s._array
}

func (s *Struct) GetStruct() []StackItemInterface {
	return s._array
}

func (s *Struct) Clone() StackItemInterface {
	var arr []StackItemInterface
	for _, v := range s._array {
		if value, ok := v.(*Struct); ok {
			arr = append(arr, value.Clone())
		} else {
			arr = append(arr, value)
		}
	}
	return &Struct{arr}
}
