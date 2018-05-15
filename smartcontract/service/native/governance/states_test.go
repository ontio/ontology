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

package governance

import (
	"bytes"
	"testing"
)

func TestSerialization_Status(t *testing.T) {
	var param1 Status = 10
	bf := new(bytes.Buffer)
	if err := param1.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	var param2 Status
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if param1 != param2 {
		t.Fatalf("TestSerialization_Status failed")
	}
}

func TestSerialization_VoteCommitDposParam(t *testing.T) {
	param1 := &VoteCommitDposParam{
		Address: "aaaa5e502c2c72eb6edaa9516735d518f09c95c3",
		Pos:     -1000,
	}
	bf := new(bytes.Buffer)
	if err := param1.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())

	param2 := new(VoteCommitDposParam)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if param1.Address != param2.Address || param1.Pos != param2.Pos {
		t.Fatalf("TestSerialization_VoteCommitDposParam failed")
	}
}
