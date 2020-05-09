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

package mock

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/ontio/ontology/p2pserver/common"
)

func init() {
	common.Difficulty = 1
}

type network struct {
	sync.RWMutex
	canEstablish map[string]struct{}
	listeners    map[string]*Listener
	startID      uint32
}

var _ Network = &network{}

func NewNetwork() Network {
	ret := &network{
		// id -> [id...]
		canEstablish: make(map[string]struct{}),
		// host:port -> Listener
		listeners: make(map[string]*Listener),
		startID:   0,
	}

	return ret
}

func (n *network) nextID() uint32 {
	return atomic.AddUint32(&n.startID, 1)
}

func (n *network) nextFakeIP() string {
	id := n.nextID()
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, id)
	ip := net.IP(b)

	return ip.String()
}

func (n *network) nextPort() uint16 {
	port := make([]byte, 2)
	rand.Read(port)
	return binary.BigEndian.Uint16(port)
}

func (n *network) nextPortString() string {
	port := n.nextPort()
	return strconv.Itoa(int(port))
}

func combineKey(id1, id2 common.PeerId) string {
	s1 := id1.ToHexString()
	s2 := id2.ToHexString()

	if s1 <= s2 {
		return s1 + "|" + s2
	}
	return s2 + "|" + s1
}

func (n *network) AllowConnect(id1, id2 common.PeerId) {
	n.Lock()
	defer n.Unlock()

	n.canEstablish[combineKey(id1, id2)] = struct{}{}
}

// DeliverRate TODO
func (n *network) DeliverRate(percent uint) {

}

type connWraper struct {
	net.Conn
	address string
	network *network
	remote  string
}

var _ net.Addr = &connWraper{}

func (cw *connWraper) Network() string {
	return "tcp"
}

func (cw *connWraper) String() string {
	return cw.address
}

func (cw connWraper) LocalAddr() net.Addr {
	return &cw
}

func (cw connWraper) RemoteAddr() net.Addr {
	w := &connWraper{
		address: cw.remote,
	}
	return w
}
