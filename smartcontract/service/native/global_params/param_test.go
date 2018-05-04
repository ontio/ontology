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

package global_params

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/ontio/ontology/core/genesis"
	"github.com/stretchr/testify/assert"
)

func TestParams_Serialize_Deserialize(t *testing.T) {
	params := new(Params)
	*params = make(map[string]string)
	for i := 0; i < 10; i++ {
		k := "key" + strconv.Itoa(i)
		v := "value" + strconv.Itoa(i)
		(*params)[k] = v
	}
	bf := new(bytes.Buffer)
	if err := params.Serialize(bf); err != nil {
		t.Fatalf("params serialize error: %v", err)
	}
	deserializeParams := new(Params)
	if err := deserializeParams.Deserialize(bf); err != nil {
		t.Fatalf("params deserialize error: %v", err)
	}
	for i := 0; i < 10; i++ {
		k := "key" + strconv.Itoa(i)
		if (*params)[k] != (*deserializeParams)[k] {
			t.Fatal("params deserialize error")
		}
	}
}

func TestAdmin_Serialize_Deserialize(t *testing.T) {
	admin := new(Admin)
	copy((*admin)[:], genesis.ParamContractAddress[:])
	bf := new(bytes.Buffer)
	if err := admin.Serialize(bf); err != nil {
		t.Fatalf("admin serialize error: %v", err)
	}
	deserializeAdmin := new(Admin)
	if err := deserializeAdmin.Deserialize(bf); err != nil {
		t.Fatal("admin version deserialize error")
	}
	assert.Equal(t, admin, deserializeAdmin)
}
