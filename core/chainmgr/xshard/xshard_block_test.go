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

package xshard

import (
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
)

func newTestShardMsg(t *testing.T) *types.CrossShardMsg {
	shardMsg := &types.CrossShardMsg{
		CrossShardMsgInfo: &types.CrossShardMsgInfo{
			FromShardID:          common.NewShardIDUnchecked(1),
			MsgHeight:            uint32(90),
			SignMsgHeight:        uint32(100),
			PreCrossShardMsgHash: common.Uint256{},
			CrossShardMsgRoot: common.Uint256{1,2,3},
		},
	}
	return shardMsg
}
func TestCrossShardPool(t *testing.T) {
	InitCrossShardPool(common.NewShardIDUnchecked(1), 100)
	shardMsg := newTestShardMsg(t)
	//acc1 := account.NewAccount("")
	ldg, err := ledger.NewLedger(config.DEFAULT_DATA_DIR, 0)
	if err != nil {
		t.Errorf("failed to new ledger")
		return
	}
	if err = AddCrossShardInfo(ldg, shardMsg); err != nil {
		t.Fatalf("failed add CrossShardInfo:%s", err)
	}
}
func TestAddShardInfo(t *testing.T) {
	shardID := common.NewShardIDUnchecked(0)
	InitCrossShardPool(shardID, 10)
	shardID = common.NewShardIDUnchecked(1)
	ldg, err := ledger.NewLedger(config.DEFAULT_DATA_DIR, 0)
	if err != nil {
		t.Errorf("failed to new ledger")
		return
	}
	AddShardInfo(ldg, shardID)
	shardInfo := GetShardInfo()
	if shardId, present := shardInfo[shardID]; !present {
		t.Errorf("shardID not found:%v", shardID)
	} else {
		t.Logf("shardId found:%v", shardId)
	}
}

func TestAddCrossShardInfo(t *testing.T) {
	shardID := common.NewShardIDUnchecked(0)
	InitCrossShardPool(shardID, 10)
	crossmsg := newTestShardMsg(t)
	lgr, err := ledger.NewLedger(config.DEFAULT_DATA_DIR, 0)
	if err != nil {
		t.Errorf("failed to new ledger")
		return
	}
	err = AddCrossShardInfo(lgr, crossmsg)
	if err != nil {
		t.Errorf("AddCrossShardInfo error")
	}
	acc1 := account.NewAccount("")
	crossShardTx, err := GetCrossShardTxs(lgr, acc1, common.NewShardIDUnchecked(10))
	if err != nil {
		t.Errorf("GetCrossShardTxs failed:%s",err)
	}
	t.Logf("GetCrossShardTxs:%d", len(crossShardTx))
}
