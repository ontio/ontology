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
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/itchyny/base58-go"
)

const ADDR_LEN = 20
var DEFALUT_ADDR = [20]byte{}

type Address [ADDR_LEN]byte

// ToHexString returns  hex string representation of Address
func (self *Address) ToHexString() string {
	return fmt.Sprintf("%x", self[:])
}

// Serialize serialize Address into io.Writer
func (self *Address) Serialize(w io.Writer) error {
	_, err := w.Write(self[:])
	return err
}

// Deserialize deserialize Address from io.Reader
func (self *Address) Deserialize(r io.Reader) error {
	n, err := r.Read(self[:])
	if n != len(self[:]) || err != nil {
		return errors.New("deserialize Address error")
	}
	return nil
}

// ToBase58 returns base58 encoded address string
func (f *Address) ToBase58() string {
	data := append([]byte{0x41}, f[:]...)
	temp := sha256.Sum256(data)
	temps := sha256.Sum256(temp[:])
	data = append(data, temps[0:4]...)

	bi := new(big.Int).SetBytes(data).String()
	encoded, _ := base58.BitcoinEncoding.Encode([]byte(bi))
	return string(encoded)
}

// AddressParseFromBytes returns parsed Address
func AddressParseFromBytes(f []byte) (Address, error) {
	if len(f) != ADDR_LEN {
		return Address{}, errors.New("[Common]: Uint160ParseFromBytes err, len != 20")
	}

	var addr Address
	copy(addr[:], f)
	return addr, nil
}

// AddressFromBase58 returns Address from encoded base58 string
func AddressFromBase58(encoded string) (Address, error) {
	decoded, err := base58.BitcoinEncoding.Decode([]byte(encoded))
	if err != nil {
		return Address{}, err
	}

	x, _ := new(big.Int).SetString(string(decoded), 10)

	buf := x.Bytes()
	if len(buf) != 1+ADDR_LEN+4 || buf[0] != byte(0x41) {
		return Address{}, errors.New("wrong encoded address")
	}
	ph, err := AddressParseFromBytes(buf[1:21])
	if err != nil {
		return Address{}, err
	}

	addr := ph.ToBase58()

	if addr != encoded {
		return Address{}, errors.New("[AddressFromBase58]: decode encoded verify failed.")
	}

	return ph, nil
}
