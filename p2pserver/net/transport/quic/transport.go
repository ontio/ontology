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
	"io"
	"math/big"
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
	"github.com/ontio/ontology/common/log"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
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

type connection struct {
	sess          quic.Session
	streamWTimeOut time.Time
}

type transport struct {
	tlsConf *tls.Config
}

func (this * connection) GetReader() (io.Reader, error) {

	stream, err := this.sess.AcceptStream()
	if err != nil {
		log.Errorf("[p2p]AcceptStream ERR:%s", err)
		return  nil, err
	}

	return stream, nil
}

func (this * connection) Write(b []byte) (int, error) {

	stream, err := this.sess.OpenStreamSync()
	if err != nil {
		log.Errorf("[p2p]OpenStreamSync ERR:%s", err)
		return 0, err
	}

	stream.SetWriteDeadline(this.streamWTimeOut)
	cntW, errW := stream.Write(b)
	if errW != nil {
		log.Errorf("[p2p] Write err by stream:%s", errW)
		return 0, errW
	}

	return cntW, errW
}

func (this * connection) Close() error {

	return this.sess.Close()
}

func (this* connection) LocalAddr() net.Addr {

	return this.sess.LocalAddr()
}

func (this* connection) RemoteAddr() net.Addr {

	return this.sess.RemoteAddr()
}

func (this * connection) SetWriteDeadline(t time.Time) error {

	this.streamWTimeOut = t

	return nil
}

func NewTransport( ) (tsp.Transport, error) {

	return &transport{	}, nil
}

func (this * transport) Dial(addr string) (tsp.Connection, error) {

	return this.DialWithTimeout(addr, -1)
}

func (this * transport) DialWithTimeout(addr string, timeout time.Duration) (tsp.Connection, error) {
	var qConfig *quic.Config = nil
	if timeout >= 0 {
		qConfig = &quic.Config{IdleTimeout: timeout}
	}
	session, err := quic.DialAddr(addr, &tls.Config{InsecureSkipVerify: true}, qConfig)
	if err != nil {
		return nil, err
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

func (this * transport) ProtocolCode() int {
	return  tsp.T_QUICK
}

func (this * transport) ProtocolName() string {
	return "QUICK"
}







