/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package crossvm_codec

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

const (
	ByteArrayType byte = 0x00
	StringType    byte = 0x01
	AddressType   byte = 0x02
	BooleanType   byte = 0x03
	IntType       byte = 0x04
	H256Type      byte = 0x05

	//reserved for other types
	ListType byte = 0x10

	MAX_PARAM_LENGTH      = 1024
	VERSION          byte = 0
)

var ERROR_PARAM_FORMAT = fmt.Errorf("error param format")
var ERROR_PARAM_NOT_SUPPORTED_TYPE = fmt.Errorf("error param format:not supported type")

// currently only used by test case
func EncodeValue(value interface{}) ([]byte, error) {
	sink := common.NewZeroCopySink(nil)
	switch val := value.(type) {
	case []byte:
		EncodeBytes(sink, val)
	case string:
		EncodeString(sink, val)
	case common.Address:
		EncodeAddress(sink, val)
	case bool:
		EncodeBool(sink, val)
	case common.Uint256:
		EncodeH256(sink, val)
	case *big.Int:
		err := EncodeBigInt(sink, val)
		if err != nil {
			return nil, err
		}
	case int:
		EncodeInt128(sink, common.I128FromInt64(int64(val)))
	case int64:
		EncodeInt128(sink, common.I128FromInt64(val))
	case []interface{}:
		err := EncodeList(sink, val)
		if err != nil {
			return nil, err
		}
	default:
		log.Warn("encode value: unsupported type:", reflect.TypeOf(val).String())
	}

	return sink.Bytes(), nil
}

func DecodeValue(source *common.ZeroCopySource) (interface{}, error) {
	ty, eof := source.NextByte()
	if eof {
		return nil, ERROR_PARAM_FORMAT
	}

	switch ty {
	case ByteArrayType:
		size, eof := source.NextUint32()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		buf, eof := source.NextBytes(uint64(size))
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		return buf, nil
	case StringType:
		size, eof := source.NextUint32()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		buf, eof := source.NextBytes(uint64(size))
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		return string(buf), nil
	case AddressType:
		addr, eof := source.NextAddress()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		return addr, nil
	case BooleanType:
		by, irr, eof := source.NextBool()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		if irr {
			return nil, ERROR_PARAM_FORMAT
		}

		return by, nil
	case IntType:
		val, eof := source.NextI128()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		return val.ToBigInt(), nil
	case H256Type:
		hash, eof := source.NextHash()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		return hash, nil
	case ListType:
		size, eof := source.NextUint32()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		list := make([]interface{}, 0)
		for i := uint32(0); i < size; i++ {
			val, err := DecodeValue(source)
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		}

		return list, nil
	default:
		return nil, ERROR_PARAM_NOT_SUPPORTED_TYPE
	}
}

func EncodeBytes(sink *common.ZeroCopySink, buf []byte) {
	sink.WriteByte(ByteArrayType)
	sink.WriteUint32(uint32(len(buf)))
	sink.WriteBytes(buf)
}

func EncodeString(sink *common.ZeroCopySink, buf string) {
	sink.WriteByte(StringType)
	sink.WriteUint32(uint32(len(buf)))
	sink.WriteBytes([]byte(buf))
}

func EncodeAddress(sink *common.ZeroCopySink, addr common.Address) {
	sink.WriteByte(AddressType)
	sink.WriteBytes(addr[:])
}

func EncodeBool(sink *common.ZeroCopySink, b bool) {
	sink.WriteByte(BooleanType)
	if b {
		sink.WriteByte(byte(1))
	} else {
		sink.WriteByte(byte(0))
	}
}

func EncodeH256(sink *common.ZeroCopySink, hash common.Uint256) {
	sink.WriteByte(H256Type)
	sink.WriteBytes(hash[:])
}

func EncodeInt128(sink *common.ZeroCopySink, val common.I128) {
	sink.WriteByte(IntType)
	sink.WriteBytes(val[:])
}

func EncodeBigInt(sink *common.ZeroCopySink, intval *big.Int) error {
	val, err := common.I128FromBigInt(intval)
	if err != nil {
		return err
	}
	EncodeInt128(sink, val)
	return nil
}

func EncodeList(sink *common.ZeroCopySink, list []interface{}) error {
	sink.WriteByte(ListType)
	sink.WriteUint32(uint32(len(list)))
	for _, elem := range list {
		switch val := elem.(type) {
		case []byte:
			EncodeBytes(sink, val)
		case string:
			EncodeString(sink, val)
		case bool:
			EncodeBool(sink, val)
		case int:
			EncodeInt128(sink, common.I128FromInt64(int64(val)))
		case int64:
			EncodeInt128(sink, common.I128FromInt64(val))
		case int32:
			EncodeInt128(sink, common.I128FromInt64(int64(val)))
		case uint32:
			EncodeInt128(sink, common.I128FromInt64(int64(val)))
		case *big.Int:
			err := EncodeBigInt(sink, val)
			if err != nil {
				return err
			}
		case common.Address:
			EncodeAddress(sink, val)
		case common.Uint256:
			EncodeH256(sink, val)
		case []interface{}:
			err := EncodeList(sink, val)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("encode list: unsupported type: %v", reflect.TypeOf(val).String())
		}
	}
	return nil
}
