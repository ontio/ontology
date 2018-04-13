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

package link

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"testing"
)

var _tlsConfig *tls.Config

func TestTLSDial(t *testing.T) {
	CertPath := "./user1-cert.pem"
	KeyPath := "./user1-cert-key.pem"
	CAPath := "./ca.pem"

	clientCertPool := x509.NewCertPool()

	cacert, err := ioutil.ReadFile(CAPath)
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		t.Error("ReadFile err:", err)
		return
	}

	ok := clientCertPool.AppendCertsFromPEM(cacert)
	if !ok {
		t.Fatalf("failed to parse root certificate")
	}

	conf := &tls.Config{
		RootCAs:      clientCertPool,
		Certificates: []tls.Certificate{cert},
		//InsecureSkipVerify: true,
	}
	println(conf)
}
