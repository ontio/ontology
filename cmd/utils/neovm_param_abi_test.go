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
	"fmt"
	"testing"
)

func TestParseNeovmFunc(t *testing.T) {
	var testNeovmAbi = `{
  "hash": "0xe827bf96529b5780ad0702757b8bad315e2bb8ce",
  "entrypoint": "Main",
  "functions": [
    {
      "name": "Main",
      "parameters": [
        {
          "name": "operation",
          "type": "String"
        },
        {
          "name": "args",
          "type": "Array"
        }
      ],
      "returntype": "Any"
    },
    {
      "name": "Add",
      "parameters": [
        {
          "name": "a",
          "type": "Integer"
        },
        {
          "name": "b",
          "type": "Integer"
        }
      ],
      "returntype": "Integer"
    }
  ],
  "events": []
}`
	contractAbi, err := NewNeovmContractAbi([]byte(testNeovmAbi))
	if err != nil {
		t.Errorf("TestParseNeovmFunc NewNeovmContractAbi error:%s", err)
		return
	}
	funcAbi := contractAbi.GetFunc("Add")
	if funcAbi == nil {
		t.Error("TestParseNeovmFunc cannot find func abi")
		return
	}

	params, err := ParseNeovmFunc([]string{"12", "34"}, funcAbi)
	if err != nil {
		t.Errorf("TestParseNeovmFunc ParseNeovmFunc error:%s", err)
		return
	}
	fmt.Printf("TestParseNeovmFunc %v\n", params)
}
