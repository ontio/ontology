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
	"io"
)

type ZeroCopyReader struct {
	Source *ZeroCopySource
	err    error
}

func (self *ZeroCopyReader) Error() error {
	return self.err
}

// Len returns the number of bytes of the unread portion of the
// slice.
func (self *ZeroCopyReader) Len() uint64 {
	return self.Source.Len()
}

func (self *ZeroCopyReader) Pos() uint64 {
	return self.Source.Pos()
}

// Size returns the original length of the underlying byte slice.
// Size is the number of bytes available for reading via ReadAt.
// The returned value is always the same and is not affected by calls
// to any other method.
func (self *ZeroCopyReader) Size() uint64 { return self.Source.Size() }

// Read implements the io.ZeroCopyReader interface.
func (self *ZeroCopyReader) ReadBytes(n uint64) (data []byte) {
	if self.err != nil {
		return
	}
	data, self.err = self.Source.ReadBytes(n)
	return
}

func (self *ZeroCopyReader) Skip(n uint64) {
	if self.err != nil {
		return
	}
	eof := self.Source.Skip(n)
	if eof {
		self.err = io.ErrUnexpectedEOF
	}
	return
}

func (self *ZeroCopyReader) ReadUint8() (data uint8) {
	if self.err != nil {
		return
	}
	data, self.err = self.Source.ReadByte()
	return
}

func (self *ZeroCopyReader) ReadBool() (data bool) {
	if self.err != nil {
		return
	}
	data, self.err = self.Source.ReadBool()
	return
}

// Backs up a number of bytes, so that the next call to NextXXX() returns data again
// that was already returned by the last call to NextXXX().
func (self *ZeroCopyReader) BackUp(n uint64) {
	self.Source.BackUp(n)
}

func (self *ZeroCopyReader) ReadUint32() (data uint32) {
	if self.err != nil {
		return
	}
	data, self.err = self.Source.ReadUint32()
	return
}

func (self *ZeroCopyReader) ReadUint64() (data uint64) {
	if self.err != nil {
		return
	}
	data, self.err = self.Source.ReadUint64()
	return
}

func (self *ZeroCopyReader) ReadInt32() (data int32) {
	return int32(self.ReadUint32())
}

func (self *ZeroCopyReader) ReadInt64() (data int64) {
	return int64(self.ReadUint64())
}

func (self *ZeroCopyReader) ReadString() (data string) {
	if self.err != nil {
		return
	}
	data, self.err = self.Source.ReadString()
	return
}

func (self *ZeroCopyReader) ReadVarBytes() (data []byte) {
	if self.err != nil {
		return nil
	}
	data, self.err = self.Source.ReadVarBytes()
	return data
}

func (self *ZeroCopyReader) ReadAddress() (addr Address) {
	if self.err != nil {
		return
	}
	addr, self.err = self.Source.ReadAddress()
	return
}

func (self *ZeroCopyReader) ReadHash() (hash Uint256) {
	if self.err != nil {
		return
	}
	hash, self.err = self.Source.ReadHash()
	return
}

func (self *ZeroCopyReader) ReadVarUint() (data uint64) {
	if self.err != nil {
		return
	}
	data, self.err = self.Source.ReadVarUint()
	return
}

func NewZeroCopyReader(b []byte) *ZeroCopyReader {
	return &ZeroCopyReader{
		Source: NewZeroCopySource(b),
	}
}
