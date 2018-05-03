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

package dht

import (
	"github.com/ontio/ontology/common"
	"net"
)

const (
	BUCKET_NUM  = 256
	BUCKET_SIZE = 8
	FACTOR      = 3
)

type Ping struct {
	Version  uint16
	FromID   NodeID
	SrcAddr  net.IP
	SrcPort  uint16
	DestAddr net.IP
	DestPort uint16
}

type Pong struct {
	Version  uint16
	FromID   NodeID
	SrcAddr  net.IP
	SrcPort  uint16
	DestAddr net.IP
	DestPort uint16
}

type FindNode struct {
	FromID   NodeID
	TargetID uint64
}

type Neighbors struct {
	FromID NodeID
	Nodes  []Node
}

type Node struct {
	ID      NodeID
	Hash    common.Uint256
	IP      net.IP
	UDPPort uint16
	TCPPort uint16
}
