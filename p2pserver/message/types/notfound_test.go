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

	cm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/p2pserver/common"
)

func Uint256ParseFromBytes(f []byte) cm.Uint256 {
	if len(f) != 32 {
		return cm.Uint256{}
	}

	var hash [32]uint8
	for i := 0; i < 32; i++ {
		hash[i] = f[i]
	}
	return cm.Uint256(hash)
}

func TestNotFoundSerializationDeserialization(t *testing.T) {
	var msg NotFound
	str := "123456"
	hash := []byte(str)
	msg.Hash = Uint256ParseFromBytes(hash)
	msg.MsgHdr.Magic = common.NETMAGIC
	cmd := "notfound"
	copy(msg.MsgHdr.CMD[0:len(cmd)], cmd)
	tmpBuffer := bytes.NewBuffer([]byte{})
	msg.Hash.Serialize(tmpBuffer)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, tmpBuffer.Bytes())
	if err != nil {
		t.Error("Binary Write failed at new notfound Msg")
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
	var demsg NotFound
	err = demsg.Deserialization(m)

	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("Notfound Test_Deserialization sucessful")
	}
}
