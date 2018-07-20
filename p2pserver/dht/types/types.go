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

// Package common privides constants, common types for other packages
package types

import (
	"net"
	"time"
)

const (
	BUCKET_NUM        = NODE_ID_BITS
	BUCKET_SIZE       = 8
	FACTOR            = 3
	MSG_CACHE         = 10240
	PING_TIMEOUT      = 500 * time.Millisecond
	FIND_NODE_TIMEOUT = 500 * time.Millisecond
	DEFAULT_TIMEOUT   = 1 * time.Second
	REFRESH_INTERVAL  = 1 * time.Hour
)

type ptype uint8

const (
	ping_rpc ptype = iota
	pong_rpc
	find_node_rpc
	neighbors_rpc
)

type DHTMessage struct {
	From    *net.UDPAddr
	Payload []byte
}

type Node struct {
	ID      NodeID `json:"node_id"`
	IP      string `json:"IP"`
	UDPPort uint16 `json:UDPPort`
	TCPPort uint16 `json:TCPPort`
}

type feedType uint8

const (
	Add feedType = iota
	Del
)

type FeedEvent struct {
	EvtType feedType
	Event   interface{}
}
