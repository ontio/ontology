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

package shardmgmt_test

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
)

func newCreateShardParam(t *testing.T, acc *account.Account) []byte {
	param := &shardmgmt.CreateShardParam{
		ParentShardID: common.NewShardIDUnchecked(100),
		Creator:       acc.Address,
	}

	buf := new(bytes.Buffer)
	if err := param.Serialize(buf); err != nil {
		t.Fatalf("serialize create shard param: %s", err)
	}
	return buf.Bytes()
}

func TestCreateShardParam(t *testing.T) {
	acc := account.NewAccount("")
	if acc == nil {
		t.Fatalf("new account failed")
	}

	paramBytes := newCreateShardParam(t, acc)

	param := &shardmgmt.CreateShardParam{}
	if err := param.Deserialize(bytes.NewBuffer(paramBytes)); err != nil {
		t.Fatalf("deserialize create shard param: %s", err)
	}

	if bytes.Compare(param.Creator[:], acc.Address[:]) != 0 {
		t.Fatalf("unmatched creator address: %v vs %v", param.Creator, acc.Address)
	}
	if param.ParentShardID.ToUint64() != 100 {
		t.Fatalf("unmatched parent shard id: %d vs %d", param.ParentShardID, 100)
	}
}
