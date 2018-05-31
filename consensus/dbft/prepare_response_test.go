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

func TestPrepareResponse_Serialize_Deserialize(t *testing.T) {
	defer os.RemoveAll(log.PATH)
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)

	testViewNum := byte(1)
	msgData := ConsensusMessageData{
		PrepareResponseMsg,
		testViewNum,
	}
	sig := SignaturesData{
		[]byte("1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527"),
		1,
	}
	prepareResponseMsg := &PrepareResponse{
		msgData:   msgData,
		Signature: sig.Signature,
	}
	buffer := new(bytes.Buffer)
	err := prepareResponseMsg.Serialize(buffer)
	if err != nil {
		t.Fatal(err)
	}
	desePrepareRes := new(PrepareResponse)
	err = desePrepareRes.Deserialize(buffer)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, desePrepareRes, prepareResponseMsg)
	assert.Equal(t, desePrepareRes.ViewNumber(), testViewNum)
	assert.Equal(t, desePrepareRes.Type(), PrepareResponseMsg)
	assert.Equal(t, *desePrepareRes.ConsensusMessageData(), msgData)
}
