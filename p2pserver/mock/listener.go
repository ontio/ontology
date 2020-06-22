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
	"errors"
	"net"

	"github.com/ontio/ontology/p2pserver/common"
)

type Listener struct {
	id      common.PeerId
	conn    chan net.Conn
	address string
}

var _ net.Listener = &Listener{}
var _ net.Addr = &Listener{}

func (n *network) NewListener(id common.PeerId) (string, net.Listener) {
	ip := n.nextFakeIP()
	hostport := ip + ":" + n.nextPortString()
	return hostport, n.NewListenerWithAddr(id, hostport)
}

func (n *network) NewListenerWithAddr(id common.PeerId, hostport string) net.Listener {
	ret := &Listener{
		id:      id,
		address: hostport,
		conn:    make(chan net.Conn),
	}

	n.Lock()
	n.listeners[hostport] = ret
	n.Unlock()

	return ret
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case conn, ok := <-l.conn:
		if ok {
			return conn, nil
		}
		return nil, errors.New("closed channel")
	}
}

func (l *Listener) Close() error {
	close(l.conn)
	return nil
}

func (l *Listener) Addr() net.Addr {
	// listeners's local listen address is useless
	return l
}

func (l *Listener) Network() string {
	return "tcp"
}

func (l *Listener) String() string {
	return l.address
}

func (l *Listener) ID() string {
	return l.id.ToHexString()
}

func (l *Listener) PushToAccept(conn net.Conn) {
	go func() {
		l.conn <- conn
	}()
}
