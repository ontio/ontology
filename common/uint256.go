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
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const UINT256_SIZE = 32

type Uint256 [UINT256_SIZE]byte

var UINT256_EMPTY  = Uint256{}

func (u *Uint256) CompareTo(o Uint256) int {
	x := u.ToArray()
	y := o.ToArray()

	for i := len(x) - 1; i >= 0; i-- {
		if x[i] > y[i] {
			return 1
		}
		if x[i] < y[i] {
			return -1
		}
	}

	return 0
}

func (u *Uint256) ToArray() []byte {
	x := make([]byte, UINT256_SIZE)
	for i := 0; i < 32; i++ {
		x[i] = byte(u[i])
	}

	return x
}

func (u *Uint256) Serialize(w io.Writer) (int, error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, u)

	len, err := w.Write(buf.Bytes())

	if err != nil {
		return 0, err
	}

	return len, nil
}

func (u *Uint256) Deserialize(r io.Reader) error {
	p := make([]byte, UINT256_SIZE)
	n, err := r.Read(p)

	if n <= 0 || err != nil {
		return err
	}

	buf := bytes.NewBuffer(p)
	binary.Read(buf, binary.LittleEndian, u)

	return nil
}

func (u *Uint256) ToString() string {
	return string(u.ToArray())
}

func Uint256ParseFromBytes(f []byte) (Uint256, error) {
	if len(f) != UINT256_SIZE {
		return Uint256{}, errors.New("[Common]: Uint256ParseFromBytes err, len != 32")
	}

	var hash [32]uint8
	for i := 0; i < 32; i++ {
		hash[i] = f[i]
	}
	return Uint256(hash), nil
}
