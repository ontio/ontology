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
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func TestUnexportFields(t *testing.T) {
	unexport := struct {
		Name string
		num  uint
		Age  int
	}{
		Name: "aaa",
		num:  100,
		Age:  123,
	}
	export := struct {
		Name string
		Age  int
	}{
		Name: "aaa",
		Age:  123,
	}

	unexportCode, err := BuildNeoVMInvokeCode(common.Address{}, []interface{}{unexport})
	assert.Nil(t, err)
	exportCode, err := BuildNeoVMInvokeCode(common.Address{}, []interface{}{export})
	assert.Nil(t, err)
	assert.Equal(t, unexportCode, exportCode)
}
