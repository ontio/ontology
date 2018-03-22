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

package rlp

import (
	"io"
	"reflect"
)

type RawValue []byte

var rawValueType = reflect.TypeOf(RawValue{})

type Kind int

const (
	Byte Kind = iota
	String
	List
)

func SplitList(b []byte) ([]byte, []byte, error) {
	k, content, rest, err := Split(b)
	if err != nil {
		return nil, b, err
	}
	if k != List {
		return nil, b, ErrExpectedList
	}
	return content, rest, nil
}

func SplitString(b []byte) ([]byte, []byte, error) {
	k, content, rest, err := Split(b)
	if err != nil {
		return nil, b, err
	}
	if k == List {
		return nil, b, ErrExpectedString
	}
	return content, rest, nil
}

func Split(b []byte) (Kind, []byte, []byte, error) {
	k, ts, cs, err := readKind(b)
	if err != nil {
		return 0, nil, b, err
	}
	return k, b[ts: ts + cs], b[ts + cs:], nil
}

func CountValues(b []byte) (int, error) {
	i := 0
	for ; len(b) > 0; i++ {
		_, tagSize, size, err := readKind(b)
		if err != nil {
			return 0, err
		}
		b = b[tagSize + size:]
	}
	return i, nil
}

func readKind(buf []byte) (k Kind, tagSize, contentSize uint64, err error) {
	if len(buf) == 0 {
		return 0, 0, 0, io.ErrUnexpectedEOF
	}
	b := buf[0]
	switch {
	case b < 0x80:
		k = Byte
		tagSize = 0
		contentSize = 1
	case b < 0xB8:
		k = String
		tagSize = 1
		contentSize = uint64(b - 0x80)
		if contentSize == 1 && buf[1] < 128 {
			return 0, 0, 0, ErrCanonSize
		}
	case b < 0xC0:
		k = String
		tagSize = uint64(b - 0xB7) + 1
		contentSize, err = readSize(buf[1:], b - 0xB7)
	case b < 0xF8:
		k = List
		tagSize = 1
		contentSize = uint64(b - 0xC0)
	default:
		k = List
		tagSize = uint64(b - 0xF7) + 1
		contentSize, err = readSize(buf[1:], b - 0xF7)
	}
	if err != nil {
		return 0, 0, 0, err
	}
	if contentSize > uint64(len(buf)) - tagSize {
		return 0, 0, 0, ErrValueTooLarge
	}
	return k, tagSize, contentSize, err
}

func readSize(b []byte, sLen byte) (uint64, error) {
	if int(sLen) > len(b) {
		return 0, io.ErrUnexpectedEOF
	}
	var s uint64
	switch sLen {
	case 1:
		s = uint64(b[0])
	case 2:
		s = uint64(b[0]) << 8 | uint64(b[1])
	case 3:
		s = uint64(b[0]) << 16 | uint64(b[1]) << 8 | uint64(b[2])
	case 4:
		s = uint64(b[0]) << 24 | uint64(b[1]) << 16 | uint64(b[2]) << 8 | uint64(b[3])
	case 5:
		s = uint64(b[0]) << 32 | uint64(b[1]) << 24 | uint64(b[2]) << 16 | uint64(b[3]) << 8 | uint64(b[4])
	case 6:
		s = uint64(b[0]) << 40 | uint64(b[1]) << 32 | uint64(b[2]) << 24 | uint64(b[3]) << 16 | uint64(b[4]) << 8 | uint64(b[5])
	case 7:
		s = uint64(b[0]) << 48 | uint64(b[1]) << 40 | uint64(b[2]) << 32 | uint64(b[3]) << 24 | uint64(b[4]) << 16 | uint64(b[5]) << 8 | uint64(b[6])
	case 8:
		s = uint64(b[0]) << 56 | uint64(b[1]) << 48 | uint64(b[2]) << 40 | uint64(b[3]) << 32 | uint64(b[4]) << 24 | uint64(b[5]) << 16 | uint64(b[6]) << 8 | uint64(b[7])
	}
	if s < 56 || b[0] == 0 {
		return 0, ErrCanonSize
	}
	return s, nil
}