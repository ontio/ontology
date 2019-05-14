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

package shardstates_test

import (
	"github.com/ontio/ontology/common"
	"testing"

	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func TestCreateShardEvent(t *testing.T) {
	evt := &shardstates.CreateShardEvent{
		SourceShardID: common.NewShardIDUnchecked(1),
		Height:        110,
		NewShardID:    common.NewShardIDUnchecked(1),
	}

	sink := common.NewZeroCopySink(0)
	evt.Serialization(sink)

	evt2 := &shardstates.CreateShardEvent{}
	if err := evt2.Deserialization(common.NewZeroCopySource(sink.Bytes())); err != nil {
		t.Fatalf("deserialize createEvt: %s", err)
	}

	if evt.SourceShardID != evt2.SourceShardID ||
		evt.Height != evt2.Height ||
		evt.NewShardID != evt2.NewShardID {
		t.Fatalf("mismatch evt")
	}
}
