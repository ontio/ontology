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

package remote

import (
	"io/ioutil"
	slog "log"
	"net"
	"os"
	"time"

	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/common/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"fmt"
)

var (
	s         *grpc.Server
	edpReader *endpointReader
)

// Start the remote server
func Start(address string, options ...RemotingOption) {
	grpclog.SetLogger(slog.New(ioutil.Discard, "", 0))
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("failed to listen",err.Error())
		os.Exit(1)
	}
	config := defaultRemoteConfig()
	fmt.Println(config.endpointWriterBatchSize)
	for _, option := range options {
		option(config)
	}

	address = lis.Addr().String()
	actor.ProcessRegistry.RegisterAddressResolver(remoteHandler)
	actor.ProcessRegistry.Address = address

	spawnActivatorActor()
	startEndpointManager(config)

	s = grpc.NewServer(config.serverOptions...)
	edpReader = &endpointReader{}
	RegisterRemotingServer(s, edpReader)
	log.Info("Starting Proto.Actor server", string(address))
	go s.Serve(lis)
}

func Shutdown(graceful bool) {
	if graceful {
		edpReader.suspend(true)
		stopEndpointManager()
		stopActivatorActor()

		//For some reason GRPC doesn't want to stop
		//Setup timeout as walkaround but need to figure out in the future.
		//TODO: grpc not stopping
		c := make(chan bool, 1)
		go func() {
			s.GracefulStop()
			c <- true
		}()

		select {
		case <-c:
			log.Info("Stopped Proto.Actor server")
		case <-time.After(time.Second * 10):
			s.Stop()
			log.Info("Stopped Proto.Actor server timeout")
		}
	} else {
		s.Stop()
		log.Info("Killed Proto.Actor server")
	}
}
