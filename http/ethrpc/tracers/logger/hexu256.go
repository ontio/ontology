// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package logger

import (
	"encoding/json"
	"reflect"

	"github.com/holiman/uint256"
)

var u256T = reflect.TypeOf((*uint256.Int)(nil))

// U256 marshals/unmarshals as a JSON string with 0x prefix.
// The zero value marshals as "0x0".
type U256 uint256.Int

// MarshalText implements encoding.TextMarshaler
func (b U256) MarshalText() ([]byte, error) {
	u256 := (*uint256.Int)(&b)
	return []byte(u256.Hex()), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *U256) UnmarshalJSON(input []byte) error {
	// The uint256.Int.UnmarshalJSON method accepts "dec", "0xhex"; we must be
	// more strict, hence we check string and invoke SetFromHex directly.
	if !isString(input) {
		return errNonString(u256T)
	}
	// The hex decoder needs to accept empty string ("") as '0', which uint256.Int
	// would reject.
	if len(input) == 2 {
		(*uint256.Int)(b).Clear()
		return nil
	}
	val, err := uint256.FromHex(string(input[1 : len(input)-1]))
	*(*uint256.Int)(b) = *val
	if err != nil {
		return &json.UnmarshalTypeError{Value: err.Error(), Type: u256T}
	}
	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler
func (b *U256) UnmarshalText(input []byte) error {
	// The uint256.Int.UnmarshalText method accepts "dec", "0xhex"; we must be
	// more strict, hence we check string and invoke SetFromHex directly.
	val, err := uint256.FromHex(string(input))
	*(*uint256.Int)(b) = *val
	return err
}

// String returns the hex encoding of b.
func (b *U256) String() string {
	return (*uint256.Int)(b).Hex()
}

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

func errNonString(typ reflect.Type) error {
	return &json.UnmarshalTypeError{Value: "non-string", Type: typ}
}
