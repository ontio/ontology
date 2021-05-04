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
package graphql

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ontio/ontology/v2/common"
)

type Uint32 uint32

func (Uint32) ImplementsGraphQLType(name string) bool {
	return name == "Uint32"
}

func (t *Uint32) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case uint32:
		*t = Uint32(input)
		return nil
	case int32:
		*t = Uint32(input)
		return nil
	case int64:
		*t = Uint32(input)
		return nil
	default:
		return fmt.Errorf("wrong type for Uint32: %T", input)
	}
}

func (t Uint32) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(t))
}

type Uint64 uint64

func (Uint64) ImplementsGraphQLType(name string) bool {
	return name == "Uint64"
}

func (t *Uint64) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case uint32:
		*t = Uint64(input)
	case int32:
		*t = Uint64(input)
	case int:
		*t = Uint64(input)
	case uint:
		*t = Uint64(input)
	case uint64:
		*t = Uint64(input)
	case int64:
		*t = Uint64(input)
	case string:
		val, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return fmt.Errorf("wrong type for Uint64: %T", input)
		}
		*t = Uint64(val)
	default:
		return fmt.Errorf("wrong type for Uint64: %T", input)
	}
	return nil
}

func (t Uint64) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(t), 10))
}

type Addr struct {
	common.Address
}

func (Addr) ImplementsGraphQLType(name string) bool {
	return name == "Address"
}

func (t *Addr) UnmarshalGraphQL(input interface{}) error {
	var err error
	switch input := input.(type) {
	case string:
		if strings.HasPrefix(input, "0x") {
			t.Address, err = common.AddressFromHexString(input[2:])
		}

		if err == nil {
			t.Address, err = common.AddressFromBase58(input)
		}
	default:
		return fmt.Errorf("wrong type for Address: %T", input)
	}
	return err
}

func (t Addr) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Address.ToBase58())
}

type H256 common.Uint256

func (H256) ImplementsGraphQLType(name string) bool {
	return name == "H256"
}

func (t *H256) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		if strings.HasPrefix(input, "0x") {
			input = input[2:]
		}

		hash, err := common.Uint256FromHexString(input)
		if err != nil {
			return err
		}
		*t = H256(hash)
		return nil
	default:
		return fmt.Errorf("wrong type for H256: %T", input)
	}
}

func (t H256) MarshalJSON() ([]byte, error) {
	hash := common.Uint256(t)
	return json.Marshal(hash.ToHexString())
}

type PubKey string

func (PubKey) ImplementsGraphQLType(name string) bool {
	return name == "PubKey"
}

func (key *PubKey) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case string:
		_, err := common.PubKeyFromHex(input)
		if err != nil {
			return err
		}
		*key = PubKey(input)
		return nil
	default:
		return fmt.Errorf("wrong type for PubKey: %T", input)
	}
}

func (key PubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(key))
}
