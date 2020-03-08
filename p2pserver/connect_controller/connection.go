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
package connect_controller

import (
	"net"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht/kbucket"
)

// Conn is a net.Conn wrapper to do some clean up when Close.
type Conn struct {
	net.Conn
	addr       string
	kid        kbucket.KadId
	boundIndex int
	connectId  uint64
	controller *ConnectController
}

// Close overwrite net.Conn
func (self *Conn) Close() error {
	log.Infof("closing connection: peer %s, address: %s", self.kid.ToHexString(), self.addr)

	self.controller.mutex.Lock()
	defer self.controller.mutex.Unlock()

	self.controller.inoutbounds[self.boundIndex].Remove(self.addr)

	p := self.controller.peers[self.kid]
	if p == nil || p.peer == nil {
		log.Fatalf("connection %s not in controller", self.kid.ToHexString())
	} else if p.connectId == self.connectId { // connection not replaced
		delete(self.controller.peers, self.kid)
	}

	return self.Conn.Close()
}
