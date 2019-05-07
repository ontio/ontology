/*
 * Copyright (C) 2019 The ontology Authors
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
package oep4

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"math/big"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/stretchr/testify/assert"
)

func TestRegisterParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &RegisterParam{
		Name:        "TestToken",
		Symbol:      "TT",
		Decimals:    12,
		TotalSupply: big.NewInt(1000000000),
		Account:     acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &RegisterParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestMigrateParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &MigrateParam{
		NewAsset: acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &MigrateParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestBalanceParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &BalanceParam{
		User: acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &BalanceParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestAllowanceParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &AllowanceParam{
		Owner:   acc.Address,
		Spender: acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &AllowanceParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestTransferParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &TransferParam{
		From:   acc.Address,
		To:     acc.Address,
		Amount: big.NewInt(100),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &TransferParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestMultiTransferParam(t *testing.T) {
	acc := account.NewAccount("")
	tranParam := &TransferParam{
		From:   acc.Address,
		To:     acc.Address,
		Amount: big.NewInt(100),
	}
	param := &MultiTransferParam{
		Transfers: []*TransferParam{tranParam, tranParam},
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &MultiTransferParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestApproveParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &ApproveParam{
		Owner:     acc.Address,
		Spender:   acc.Address,
		Allowance: big.NewInt(100),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &ApproveParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestTransferFromParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &TransferFromParam{
		Spender: acc.Address,
		From:    acc.Address,
		To:      acc.Address,
		Amount:  big.NewInt(100),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &TransferFromParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestXShardTransferParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &XShardTransferParam{
		ToShard: common.NewShardIDUnchecked(22),
		From:    acc.Address,
		To:      acc.Address,
		Amount:  big.NewInt(100),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &XShardTransferParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestXShardTransferRetryParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &XShardTransferRetryParam{
		From:       acc.Address,
		TransferId: big.NewInt(19),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &XShardTransferRetryParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestShardMintParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &ShardMintParam{
		Asset:       19,
		Account:     acc.Address,
		FromShard:   common.NewShardIDUnchecked(9),
		FromAccount: acc.Address,
		TransferId:  big.NewInt(19),
		Amount:      big.NewInt(19999),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &ShardMintParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestXShardTranSuccParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &XShardTranSuccParam{
		Asset:      19,
		Account:    acc.Address,
		TransferId: big.NewInt(19),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &XShardTranSuccParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestGetXShardTransferInfoParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &GetXShardTransferInfoParam{
		Asset:      19,
		Account:    acc.Address,
		TransferId: big.NewInt(19),
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &GetXShardTransferInfoParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}

func TestGetPendingXShardTransferParam(t *testing.T) {
	acc := account.NewAccount("")
	param := &GetPendingXShardTransferParam{
		Asset:   19,
		Account: acc.Address,
	}
	bf := new(bytes.Buffer)
	err := param.Serialize(bf)
	assert.Nil(t, err)
	newParam := &GetPendingXShardTransferParam{}
	err = newParam.Deserialize(bf)
	assert.Nil(t, err)
	assert.Equal(t, param, newParam)
}
