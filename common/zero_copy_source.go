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
	"errors"
)

var ErrIrregularData = errors.New("irregular data")

type ZeroCopySource struct {
	s   []byte
	off uint64 // current reading index
}

// Len returns the number of bytes of the unread portion of the
// slice.
func (self *ZeroCopySource) Len() uint64 {
	length := uint64(len(self.s))
	if self.off >= length {
		return 0
	}
	return length - self.off
}

func (self *ZeroCopySource) Pos() uint64 {
	return self.off
}

// Size returns the original length of the underlying byte slice.
// Size is the number of bytes available for reading via ReadAt.
// The returned value is always the same and is not affected by calls
// to any other method.
func (self *ZeroCopySource) Size() uint64 { return uint64(len(self.s)) }

// Read implements the io.ZeroCopySource interface.
func (self *ZeroCopySource) NextBytes(n uint64) (data []byte, eof bool) {
	m := uint64(len(self.s))
	end, overflow := SafeAdd(self.off, n)
	if overflow || end > m {
		end = m
		eof = true
	}
	data = self.s[self.off:end]
	self.off = end

	return
}

func (self *ZeroCopySource) Skip(n uint64) (eof bool) {
	m := uint64(len(self.s))
	end, overflow := SafeAdd(self.off, n)
	if overflow || end > m {
		end = m
		eof = true
	}
	self.off = end

	return
}

// ReadByte implements the io.ByteReader interface.
func (self *ZeroCopySource) NextByte() (data byte, eof bool) {
	if self.off >= uint64(len(self.s)) {
		return 0, true
	}

	b := self.s[self.off]
	self.off++
	return b, false
}

func (self *ZeroCopySource) NextUint8() (data uint8, eof bool) {
	var val byte
	val, eof = self.NextByte()
	return uint8(val), eof
}

func (self *ZeroCopySource) NextBool() (data bool, irregular bool, eof bool) {
	val, eof := self.NextByte()
	if val == 0 {
		data = false
	} else if val == 1 {
		data = true
	} else {
		data = true
		irregular = true
	}

	return
}

// Backs up a number of bytes, so that the next call to NextXXX() returns data again
// that was already returned by the last call to NextXXX().
func (self *ZeroCopySource) BackUp(n uint64) {
	self.off -= n
}

func (self *ZeroCopySource) NextUint16() (data uint16, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(UINT16_SIZE)
	if eof {
		return
	}

	return binary.LittleEndian.Uint16(buf), eof
}

func (self *ZeroCopySource) NextUint32() (data uint32, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(UINT32_SIZE)
	if eof {
		return
	}

	return binary.LittleEndian.Uint32(buf), eof
}

func (self *ZeroCopySource) NextUint64() (data uint64, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(UINT64_SIZE)
	if eof {
		return
	}

	return binary.LittleEndian.Uint64(buf), eof
}

func (self *ZeroCopySource) NextInt32() (data int32, eof bool) {
	var val uint32
	val, eof = self.NextUint32()
	return int32(val), eof
}

func (self *ZeroCopySource) NextInt64() (data int64, eof bool) {
	var val uint64
	val, eof = self.NextUint64()
	return int64(val), eof
}

func (self *ZeroCopySource) NextInt16() (data int16, eof bool) {
	var val uint16
	val, eof = self.NextUint16()
	return int16(val), eof
}

func (self *ZeroCopySource) NextVarBytes() (data []byte, size uint64, irregular bool, eof bool) {
	var count uint64
	count, size, irregular, eof = self.NextVarUint()
	size += count

	if count > 0 {
		data, eof = self.NextBytes(count)
	}

	return
}

func (self *ZeroCopySource) NextAddress() (data Address, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(ADDR_LEN)
	if eof {
		return
	}
	copy(data[:], buf)

	return
}

func (self *ZeroCopySource) NextHash() (data Uint256, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(UINT256_SIZE)
	if eof {
		return
	}
	copy(data[:], buf)

	return
}

func (self *ZeroCopySource) NextString() (data string, size uint64, irregular bool, eof bool) {
	var val []byte
	val, size, irregular, eof = self.NextVarBytes()
	data = string(val)
	return
}

func (self *ZeroCopySource) NextVarUint() (data uint64, size uint64, irregular bool, eof bool) {
	var fb byte
	fb, eof = self.NextByte()
	if eof {
		return
	}

	switch fb {
	case 0xFD:
		val, e := self.NextUint16()
		if e {
			eof = e
			return
		}
		data = uint64(val)
		size = 3
	case 0xFE:
		val, e := self.NextUint32()
		if e {
			eof = e
			return
		}
		data = uint64(val)
		size = 5
	case 0xFF:
		val, e := self.NextUint64()
		if e {
			eof = e
			return
		}
		data = uint64(val)
		size = 9
	default:
		data = uint64(fb)
		size = 1
	}

	irregular = size != getVarUintSize(data)

	return
}

func getVarUintSize(value uint64) uint64 {
	if value < 0xfd {
		return 1
	} else if value <= 0xffff {
		return 3
	} else if value <= 0xFFFFFFFF {
		return 5
	} else {
		return 9
	}
}

// NewReader returns a new ZeroCopySource reading from b.
func NewZeroCopySource(b []byte) *ZeroCopySource { return &ZeroCopySource{b, 0} }
