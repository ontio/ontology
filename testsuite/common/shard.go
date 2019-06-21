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

package TestCommon

import (
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/chainmgr/xshard"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func GetShardStateFromLedger(t *testing.T, lgr *ledger.Ledger, shardID common.ShardID) *shardstates.ShardState {
	state, err := xshard.GetShardState(lgr, shardID)
	if err != nil {
		t.Fatalf(err.Error())
	}
	return state
}
