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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/types"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestConsensusMessageData_Serialize_Deserialize(t *testing.T) {
	defer os.RemoveAll(log.PATH)
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)

	msgData := ConsensusMessageData{
		PrepareRequestMsg,
		byte(1),
	}
	bf := new(bytes.Buffer)
	msgData.Serialize(bf)
	desMsgData := new(ConsensusMessageData)
	desMsgData.Deserialize(bf)
	assert.Equal(t, *desMsgData, msgData)
}

func TestDeserializeMessage(t *testing.T) {
	defer os.RemoveAll(log.PATH)
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)

	nonce := uint64(time.Now().Unix())
	nextBookkeeper, _ := common.AddressFromBase58("TA9TVuR4Ynn4VotfpExY5SaEy8a99obFPr")
	sig := SignaturesData{
		[]byte("1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527"),
		1,
	}
	buffer := new(bytes.Buffer)
	// test prepare request msg
	msgData := ConsensusMessageData{
		PrepareRequestMsg,
		byte(1),
	}
	prepareRequest := &PrepareRequest{
		msgData:        msgData,
		Nonce:          nonce,
		NextBookkeeper: nextBookkeeper,
		Signature:      sig.Signature,
		Transactions:   []*types.Transaction{},
	}
	err := prepareRequest.Serialize(buffer)
	if err != nil {
		t.Fatal(err)
	}
	consensusMsg, err := DeserializeMessage(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, consensusMsg.ConsensusMessageData(), prepareRequest.ConsensusMessageData())
	assert.Equal(t, consensusMsg.Type(), prepareRequest.Type())
	assert.Equal(t, consensusMsg.ViewNumber(), prepareRequest.ViewNumber())

	// test prepare response msg
	msgData.Type = PrepareResponseMsg
	prepareResponse := &PrepareResponse{
		msgData:   msgData,
		Signature: sig.Signature,
	}
	buffer.Reset()
	err = prepareResponse.Serialize(buffer)
	if err != nil {
		t.Fatal(err)
	}
	consensusMsg, err = DeserializeMessage(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, consensusMsg.ConsensusMessageData(), prepareResponse.ConsensusMessageData())
	assert.Equal(t, consensusMsg.Type(), prepareResponse.Type())
	assert.Equal(t, consensusMsg.ViewNumber(), prepareResponse.ViewNumber())

	// test block signatures msg
	msgData.Type = BlockSignaturesMsg
	blockSigMsg := &BlockSignatures{
		msgData:    msgData,
		Signatures: []SignaturesData{sig},
	}
	buffer.Reset()
	err = blockSigMsg.Serialize(buffer)
	if err != nil {
		t.Fatal(err)
	}
	consensusMsg, err = DeserializeMessage(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, consensusMsg.ConsensusMessageData(), blockSigMsg.ConsensusMessageData())
	assert.Equal(t, consensusMsg.Type(), blockSigMsg.Type())
	assert.Equal(t, consensusMsg.ViewNumber(), blockSigMsg.ViewNumber())

	// test change view msg
	msgData.Type = ChangeViewMsg
	changeViewMsg := &ChangeView{
		msgData:       msgData,
		NewViewNumber: byte(1),
	}
	buffer.Reset()
	err = changeViewMsg.Serialize(buffer)
	if err != nil {
		t.Fatal(err)
	}
	consensusMsg, err = DeserializeMessage(buffer.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, consensusMsg.ConsensusMessageData(), changeViewMsg.ConsensusMessageData())
	assert.Equal(t, consensusMsg.Type(), changeViewMsg.Type())
	assert.Equal(t, consensusMsg.ViewNumber(), changeViewMsg.ViewNumber())
}
