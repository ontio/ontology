package types

import (
	"math"
	"math/big"

	"github.com/ontio/ontology/vm/neovm/errors"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/vm/neovm/constants"
)

type NeoVmValueType uint8

const (
	boolType NeoVmValueType = iota
	integerType
	bigintType
	bytearrayType
	interopType
	arrayType
	structType
	mapType
)

type VmValue struct {
	valType   NeoVmValueType
	integer   int64
	bigInt    *big.Int
	byteArray []byte
	array     *[]VmValue // array or struct, since array is mutable, need use pointer here
	mapval    map[string]VmValue
	interop   interface{}
}

func VmValueFromInt64(val int64) VmValue {
	return VmValue{valType: integerType, integer: val}
}

func VmValueFromBytes(val []byte) (result VmValue, err error) {
	if len(val) > constants.MAX_INT_SIZE{
		err = errors.ERR_OVER_MAX_ITEM_SIZE
		return
	}
	result.valType = bytearrayType
	result.byteArray = val
	return
}

func VmValueFromUint64(val uint64) VmValue {
	if val <= math.MaxInt64 {
		return VmValueFromInt64(int64(val))
	}

	b := big.NewInt(0)
	b.SetUint64(val)
	return VmValue{valType: bigintType, bigInt: b}
}

func VmValueFromBigInt(val *big.Int) (result VmValue, err error) {
	value, e := IntValFromBigInt(val)
	if e!= nil {
		err = e
		return
	}

	return VmValueFromIntValue(value), nil
}

func VmValueFromIntValue(val IntValue) VmValue {
	if val.isbig {
		return VmValue{valType: bigintType, bigInt: val.bigint}
	} else {
		return VmValue{valType: integerType, integer: val.integer}
	}
}

func (self *VmValue) AsBytes() ([]byte, error) {
	switch self.valType {
	case integerType, boolType:
		return common.BigIntToNeoBytes(big.NewInt(self.integer)), nil
	case bigintType:
		return common.BigIntToNeoBytes(self.bigInt), nil
	case bytearrayType:
		return self.byteArray, nil
	case arrayType, mapType, structType, interopType:
		return nil, errors.ERR_BAD_TYPE
	default:
		panic("unreacheable!")
	}
}

func (self *VmValue) AsIntValue() (IntValue, error) {
	switch self.valType {
	case integerType, boolType:
		return IntValFromInt(self.integer), nil
	case bigintType:
		return IntValFromBigInt(self.bigInt)
	case bytearrayType:
		return IntValFromNeoBytes(self.byteArray)
	case arrayType, mapType, structType, interopType:
		return IntValue{}, errors.ERR_BAD_TYPE
	default:
		panic("unreachable!")
	}
}

func (self *VmValue) AsBool() (bool, error) {
	switch self.valType {
	case integerType, boolType:
		return self.integer != 0, nil
	case bigintType:
		return self.bigInt.Sign() != 0, nil
	case bytearrayType:
		for _, b := range self.byteArray {
			if b != 0 {
				return true, nil
			}
		}
		return false, nil
	case structType, mapType:
		return true, nil
	case arrayType:
		return false, errors.ERR_BAD_TYPE
	case interopType:
		return self.interop != nil, nil
	default:
		panic("unreachable!")
	}
}

func (self *VmValue) GetMapKey() (string, error) {
	val , err := self.AsBytes()
	if err != nil {
		return "", err
	}
	return string(val), nil
}
