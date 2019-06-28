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
	"github.com/ontio/ontology/common"
	"math/big"
)

const (
	ByteArrayType byte = 0x00
	AddressType   byte = 0x01
	BooleanType   byte = 0x02
	IntType       byte = 0x03
	H256Type      byte = 0x04
	//reserved for other types
	ListType byte = 0x10

	MAX_PARAM_LENGTH      = 1024
	VERSION          byte = 0
)

var ERROR_PARAM_FORMAT = fmt.Errorf("error param format")
var ERROR_PARAM_TOO_LONG = fmt.Errorf("param length is exceeded")
var ERROR_PARAM_NOT_SUPPORTED_TYPE = fmt.Errorf("error param format:not supported type")

//input byte array should be the following format
// version(1byte) + type(1byte) + usize( bytearray or list) (4 bytes) + data...

func DeserializeInput(input []byte) ([]interface{}, error) {
	if len(input) == 0 {
		return nil, ERROR_PARAM_FORMAT
	}
	if len(input) > MAX_PARAM_LENGTH {
		return nil, ERROR_PARAM_TOO_LONG
	}
	version := input[0]
	//current only support "0" version
	if version != VERSION {
		return nil, ERROR_PARAM_FORMAT
	}

	paramlist := make([]interface{}, 0)
	source := common.NewZeroCopySource(input[1:])
	for source.Len() != 0 {
		val, err := DecodeValue(source)
		if err != nil {
			return nil, err
		}
		paramlist = append(paramlist, val)
	}

	return paramlist, nil
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
	case AddressType:
		addr, eof := source.NextAddress()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		return addr, nil
	case BooleanType:
		by, eof := source.NextByte()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}

		return by != 0, nil
	case IntType:
		size, eof := source.NextUint32()
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}
		if size == 0 {
			return big.NewInt(0), nil
		}

		buf, eof := source.NextBytes(uint64(size))
		if eof {
			return nil, ERROR_PARAM_FORMAT
		}
		bi := common.BigIntFromNeoBytes(buf)
		return bi, nil
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
func EncodeInt(sink *common.ZeroCopySink, intval *big.Int) {
	sink.WriteByte(IntType)
	sink.WriteBytes(common.BigIntToNeoBytes(intval))
}

func EncodeList(sink *common.ZeroCopySink, list []interface{}) error {
	sink.WriteByte(ListType)
	sink.WriteUint32(uint32(len(list)))
	for _, elem := range list {
		switch elem.(type) {
		case []byte:
			EncodeBytes(sink, elem.([]byte))
		case string:
			EncodeBytes(sink, []byte(elem.(string)))
		case *big.Int:
			EncodeInt(sink, elem.(*big.Int))
		case common.Address:
			EncodeAddress(sink, elem.(common.Address))
		case common.Uint256:
			EncodeH256(sink, elem.(common.Uint256))
		case []interface{}:
			err := EncodeList(sink, elem.([]interface{}))
			if err != nil {
				return err
			}
		default:
			return ERROR_PARAM_NOT_SUPPORTED_TYPE
		}
	}
	return nil
}
