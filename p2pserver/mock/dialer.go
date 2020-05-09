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
	"github.com/ontio/ontology/p2pserver/connect_controller"
)

type dialer struct {
	id      common.PeerId
	address string
	network *network
}

var _ connect_controller.Dialer = &dialer{}

func (d *dialer) Dial(nodeAddr string) (net.Conn, error) {
	d.network.Lock()
	defer d.network.Unlock()
	l, exist := d.network.listeners[nodeAddr]

	if !exist {
		return nil, errors.New("can not be reached")
	}

	if _, allow := d.network.canEstablish[combineKey(d.id, l.id)]; !allow {
		return nil, errors.New("can not be reached")
	}

	c, s := net.Pipe()

	cw := connWraper{c, d.address, d.network, l.address}
	sw := connWraper{s, l.address, d.network, d.address}
	l.PushToAccept(sw)

	return cw, nil
}

func (n *network) NewDialer(pid common.PeerId) connect_controller.Dialer {
	host := n.nextFakeIP()
	return n.NewDialerWithHost(pid, host)
}

func (n *network) NewDialerWithHost(pid common.PeerId, host string) connect_controller.Dialer {
	addr := host + ":" + n.nextPortString()

	d := &dialer{
		id:      pid,
		address: addr,
		network: n,
	}

	return d
}

func (d *dialer) ID() common.PeerId {
	return d.id
}
