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
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
)

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
