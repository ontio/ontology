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
package payload

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeployCode(t *testing.T) {
	deploy := DeployCode{
		Code:        []byte{1, 2, 3},
		NeedStorage: true,
		Name:        "test",
		Version:     "1.0.0",
		Author:      "ontology",
		Email:       "1@1.com",
		Description: "test",
	}

	buf := bytes.NewBuffer(nil)
	err := deploy.Serialize(buf)
	if err != nil {
		t.Fatal(err)
	}
	bs := buf.Bytes()
	var deploy2 DeployCode
	err = deploy2.Deserialize(buf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, deploy2, deploy)

	buf = bytes.NewBuffer(bs)
	err = deploy2.Deserialize(buf)
	if err != nil {
		t.Fatal(err)
	}
}
