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
	"net"
	"testing"

	comm "github.com/ontio/ontology/p2pserver/common"
)

func TestAddressSerializationDeserialization(t *testing.T) {
	var msg Addr
	msg.Hdr.Magic = comm.NETMAGIC
	copy(msg.Hdr.CMD[0:7], "addr")
	var addr [16]byte
	ip := net.ParseIP("192.168.0.1")
	ip.To16()
	copy(addr[:], ip[:16])
	nodeAddr := comm.PeerAddr{
		Time:          12345678,
		Services:      100,
		IpAddr:        addr,
		Port:          8080,
		ConsensusPort: 8081,
		ID:            987654321,
	}
	msg.NodeAddrs = append(msg.NodeAddrs, nodeAddr)
	t.Log("new addr message before serialize time = ", nodeAddr.Time)
	t.Log("new addr message before serialize Services = ", nodeAddr.Services)
	t.Log("new addr message before serialize IpAddr = 192.168.0.1")
	t.Log("new addr message before serialize Port = ", nodeAddr.Port)
	t.Log("new addr message before serialize ConsensusPort = ", nodeAddr.ConsensusPort)
	t.Log("new addr message before serialize ID = ", nodeAddr.ID)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, uint64(1))
	if err != nil {
		t.Error("Binary Write failed at new Msg: ", err.Error())
		return
	}

	err = binary.Write(p, binary.LittleEndian, msg.NodeAddrs)
	if err != nil {
		t.Error("Binary Write failed at new Msg: ", err.Error())
		return
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.Hdr.Checksum))
	msg.Hdr.Length = uint32(len(p.Bytes()))

	m, err := msg.Serialization()
	if err != nil {
		t.Error("Error Convert net message ", err.Error())
		return
	}

	var demsg Addr
	err = demsg.Deserialization(m)
	if err != nil {
		t.Error(err)
		return
	} else {
		t.Log("addr Test_Deserialization sucessful")
	}
	for _, v := range demsg.NodeAddrs {
		t.Log("new addr message after deserialize time = ", v.Time)
		t.Log("new addr message after deserialize Services = ", v.Services)
		var ip net.IP
		ip = v.IpAddr[:]
		t.Log("new addr message after deserialize IpAddr = ", ip.To16().String())
		t.Log("new addr message after deserialize Port = ", v.Port)
		t.Log("new addr message after deserialize ConsensusPort = ", v.ConsensusPort)
		t.Log("new addr message after deserialize ID = ", v.ID)
	}
}
