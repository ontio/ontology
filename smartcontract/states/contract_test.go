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
package states

import (
	"bytes"
	"testing"

	"github.com/ontio/ontology/smartcontract/types"
)

func TestContract_Serialize_Deserialize(t *testing.T) {
	vmcode := types.VmCode{
		VmType: types.Native,
		Code:   []byte{1},
	}

	addr := vmcode.AddressFromVmCode()

	c := &Contract{
		Version: 0,
		Code:    []byte{1},
		Address: addr,
		Method:  "init",
		Args:    []byte{2},
	}
	bf := new(bytes.Buffer)
	if err := c.Serialize(bf); err != nil {
		t.Fatalf("Contract serialize error: %v", err)
	}

	v := new(Contract)
	if err := v.Deserialize(bf); err != nil {
		t.Fatalf("Contract deserialize error: %v", err)
	}
}
