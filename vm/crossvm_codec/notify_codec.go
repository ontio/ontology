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
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

func DeserializeNotify(input []byte) interface{} {
	val, err := parseNotify(input)
	if err != nil {
		return input
	}

	return stringify(val)
}

func stringify(notify interface{}) interface{} {
	switch val := notify.(type) {
	case []byte:
		return hex.EncodeToString(val)
	case common.Address:
		return val.ToBase58()
	case bool, string:
		return val
	case common.Uint256:
		return val.ToHexString()
	case *big.Int:
		return fmt.Sprintf("%d", val)
	case []interface{}:
		list := make([]interface{}, 0, len(val))
		for _, v := range val {
			list = append(list, stringify(v))
		}
		return list
	default:
		log.Warn("notify codec: unsupported type:", reflect.TypeOf(val).String())

		return val
	}
}

// input byte array should be the following format
// evt\0(4byte) + type(1byte) + usize( bytearray or list) (4 bytes) + data...
func parseNotify(input []byte) (interface{}, error) {
	if bytes.HasPrefix(input, []byte("evt\x00")) == false {
		return nil, ERROR_PARAM_FORMAT
	}

	source := common.NewZeroCopySource(input[4:])

	return DecodeValue(source)
}
