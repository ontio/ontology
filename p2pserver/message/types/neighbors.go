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
	//"errors"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/dht/types"
)

type Neighbors struct {
	FromID types.NodeID
	Nodes  []types.Node
}

func (this *Neighbors) CmdType() string {
	return common.DHT_NEIGHBORS
}

//Serialize message payload
func (this Neighbors) Serialization() ([]byte, error) {
	p := bytes.NewBuffer([]byte{})
	err := serialization.WriteVarBytes(p, this.FromID[:])
	if err != nil {
		log.Errorf("failed to serialize from id %v. FromID %x", err, this.FromID)
		return nil, err
	}

	err = serialization.WriteVarUint(p, uint64(len(this.Nodes)))
	if err != nil {
		log.Errorf("failed to serialize the length of nodes %v. len %d", err, len(this.Nodes))
		return nil, err
	}

	for _, node := range this.Nodes {
		err := serialization.WriteVarBytes(p, node.ID[:])
		if err != nil {
			log.Errorf("failed to serialize node id %v. ID %x", err, node.ID)
			return nil, err
		}
		err = serialization.WriteString(p, node.IP)
		if err != nil {
			log.Errorf("failed to serialize node ip %v. ip %s", err, node.IP)
			return nil, err
		}
		err = serialization.WriteUint16(p, node.UDPPort)
		if err != nil {
			log.Errorf("failed to serialize node udp port %v. udp port %s", err, node.UDPPort)
			return nil, err
		}
		err = serialization.WriteUint16(p, node.TCPPort)
		if err != nil {
			log.Errorf("failed to serialize node udp port %v. tcp port %s", err, node.TCPPort)
			return nil, err
		}
	}
	return p.Bytes(), nil
}

//Deserialize message payload
func (this *Neighbors) Deserialization(p []byte) error {
	buf := bytes.NewBuffer(p)
	id, err := serialization.ReadVarBytes(buf)
	if err != nil {
		return err
	}
	copy(this.FromID[:], id)

	num, err := serialization.ReadVarUint(buf, 0)
	if err != nil {
		return err
	}
	this.Nodes = make([]types.Node, 0, num)
	for i := 0; i < int(num); i++ {
		node := new(types.Node)
		id, err := serialization.ReadVarBytes(buf)
		if err != nil {
			return err
		}
		copy(node.ID[:], id)
		node.IP, err = serialization.ReadString(buf)
		if err != nil {
			log.Errorf("failed to deserialize node ip %v", err)
			return err
		}
		node.UDPPort, err = serialization.ReadUint16(buf)
		if err != nil {
			log.Errorf("failed to deserialize node udp port %v", err)
			return err
		}
		node.TCPPort, err = serialization.ReadUint16(buf)
		if err != nil {
			log.Errorf("failed to deserialize node tcp port %v", err)
			return err
		}
		this.Nodes = append(this.Nodes, *node)
	}

	return nil
}
