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
	"errors"
	"fmt"
	"io"
)

const (
	UINT16_SIZE  = 2
	UINT32_SIZE  = 4
	UINT64_SIZE  = 8
	UINT256_SIZE = 32
)

type Uint256 [UINT256_SIZE]byte

var UINT256_EMPTY = Uint256{}

func (u *Uint256) ToArray() []byte {
	x := make([]byte, UINT256_SIZE)
	for i := 0; i < 32; i++ {
		x[i] = byte(u[i])
	}

	return x
}

func (u *Uint256) ToHexString() string {
	return fmt.Sprintf("%x", ToArrayReverse(u[:]))
}

func (u *Uint256) Serialize(w io.Writer) error {
	_, err := w.Write(u[:])
	return err
}

func (u *Uint256) Deserialize(r io.Reader) error {
	_, err := io.ReadFull(r, u[:])
	if err != nil {
		return errors.New("deserialize Uint256 error")
	}
	return nil
}

func Uint256ParseFromBytes(f []byte) (Uint256, error) {
	if len(f) != UINT256_SIZE {
		return Uint256{}, errors.New("[Common]: Uint256ParseFromBytes err, len != 32")
	}

	var hash Uint256
	copy(hash[:], f)
	return hash, nil
}

func Uint256FromHexString(s string) (Uint256, error) {
	hx, err := HexToBytes(s)
	if err != nil {
		return UINT256_EMPTY, err
	}
	return Uint256ParseFromBytes(ToArrayReverse(hx))
}
