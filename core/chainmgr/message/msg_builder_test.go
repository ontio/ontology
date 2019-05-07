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

package message_test

import (
	"bytes"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/xshard_types"
	"testing"

	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/chainmgr/message"
)

func TestNewCrossShardTxMsg(t *testing.T) {
	acc := account.NewAccount("")
	if acc == nil {
		t.Fatalf("failed to new account")
	}
	payload := []xshard_types.CommonShardMsg{&xshard_types.XShardCommitMsg{}}
	tx, err := message.NewCrossShardTxMsg(acc, 100, common.NewShardIDUnchecked(10), 500, 20000, payload)
	if err != nil {
		t.Fatalf("failed to build cross shard tx: %s", err)
	}

	buf := new(bytes.Buffer)
	if err := tx.Serialize(buf); err != nil {
		t.Fatalf("failed to serialize cross shard tx: %s", err)
	}
}
