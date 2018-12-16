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

package quic

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/p2pserver/common"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
)

const (
	MAX_INCOMING_STREAMS   = 10000
	MAX_INCOMING_UNISTREAM = 10000
)


// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}
}

var quicConfig = &quic.Config{
	Versions:                              []quic.VersionNumber{101},
	IdleTimeout:                           time.Second * (config.DEFAULT_GEN_BLOCK_TIME / common.UPDATE_RATE_PER_BLOCK) * common.KEEPALIVE_TIMEOUT,
	MaxIncomingStreams:                    MAX_INCOMING_STREAMS,
	MaxIncomingUniStreams:                 MAX_INCOMING_UNISTREAM,              // disable unidirectional streams
	MaxReceiveStreamFlowControlWindow:     3 * (1 << 20),   // 3 MB
	MaxReceiveConnectionFlowControlWindow: 4.5 * (1 << 20), // 4.5 MB
	AcceptCookie: func(clientAddr net.Addr, cookie *quic.Cookie) bool {
		// TODO(#6): require source address validation when under load
		return true
	},
	KeepAlive: false,
}

type transport struct {
	tlsConf *tls.Config
}

func NewTransport( ) (tsp.Transport, error) {

	return &transport{	}, nil
}

func (this * transport) Dial(addr string) (tsp.Connection, error) {

	return this.DialWithTimeout(addr, time.Second * common.DIAL_TIMEOUT)
}

func (this * transport) DialWithTimeout(addr string, timeout time.Duration) (tsp.Connection, error) {
	//var qConfig *quic.Config = nil
	if timeout >= 0 {
		//qConfig = &quic.Config{IdleTimeout: timeout}
		quicConfig.HandshakeTimeout = timeout
	}

	session, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, quicConfig)
	if err != nil {
		return nil, &tsp.DialError{"TCP", addr, err.Error()}
	}

	return &connection{sess: session}, nil
}

func (this * transport) Listen(port uint16) (tsp.Listener, error) {

	tlsConf, _ := tsp.GetServerTLSConfig()
	if tlsConf == nil {
		tlsConf = generateTLSConfig()
	}

	return  newListener(port, tlsConf)
}

func (this* transport) GetReqInterval() int {

	return common.REQ_INTERVAL_QUIC
}

func (this * transport) ProtocolCode() int {
	return  common.T_QUIC
}

func (this * transport) ProtocolName() string {
	return "QUIC"
}







