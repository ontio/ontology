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

package zmqremote

import (

	"github.com/Ontology/eventbus/actor"
	zmq "github.com/pebbe/zmq4"
	"github.com/Ontology/common/log"
)

var (
	edpReader *endpointReader
	conn      *zmq.Socket
)

func Start(address string) {

	//fmt.Println("address1:" + address)
	actor.ProcessRegistry.RegisterAddressResolver(remoteHandler)
	actor.ProcessRegistry.Address = address

	spawnActivatorActor()
	startEndpointManager()

	edpReader = &endpointReader{}

	conn, _ = zmq.NewSocket(zmq.ROUTER)
	err := conn.Bind("tcp://" + address)
	if err != nil {
		log.Error("Connect bind error.......", err)
	}
	//fmt.Println("after bind " + address)
	go func() {
		edpReader.Receive(conn)
	}()
}

func Shutdonw() {
	edpReader.suspend(true)
	stopEndpointManager()
	stopActivatorActor()
	conn.Close()
}
