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

package transport

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"errors"

	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
)

func generateCertAndCertPool() (tls.Certificate, *x509.CertPool, error) {

	CertPath := config.DefConfig.P2PNode.CertPath
	KeyPath := config.DefConfig.P2PNode.KeyPath
	CAPath := config.DefConfig.P2PNode.CAPath

	// load cert
	cert, err := tls.LoadX509KeyPair(CertPath, KeyPath)
	if err != nil {
		log.Error("[p2p]load keys fail", err)
		return cert, nil, err
	}
	// load root ca
	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		log.Error("[p2p]read ca fail", err)
		return cert, nil, err
	}
	pool := x509.NewCertPool()
	ret := pool.AppendCertsFromPEM(caData)
	if !ret {
		return cert, nil, errors.New("[p2p]failed to parse root certificate")
	}

	return cert, pool, nil
}

func getTLSConfig(role string) (*tls.Config, error) {

	isTls := config.DefConfig.P2PNode.IsTLS
	if !isTls {
		return nil, nil
	}

	cert, pool, err := generateCertAndCertPool()
	if err != nil {
		log.Error("[p2p]fail to generateCertAndCertPool", err)
		return nil, err
	}

	switch role {
	case "client":
		return  &tls.Config{
			RootCAs:      pool,
			Certificates: []tls.Certificate{cert},
		}, nil
	case "server":
		return  &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      pool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    pool,
		}, nil
	}

	return nil, errors.New("Invalid role input")
}

func GetClientTLSConfig() (*tls.Config, error) {

	return getTLSConfig("client")
}

func GetServerTLSConfig() (*tls.Config, error) {

	return getTLSConfig("server")
}