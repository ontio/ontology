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

package ledgerstore

import (
	"math/rand"
	"testing"

	"github.com/ontio/ontology/common"
	payload "github.com/ontio/ontology/core/payload"
	msg "github.com/ontio/ontology/events/message"
)

func TestSaveMetatEventMsg(t *testing.T) {
	var addr common.Address
	rand.Read(addr[:])
	metaDataCode := &payload.MetaDataCode{
		OntVersion: 1,
		Contract:   addr,
		ShardId:    1,
	}
	metaDataEvent := &msg.MetaDataEvent{
		Height:   123,
		MetaData: metaDataCode,
	}
	testEventStore.NewBatch()
	err := testEventStore.SaveMetaDataEvent(metaDataEvent.Height, metaDataCode)
	if err != nil {
		t.Errorf("SaveMetaEvent err:%s", err)
		return
	}
	err = testEventStore.CommitTo()
	if err != nil {
		t.Errorf("CommitTo err :%s", err)
		return
	}
	msg, err := testEventStore.GetMetaDataEvent(metaDataEvent.Height, metaDataCode.Contract)
	if err != nil {
		t.Errorf("GetMeteEvent err:%s", err)
	}
	if metaDataCode.ShardId != msg.ShardId {
		t.Errorf("metaData shardId:%d not match msg shardID:%d", metaDataCode.ShardId, msg.ShardId)
	}
}
