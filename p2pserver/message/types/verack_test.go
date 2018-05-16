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

package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/p2pserver/common"
)

func TestVerackSerializationDeserialization(t *testing.T) {
	var msg VerACK
	msg.MsgHdr.Magic = common.NETMAGIC
	copy(msg.MsgHdr.CMD[0:7], "verack")
	msg.IsConsensus = false
	t.Log("new verack message before serialize msg.IsConsensus = false")
	tmpBuffer := bytes.NewBuffer([]byte{})
	serialization.WriteBool(tmpBuffer, msg.IsConsensus)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		t.Error("Binary Write failed at new Msg")
		return
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.MsgHdr.Checksum))
	msg.MsgHdr.Length = uint32(len(b.Bytes()))

	p, err := msg.Serialization()
	if err != nil {
		t.Error("Error Convert net message ", err.Error())
		return
	}

	var demsg VerACK
	err = demsg.Deserialization(p)

	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("VerACK Test_Deserialization successful")
	}
	t.Log("deserialize verack message, msg.IsConsensus = ", demsg.IsConsensus)
}
