// Copyright (C) 2021 The Ontology Authors
// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package evm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"

	oComm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/http/ethrpc/utils"
)

func Test_event(t *testing.T) {
	txhash := "0xec56538d2cd67f585560a3769f0694e0b03354eb45258a4b2533cd2ac7cfbd74"
	fmt.Printf("%s", utils.EthToOntHash(common.HexToHash(txhash)).ToHexString())
}

func Test_deseiralizeLog(t *testing.T) {

	//transfer ong from : 0x96216849c49358b10257cb55b28ea603c874b05e to 0x4592d8f8d7b001e72cb26a73e4fa1806a51ac79d amount 10^9 (1 ong)
	states := "0x96216849c49358b10257cb55b28ea603c874b05e03000000ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef00000000000000000000000096216849c49358b10257cb55b28ea603c874b05e0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d043b9aca00"
	data, err := hexutil.Decode(states)
	assert.Nil(t, err)
	source := oComm.NewZeroCopySource(data)
	var storageLog types.StorageLog
	err = storageLog.Deserialization(source)
	assert.Nil(t, err)

	for _, t := range storageLog.Topics {
		fmt.Printf("%s\n", t.Hex())
	}
	d := big.NewInt(0).SetBytes(storageLog.Data)

	fmt.Printf("%d\n", d)

	assert.Equal(t, len(storageLog.Topics), 3)
	assert.Equal(t, storageLog.Topics[0].Hex(), "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	assert.Equal(t, storageLog.Topics[1].Hex(), "0x00000000000000000000000096216849c49358b10257cb55b28ea603c874b05e")
	assert.Equal(t, storageLog.Topics[2].Hex(), "0x0000000000000000000000004592d8f8d7b001e72cb26a73e4fa1806a51ac79d")

	assert.Equal(t, d.Int64(), int64(1000000000))
}
