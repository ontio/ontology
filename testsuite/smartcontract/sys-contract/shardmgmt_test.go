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

package TestContracts

import (
	"fmt"
	"os"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/chainmgr"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/shardasset/oep4"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	shardstates "github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/testsuite"
	"github.com/ontio/ontology/testsuite/common"
	tutils "github.com/ontio/ontology/testsuite/utils"
	"github.com/stretchr/testify/assert"
)

func init() {
	TestConsts.TestRootDir = "../../"
}

func TestClearData(t *testing.T) {
	err := os.RemoveAll(TestConsts.TestRootDir + "Chain/")
	if err != nil {
		t.Fatal(err)
	}
	err = os.RemoveAll(TestConsts.TestRootDir + "wallets/")
	if err != nil {
		t.Fatal(err)
	}
}

func Test_ShardMgmtInit(t *testing.T) {

	// 1. create root chain
	shardID := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	TestCommon.CreateChain(t, "test", shardID, 0)

	// 2. build shard-mgmt init tx

	tx := TestCommon.CreateAdminTx(t, shardID, 0, utils.ShardMgmtContractAddress, shardmgmt.INIT_NAME, nil)

	// 3. create new block
	blk := TestCommon.CreateBlock(t, ledger.GetShardLedger(shardID), []*types.Transaction{tx})

	// 4. add block
	TestCommon.ExecBlock(t, shardID, blk)
	TestCommon.SubmitBlock(t, shardID, blk)

	// 5. query db
	state := TestCommon.GetShardStateFromLedger(t, ledger.GetShardLedger(shardID), shardID)
	fmt.Printf("%v", state)
}

func TestStartShard(t *testing.T) {
	rootShardId := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	rootLedger := ledger.GetShardLedger(rootShardId)
	if rootLedger == nil {
		TestCommon.CreateChain(t, "root", rootShardId, 0)
		assetInitTx := TestCommon.CreateNativeTx(t, TestCommon.GetOwnerName(rootShardId, 0), 0,
			utils.ShardAssetAddress, oep4.INIT, nil)
		mgmtInitTx := TestCommon.CreateAdminTx(t, rootShardId, 0, utils.ShardMgmtContractAddress,
			shardmgmt.INIT_NAME, nil)
		initBlock := TestCommon.CreateBlock(t, ledger.GetShardLedger(rootShardId),
			[]*types.Transaction{mgmtInitTx, assetInitTx})
		TestCommon.ExecBlock(t, rootShardId, initBlock)
		TestCommon.SubmitBlock(t, rootShardId, initBlock)
		rootLedger = ledger.GetShardLedger(rootShardId)
	}

	globalShardState, err := chainmgr.GetShardMgmtGlobalState(rootLedger)
	if err != nil {
		t.Fatal(err)
	}
	nextShard := globalShardState.NextSubShardIndex
	newShardId := common.NewShardIDUnchecked(uint64(nextShard))
	assert.True(t, newShardId.ParentID() == rootShardId)

	initAssetBlock := tutils.GenInitShardAssetBlock(t)
	TestCommon.ExecBlock(t, rootShardId, initAssetBlock)
	TestCommon.SubmitBlock(t, rootShardId, initAssetBlock)

	creatorName := TestCommon.GetUserName(rootShardId, 1)
	shardBlock := tutils.GenRunShardBlock(t, rootShardId, newShardId, creatorName)
	TestCommon.ExecBlock(t, rootShardId, shardBlock)
	TestCommon.SubmitBlock(t, rootShardId, shardBlock)

	newShard := TestCommon.GetShardStateFromLedger(t, rootLedger, newShardId)
	assert.Equal(t, newShardId, newShard.ShardID)
	assert.Equal(t, TestCommon.GetAccount(creatorName).Address, newShard.Creator)
	assert.Equal(t, uint32(shardstates.SHARD_STATE_ACTIVE), newShard.State)
	assert.Equal(t, shardBlock.Header.Height, newShard.GenesisParentHeight)
	shardConfig := TestCommon.GetConfig(t, newShardId)
	assertVbftConfig(t, shardConfig.Genesis.VBFT, newShard.Config.VbftCfg)
	assert.Equal(t, 7, len(newShard.Peers))
	assert.Equal(t, 7, len(newShard.Config.VbftCfg.Peers))
}

func assertVbftConfig(t *testing.T, except, actual *config.VBFTConfig) {
	assert.Equal(t, except.N, actual.N)
	assert.Equal(t, except.C, actual.C)
	assert.Equal(t, except.K, actual.K)
	assert.Equal(t, except.L, actual.L)
	assert.Equal(t, except.BlockMsgDelay, actual.BlockMsgDelay)
	assert.Equal(t, except.HashMsgDelay, actual.HashMsgDelay)
	assert.Equal(t, except.PeerHandshakeTimeout, actual.PeerHandshakeTimeout)
	assert.Equal(t, except.MaxBlockChangeView, actual.MaxBlockChangeView)
	assert.Equal(t, except.MinInitStake, actual.MinInitStake)
	assert.Equal(t, except.AdminOntID, actual.AdminOntID)
	assert.Equal(t, except.VrfValue, actual.VrfValue)
	assert.Equal(t, except.VrfProof, actual.VrfProof)
}
