package common

import (
	"github.com/Ontology/vm/neovm/types"
	"github.com/Ontology/common"
)

type States struct {
	Key string
	Value []string
}

func ConvertTypes(item types.StackItemInterface) (results []States) {
	switch v := item.(type) {
	case *types.ByteArray:
		results = append(results, States{"ByteArray", common.ToHexString(v.GetByteArray())})
	case *types.Integer:
		if v.GetBigInteger().Sign() == 0 {
			results = append(results, States{"Integer", common.ToHexString([]byte{0})})
		} else {
			results = append(results, States{"Integer", common.ToHexString(types.ConvertBigIntegerToBytes(v.GetBigInteger()))})
		}
	case *types.Boolean:
		if v.GetBoolean() {
			results = append(results, States{"Boolean", common.ToHexString([]byte{1})})
		} else {
			results = append(results, States{"Boolean", common.ToHexString([]byte{0})})
		}
	case *types.Array:
		var arr []States
		for _, val := range v.GetArray() {
			arr = append(arr, ConvertTypes(val))
		}
		results = append(results, States{"Array", arr})
	case *types.InteropInterface:
		results = append(results, States{"InteropInterface", common.ToHexString(v.GetInterface().ToArray())})
	case *types.StackItemInterface:
		ConvertTypes(v)
	default:
		panic("[ConvertTypes] Invalid Types!")
	}
	return
}