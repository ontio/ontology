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
package util

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestFloat32ToByte(t *testing.T) {
	f := float32(3.1415926)
	b := Float32ToBytes(f)

	if ByteToFloat32(b) != f {
		t.Error("TestFloat32ToByte failed!")
	}

}

func TestFloat64ToByte(t *testing.T) {
	f := 3.1415926
	b := Float64ToBytes(f)

	if ByteToFloat64(b) != f {
		t.Error("TestFloat64ToByte failed!")
	}
}

func TestInt32ToBytes(t *testing.T) {
	i := math.MaxInt32
	b := Int32ToBytes(uint32(i))
	if int(binary.LittleEndian.Uint32(b)) != i {
		t.Error("TestInt32ToBytes failed!")
	}
}

func TestInt64ToBytes(t *testing.T) {
	i := math.MaxInt64
	b := Int64ToBytes(uint64(i))
	if int(binary.LittleEndian.Uint64(b)) != i {
		t.Error("TestInt32ToBytes failed!")
	}
}

func TestTrimBuffToString(t *testing.T) {
	s := "helloworld"
	b := []byte(s)
	if TrimBuffToString(b) != s {
		t.Error("TestTrimBuffToString failed")
	}
	b = append(b, byte(0))
	if TrimBuffToString(b) != s {
		t.Error("TestTrimBuffToString failed")
	}
	b = append(b, []byte("some other string")...)
	if TrimBuffToString(b) != s {
		t.Error("TestTrimBuffToString failed")
	}
}
