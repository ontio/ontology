package common

import (
	"testing"
	"github.com/Ontology/vm/neovm/types"
	"math/big"
)

func TestConvertTypes(t *testing.T) {
	arr := types.NewArray([]types.StackItemInterface{types.NewByteArray([]byte{1,2,3}), types.NewInteger(big.NewInt(32))})
	var states []States
	for _, v := range arr.GetArray() {
		states = append(states, ConvertTypes(v)...)
	}
	t.Log("result:", states)
}

func TestConvertReturnTypes(t *testing.T) {
	arr := types.NewArray([]types.StackItemInterface{types.NewByteArray([]byte{1,2,3}), types.NewInteger(big.NewInt(32)), types.NewArray([]types.StackItemInterface{types.NewByteArray([]byte{1,2,3}), types.NewInteger(big.NewInt(32))})})
	var states []interface{}
	for _, v := range arr.GetArray() {
		states = append(states, ConvertReturnTypes(v)...)
	}
	t.Log("result:", states)
}
