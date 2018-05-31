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
package dbft

import (
	"bytes"
	"github.com/ontio/ontology/common/log"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestChangeView_Serialize_Deserialize(t *testing.T) {
	defer os.RemoveAll(log.PATH)
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)

	testViewNum := byte(1)
	msgData := ConsensusMessageData{
		PrepareRequestMsg,
		testViewNum,
	}
	changeView := new(ChangeView)
	changeView.msgData = msgData
	changeView.NewViewNumber = testViewNum
	bf := new(bytes.Buffer)
	err := changeView.Serialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	deserializeMsgData := new(ChangeView)
	err = deserializeMsgData.Deserialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, changeView, deserializeMsgData)

	assert.Equal(t, testViewNum, deserializeMsgData.ViewNumber())
	assert.Equal(t, PrepareRequestMsg, deserializeMsgData.Type())
	assert.Equal(t, msgData, *deserializeMsgData.ConsensusMessageData())
}
