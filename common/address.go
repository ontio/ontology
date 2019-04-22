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
	"golang.org/x/crypto/ripemd160"
)

const ADDR_LEN = 20

type Address [ADDR_LEN]byte

var ADDRESS_EMPTY = Address{}

// ToHexString returns  hex string representation of Address
func (self *Address) ToHexString() string {
	return fmt.Sprintf("%x", ToArrayReverse(self[:]))
}

// Serialize serialize Address into io.Writer
func (self *Address) Serialize(w io.Writer) error {
	_, err := w.Write(self[:])
	return err
}

// Deserialize deserialize Address from io.Reader
func (self *Address) Deserialize(r io.Reader) error {
	_, err := io.ReadFull(r, self[:])
	if err != nil {
		return errors.New("deserialize Address error")
	}
	return nil
}

// ToBase58 returns base58 encoded address string
func (f *Address) ToBase58() string {
	data := append([]byte{23}, f[:]...)
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
		return ADDRESS_EMPTY, errors.New("[Common]: AddressParseFromBytes err, len != 20")
	}

	var addr Address
	copy(addr[:], f)
	return addr, nil
}

// AddressParseFromHexString returns parsed Address
func AddressFromHexString(s string) (Address, error) {
	hx, err := HexToBytes(s)
	if err != nil {
		return ADDRESS_EMPTY, err
	}
	return AddressParseFromBytes(ToArrayReverse(hx))
}

const MaxBase58AddrLen = 2048 // just to avoid dos
// AddressFromBase58 returns Address from encoded base58 string
func AddressFromBase58(encoded string) (Address, error) {
	if encoded == "" || len(encoded) > MaxBase58AddrLen {
		return ADDRESS_EMPTY, errors.New("invalid address")
	}
	decoded, err := base58.BitcoinEncoding.Decode([]byte(encoded))
	if err != nil {
		return ADDRESS_EMPTY, err
	}

	x, ok := new(big.Int).SetString(string(decoded), 10)
	if !ok {
		return ADDRESS_EMPTY, errors.New("invalid address")
	}

	buf := x.Bytes()
	if len(buf) != 1+ADDR_LEN+4 || buf[0] != byte(23) {
		return ADDRESS_EMPTY, errors.New("wrong encoded address")
	}

	ph, err := AddressParseFromBytes(buf[1:21])
	if err != nil {
		return ADDRESS_EMPTY, err
	}

	addr := ph.ToBase58()

	if addr != encoded {
		return ADDRESS_EMPTY, errors.New("[AddressFromBase58]: decode encoded verify failed.")
	}

	return ph, nil
}

func AddressFromVmCode(code []byte) Address {
	var addr Address
	temp := sha256.Sum256(code)
	md := ripemd160.New()
	md.Write(temp[:])
	md.Sum(addr[:0])

	return addr
}
