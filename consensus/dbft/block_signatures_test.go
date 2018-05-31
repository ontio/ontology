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

func TestBlockSignatures_Serialize_Deserialize(t *testing.T) {
	defer os.RemoveAll(log.PATH)
	log.InitLog(log.InfoLog, log.PATH, log.Stdout)

	sigData := []SignaturesData{
		{
			[]byte("1202028541d32f3b09180b00affe67a40516846c16663ccb916fd2db8106619f087527"),
			1,
		},
		{
			[]byte("120202dfb161f757921898ec2e30e3618d5c6646d993153b89312bac36d7688912c0ce"),
			2,
		},
		{
			[]byte("1202039dab38326268fe82fb7967fe2e7f5f6eaced6ec711148a66fbb8480c321c19dd"),
			3,
		},
	}
	blockSig := new(BlockSignatures)
	blockSig.Signatures = sigData
	testViewNum := byte(1)
	blockSig.msgData = ConsensusMessageData{
		PrepareRequestMsg,
		testViewNum,
	}
	bf := new(bytes.Buffer)
	err := blockSig.Serialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	deserializeBlockSig := new(BlockSignatures)
	err = deserializeBlockSig.Deserialize(bf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, blockSig, deserializeBlockSig)

	assert.Equal(t, deserializeBlockSig.Type(), PrepareRequestMsg)

	assert.Equal(t, deserializeBlockSig.ViewNumber(), testViewNum)

	assert.Equal(t, *deserializeBlockSig.ConsensusMessageData(), blockSig.msgData)
}
