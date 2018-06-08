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
	ctype "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTransactionBuilder(t *testing.T) {
	testCode := types.VmCode{Code: utils.AuthContractAddress[:], VmType: types.Native}
	deployTx := NewDeployTransaction(testCode, "AuthContract", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Authorization Contract", true)
	assert.NotNil(t, deployTx)
	assert.Equal(t, deployTx.TxType, ctype.Deploy)

	invokeTx := NewInvokeTransaction(testCode)
	assert.NotNil(t, invokeTx)
	assert.Equal(t, invokeTx.TxType, ctype.Invoke)
}
