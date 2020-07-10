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

	"github.com/ontio/ontology/v2/p2pserver/common"
)

// Conn is a net.Conn wrapper to do some clean up when Close.
type Conn struct {
	net.Conn
	addr       string
	listenAddr string
	kid        common.PeerId
	boundIndex int
	connectId  uint64
	controller *ConnectController
}

// Close overwrite net.Conn
// warning: this method will try to lock the controller, be carefull to avoid deadlock
func (self *Conn) Close() error {
	self.controller.logger.Infof("closing connection: peer %s, address: %s", self.kid.ToHexString(), self.addr)

	self.controller.removePeer(self)

	return self.Conn.Close()
}
