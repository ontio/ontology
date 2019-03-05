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
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
)

type transport struct { }

// createListener creates a net listener on the port
func createListener(port uint16) (net.Listener, error) {
	var listener net.Listener
	var err error

	isTls := config.DefConfig.P2PNode.IsTLS
	if isTls {
		tlsConf, _:= tsp.GetServerTLSConfig()
		if tlsConf == nil {
			log.Error("[p2p]GetServerTLSConfig failed")
			return nil, errors.New("[p2p]GetServerTLSConfig failed")
		}

		listener, err = initTlsListen(port, tlsConf)
		if err != nil {
			log.Error("[p2p]initTlslisten failed")
			return nil, errors.New("[p2p]initTlslisten failed")
		}
	} else {
		listener, err = initNonTlsListen(port)
		if err != nil {
			log.Error("[p2p]initNonTlsListen failed")
			return nil, errors.New("[p2p]initNonTlsListen failed")
		}
	}
	return listener, nil
}

//nonTLSDial return net.Conn with nonTls
func nonTLSDial(addr string, timeout time.Duration) (net.Conn, error) {
	log.Trace()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//TLSDial return net.Conn with TLS
func TLSDial(nodeAddr string, timeout time.Duration, tlsConf *tls.Config) (net.Conn, error) {

	var dialer net.Dialer
	dialer.Timeout = timeout
	conn, err := tls.DialWithDialer(&dialer, "tcp", nodeAddr, tlsConf)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

//initNonTlsListen return net.Listener with nonTls mode
func initNonTlsListen(port uint16) (net.Listener, error) {
	log.Trace()
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(int(port)))
	if err != nil {
		log.Error("[p2p]Error listening\n", err.Error())
		return nil, err
	}
	return listener, nil
}

//initTlsListen return net.Listener with Tls mode
func initTlsListen(port uint16, tlsConf *tls.Config) (net.Listener, error) {

	log.Info("[p2p]TLS listen port is ", strconv.Itoa(int(port)))
	listener, err := tls.Listen("tcp", ":"+strconv.Itoa(int(port)), tlsConf)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
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
			return nil, &tsp.DialError{"TCP", addr, "[p2p]GetServerTLSConfig failed"}
		}
		c, err = TLSDial(addr, timeout, tlsConf)
	}

	if err != nil {
		return nil, &tsp.DialError{"TCP", addr, err.Error()}
	}

	reader := bufio.NewReaderSize(c, common.MAX_BUF_LEN)

	return &connection{c, reader}, nil
}

func (this * transport) Listen(port uint16) (tsp.Listener, error) {

	return newListen(port)
}

func (this* transport) GetReqInterval() int {

	return common.REQ_INTERVAL
}

func (this * transport) ProtocolCode() common.TransportType {

	return common.T_TCP
}

func (this * transport) ProtocolName() string {

	return "TCP"
}
