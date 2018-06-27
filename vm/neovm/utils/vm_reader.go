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

package utils

import (
	"bytes"
	"encoding/binary"
)

type VmReader struct {
	reader     *bytes.Reader
	BaseStream []byte
}

func NewVmReader(b []byte) *VmReader {
	var vmreader VmReader
	vmreader.reader = bytes.NewReader(b)
	vmreader.BaseStream = b
	return &vmreader
}

func (r *VmReader) Reader() *bytes.Reader {
	return r.reader
}

func (r *VmReader) ReadByte() (byte, error) {
	byte, err := r.reader.ReadByte()
	return byte, err
}

func (r *VmReader) ReadBytes(count int) []byte {
	b := make([]byte, count)
	r.reader.Read(b)
	return b
}

func (r *VmReader) ReadUint16() uint16 {
	b := r.ReadBytes(2)
	return binary.LittleEndian.Uint16(b)
}

func (r *VmReader) ReadUInt32() uint32 {
	b := r.ReadBytes(4)
	return binary.LittleEndian.Uint32(b)
}

func (r *VmReader) ReadUInt64() uint64 {
	b := r.ReadBytes(8)
	return binary.LittleEndian.Uint64(b)
}

func (r *VmReader) ReadInt16() int16 {
	b := r.ReadBytes(2)
	bytesBuffer := bytes.NewBuffer(b)
	var vi int16
	binary.Read(bytesBuffer, binary.LittleEndian, &vi)
	return vi

}

func (r *VmReader) ReadInt32() int32 {
	b := r.ReadBytes(4)
	bytesBuffer := bytes.NewBuffer(b)
	var vi int32
	binary.Read(bytesBuffer, binary.LittleEndian, &vi)
	return vi
}

func (r *VmReader) Position() int {
	return int(r.reader.Size()) - r.reader.Len()
}

func (r *VmReader) Length() int {
	return r.reader.Len()
}

func (r *VmReader) Seek(offset int64, whence int) (int64, error) {
	return r.reader.Seek(offset, whence)
}

func (r *VmReader) ReadVarBytes(max uint32) []byte {
	n := int(r.ReadVarInt(uint64(max)))
	return r.ReadBytes(n)
}

func (r *VmReader) ReadVarInt(max uint64) uint64 {
	fb, _ := r.ReadByte()
	var value uint64

	switch fb {
	case 0xFD:
		value = uint64(r.ReadInt16())
	case 0xFE:
		value = uint64(r.ReadUInt32())
	case 0xFF:
		value = uint64(r.ReadUInt64())
	default:
		value = uint64(fb)
	}
	if value > max {
		return 0
	}
	return value
}

func (r *VmReader) ReadVarString(maxlen uint32) string {
	bs := r.ReadVarBytes(maxlen)
	return string(bs)
}
