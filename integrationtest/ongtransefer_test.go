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
package integrationtest

import (
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	types2 "github.com/ethereum/go-ethereum/core/types"
	oComm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/integrationtest/erc20"
	"github.com/stretchr/testify/assert"
)

func TestNewERC20(t *testing.T) {
	states := "0x000000000000000000000000000000000000000203000000ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef00000000000000000000000070997970c51812dc3a010c7d01b50e0d17dc79c8000000000000000000000000000000000000000000000000000000000000000720000000000000000000000000000000000000000000000000000000000113e4e8"
	data, err := hexutil.Decode(states)
	assert.Nil(t, err)
	source := oComm.NewZeroCopySource(data)
	var storageLog types.StorageLog
	err = storageLog.Deserialization(source)
	assert.Nil(t, err)
	parsed, _ := abi.JSON(strings.NewReader(erc20.ERC20ABI))
	nbc := bind.NewBoundContract(common.Address{}, parsed, nil, nil, nil)
	tf := new(erc20.ERC20Transfer)
	l := types2.Log{
		Address: storageLog.Address,
		Topics:  storageLog.Topics,
		Data:    storageLog.Data,
	}
	err = nbc.UnpackLog(tf, "Transfer", l)
	assert.Nil(t, err)
	assert.Equal(t, "0x70997970C51812dc3A010C7d01b50e0d17dc79C8", tf.From.String())
	assert.Equal(t, "0x0000000000000000000000000000000000000007", tf.To.String())
	assert.Equal(t, "18081000", tf.Value.String())
}
