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

func (r *VmReader) ReadByte() (byte, error) {
	byte, err := r.reader.ReadByte()
	return byte, err
}

func (r *VmReader) ReadBytes(count int) ([]byte, error) {
	b := make([]byte, count)
	_, err := r.reader.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *VmReader) ReadBytesInto(b []byte) error {
	_, err := r.reader.Read(b)
	if err != nil {
		return err
	}
	return nil
}

func (r *VmReader) ReadUint16() (uint16, error) {
	var b [2]byte
	err := r.ReadBytesInto(b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b[:]), nil
}

func (r *VmReader) ReadUint32() (uint32, error) {
	var b [4]byte
	err := r.ReadBytesInto(b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func (r *VmReader) ReadUint64() (uint64, error) {
	var b [8]byte
	err := r.ReadBytesInto(b[:])
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b[:]), nil
}

func (r *VmReader) ReadInt16() (int16, error) {
	val, err := r.ReadUint16()
	return int16(val), err
}

func (r *VmReader) ReadInt32() (int32, error) {
	val, err := r.ReadUint32()
	return int32(val), err
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

func (r *VmReader) ReadVarBytes(max uint32) ([]byte, error) {
	n, err := r.ReadVarInt(uint64(max))
	if err != nil {
		return nil, err
	}
	return r.ReadBytes(int(n))
}

func (r *VmReader) ReadVarInt(max uint64) (uint64, error) {
	fb, _ := r.ReadByte()
	var value uint64
	var err error

	switch fb {
	case 0xFD:
		val, e := r.ReadInt16()
		value = uint64(val)
		err = e
	case 0xFE:
		val, e := r.ReadUint32()
		value = uint64(val)
		err = e
	case 0xFF:
		val, e := r.ReadUint64()
		value = uint64(val)
		err = e
	default:
		value = uint64(fb)
	}
	if err != nil {
		return 0, err
	}
	if value > max {
		return 0, nil
	}
	return value, nil
}

func (r *VmReader) ReadVarString(maxlen uint32) (string, error) {
	bs, err := r.ReadVarBytes(maxlen)
	return string(bs), err
}
