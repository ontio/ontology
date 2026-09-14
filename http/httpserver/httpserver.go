/*
 * Copyright (C) 2024 The ontology Authors
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

// Package httpserver holds the read and idle timeouts shared by the node's HTTP
// entry points (JSON-RPC, local RPC, ETH RPC, REST, WebSocket, GraphQL, the
// node info page and the signature server). Without them a client can trickle
// request headers or a body indefinitely and pin a connection and its
// goroutine.
package httpserver

import (
	"net/http"
	"time"
)

const (
	// ReadHeaderTimeout bounds how long a client may take to send the request
	// headers. This is the main defence against slow-header (slowloris)
	// attacks.
	ReadHeaderTimeout = 10 * time.Second

	// ReadTimeout bounds the total time spent reading a request, body
	// included. Requests are small (JSON-RPC envelopes, transactions), so this
	// is generous while still capping a slow sender.
	ReadTimeout = 30 * time.Second

	// IdleTimeout bounds how long a keep-alive connection may sit idle between
	// requests.
	IdleTimeout = 120 * time.Second
)

// New returns an http.Server with the shared timeouts applied. Callers can
// still set Addr, TLSConfig, etc. WriteTimeout is deliberately left unset so
// that long running handlers are not cut off mid-response.
func New(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		IdleTimeout:       IdleTimeout,
	}
}
