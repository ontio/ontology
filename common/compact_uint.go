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

package common

import (
	"encoding/binary"
	"fmt"
)

type CompactUint struct {
	cptUint []byte
}

func GetCompactUint(buf []byte) (uint64, uint8) {
	var val uint64
	var size uint8

	fmt.Printf("The buffer len is %d\n", len(buf))
	len := uint64(buf[0])
	if len < 0xfd {
		val = len
		size = 1
	} else if len == 0xfd {
		val = binary.LittleEndian.Uint64(buf[1:3])
		size = 3
	} else if len == 0xfe {
		val = binary.LittleEndian.Uint64(buf[1:5])
		size = 5
	} else if len == 0xff {
		val = binary.LittleEndian.Uint64(buf[1:9])
		size = 9
	}
	return val, size
}

// TODO Fix the return value to the correct number
func SetCompactUint(num uint64) []byte {
	var buf []byte
	if num <= 0xff {
		buf = make([]byte, 1)
		buf[0] = uint8(num)
	} else if num <= 0xffff {
		buf = make([]byte, 3)
		buf[0] = 0xfd
		binary.LittleEndian.PutUint16(buf[1:], uint16(num))
	} else if num <= 0xffffffff {
		buf = make([]byte, 5)
		buf[0] = 0xfe
		binary.LittleEndian.PutUint32(buf[1:], uint32(num))
	} else if num <= 0xffffffffffffffff {
		buf = make([]byte, 9)
		buf[0] = 0xff
		binary.LittleEndian.PutUint64(buf[1:], uint64(num))
	}

	return buf
}
