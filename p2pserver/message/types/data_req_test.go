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
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package types

import (
	"bytes"
	"encoding/binary"
	"testing"

	cm "github.com/ontio/ontology/common"
	comm "github.com/ontio/ontology/p2pserver/common"
)

func TestDataReqSerializationDeserialization(t *testing.T) {
	var msg DataReq
	msg.MsgHdr.Magic = comm.NETMAGIC
	copy(msg.MsgHdr.CMD[0:7], "getdata")
	msg.DataType = 0x02

	hashstr := "8932da73f52b1e22f30c609988ed1f693b6144f74fed9a2a20869afa7abfdf5e"
	bhash, _ := cm.HexToBytes(hashstr)
	copy(msg.Hash[:], bhash)
	t.Log("new getdata message before serialize Hash = ", msg.Hash)
	p := bytes.NewBuffer([]byte{})
	err := binary.Write(p, binary.LittleEndian, &(msg.DataType))
	msg.Hash.Serialize(p)
	if err != nil {
		t.Error("Binary Write failed at new getdata Msg")
		return
	}

	checkSumBuf := CheckSum(p.Bytes())
	msg.Init("getdata", checkSumBuf, uint32(len(p.Bytes())))

	m, err := msg.Serialization()
	if err != nil {
		t.Error("Error Convert net message ", err.Error())
		return
	}

	var demsg DataReq
	err = demsg.Deserialization(m)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("getheaders Test_Deserialization sucessful")
	}

	t.Log("new getdata message after deserialize DataType = ", demsg.DataType)
	t.Log("new getdata message after deserialize Hash = ", demsg.Hash)
}
