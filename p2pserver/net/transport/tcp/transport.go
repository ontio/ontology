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

package tcp

import (
	"bufio"
	"errors"
	"io"
	"net"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
)

type connection struct {
	net.Conn
	io.Reader
}

type transport struct { }

func (this * connection) GetReader() (io.Reader, error) {

	return this.Reader, nil
}

func (this * connection) Write(b []byte) (n int, err error) {

	return this.Conn.Write(b)
}

func (this * connection) Close() error {

	return  this.Conn.Close()
}

func (this* connection) LocalAddr() net.Addr {

	return this.Conn.LocalAddr()
}

func (this* connection) RemoteAddr() net.Addr {

	return this.Conn.RemoteAddr()
}
func (this * connection) SetWriteDeadline(t time.Time) error {

	return  this.Conn.SetWriteDeadline(t)
}

func NewTransport() (tsp.Transport, error) {
	return &transport{}, nil
}

func (this * transport) Dial(addr string) (tsp.Connection, error) {

	return this.DialWithTimeout(addr, time.Second * common.DIAL_TIMEOUT)
}

func (this * transport) DialWithTimeout(addr string, timeout time.Duration) (tsp.Connection, error){

	var c net.Conn = nil
	var err error = nil
	if !config.DefConfig.P2PNode.IsTLS {
		c, err = nonTLSDial(addr, timeout)
	} else {
		tlsConf, _ := tsp.GetClientTLSConfig()
		if tlsConf == nil {
			log.Error("[p2p]GetServerTLSConfig failed")
			return nil, errors.New("[p2p]GetServerTLSConfig failed")
		}
		c, err = TLSDial(addr, timeout, tlsConf)
	}

	if err != nil {
		return nil, err
	}

	reader := bufio.NewReaderSize(c, common.MAX_BUF_LEN)

	return &connection{c, reader}, nil
}

func (this * transport) Listen(port uint16) (tsp.Listener, error) {
	return newListen(port)
}

func (this * transport) ProtocolCode() int {

	return tsp.T_TCP
}

func (this * transport) ProtocolName() string {

	return "TCP"
}
