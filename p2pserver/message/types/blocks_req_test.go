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
	"crypto/sha256"
	"encoding/binary"
	"testing"

	cm "github.com/ontio/ontology/common"
	comm "github.com/ontio/ontology/p2pserver/common"
)

func TestBlkReqSerializationDeserialization(t *testing.T) {
	var msg BlocksReq
	msg.MsgHdr.Magic = comm.NETMAGIC
	copy(msg.MsgHdr.CMD[0:7], "getblocks")
	msg.P.HeaderHashCount = 1

	hashstr := "8932da73f52b1e22f30c609988ed1f693b6144f74fed9a2a20869afa7abfdf5e"
	bhash, _ := cm.HexToBytes(hashstr)
	copy(msg.P.HashStart[:], bhash)
	t.Log("new getblocks message before serialize HashStart = ", msg.P.HashStart)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(msg.P))
	if err != nil {
		t.Error("Binary Write failed at new getblocks")
		return
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr.Checksum))
	msg.MsgHdr.Length = uint32(len(p.Bytes()))
	m, err := msg.Serialization()
	if err != nil {
		t.Error("Error Convert net message ", err.Error())
		return
	}

	var demsg BlocksReq
	err = demsg.Deserialization(m)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("getblocks Test_Deserialization sucessful")
	}

	t.Log("new getblocks message after deserialize HeaderHashCount = ", demsg.P.HeaderHashCount)
	t.Log("new getblocks message after deserialize HashStart = ", demsg.P.HashStart)
}
