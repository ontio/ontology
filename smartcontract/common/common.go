package common

import (
	"github.com/Ontology/vm/neovm/types"
	"github.com/Ontology/common"
	"fmt"
	"reflect"
)

type States struct {
	Key string
	Value interface{}
}

func ConvertTypes(item types.StackItemInterface) (results []States) {
	if item == nil {
		return
	}
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
			arr = append(arr, ConvertTypes(val)...)
		}
		results = append(results, States{"Array", arr})
	case *types.InteropInterface:
		results = append(results, States{"InteropInterface", common.ToHexString(v.GetInterface().ToArray())})
	case types.StackItemInterface:
		ConvertTypes(v)
	default:
		panic(fmt.Sprintf("[ConvertTypes] Invalid Types: %v", reflect.TypeOf(v)))
	}
	return
}

func ConvertReturnTypes(item types.StackItemInterface) (results []interface{}) {
	if item == nil {
		return
	}
	switch v := item.(type) {
	case *types.ByteArray:
		results = append(results, common.ToHexString(v.GetByteArray()))
	case *types.Integer:
		results = append(results, v.GetBigInteger())
	case *types.Boolean:
		results = append(results, v.GetBoolean())
	case *types.Array:
		var arr []interface{}
		for _, val := range v.GetArray() {
			arr = append(arr, ConvertReturnTypes(val)...)
		}
		results = append(results, arr)
	case *types.InteropInterface:
		results = append(results, common.ToHexString(v.GetInterface().ToArray()))
	case types.StackItemInterface:
		ConvertTypes(v)
	default:
		panic("[ConvertTypes] Invalid Types!")
	}
	return
}