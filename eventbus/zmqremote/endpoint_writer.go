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
	"time"

	"github.com/Ontology/common/log"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/eventstream"
	zmq "github.com/pebbe/zmq4"
)

func newEndpointWriter(address string) actor.Producer {
	return func() actor.Actor {
		return &endpointWriter{
			address: address,
			//config:  config,
		}
	}
}

type endpointWriter struct {
	//config              *remoteConfig
	address string
	conn    *zmq.Socket
	//stream              Remoting_ReceiveClient
	defaultSerializerId int32
}

func (state *endpointWriter) initialize() {
	err := state.initializeInternal()
	if err != nil {
		time.Sleep(2 * time.Second)
		panic(err)
	}
}

func (state *endpointWriter) initializeInternal() error {

	state.conn, _ = zmq.NewSocket(zmq.DEALER)
	err := state.conn.Connect("tcp://" + state.address)
	if err != nil {
		log.Error("error while connect ", state.address, err.Error())
		return err
	}

	go func() {
		connected := &EndpointConnectedEvent{Address: state.address}
		eventstream.Publish(connected)
	}()
	return nil
}

func (state *endpointWriter) sendEnvelopes(msg []interface{}, ctx actor.Context) {

	envelopes := make([]*MessageEnvelope, len(msg))

	//type name uniqueness map name string to type index
	typeNames := make(map[string]int32)
	typeNamesArr := make([]string, 0)
	targetNames := make(map[string]int32)
	targetNamesArr := make([]string, 0)
	var header *MessageHeader
	var typeID int32
	var targetID int32
	var serializerID int32

	for i, tmp := range msg {

		rd := tmp.(*remoteDeliver)

		if rd.serializerID == -1 {
			serializerID = state.defaultSerializerId
		} else {
			serializerID = rd.serializerID
		}

		if rd.header == nil || rd.header.Length() == 0 {
			header = nil
		} else {
			header = &MessageHeader{rd.header.ToMap()}
		}

		bytes, typeName, err := Serialize(rd.message, serializerID)
		if err != nil {
			log.Error("serialize error:", err.Error())
			panic(err)
		}
		typeID, typeNamesArr = addToLookup(typeNames, typeName, typeNamesArr)
		targetID, targetNamesArr = addToLookup(targetNames, rd.target.Id, targetNamesArr)

		envelopes[i] = &MessageEnvelope{
			MessageHeader: header,
			MessageData:   bytes,
			Sender:        rd.sender,
			Target:        targetID,
			TypeId:        typeID,
			SerializerId:  serializerID,
		}
	}

	batch := &MessageBatch{
		TypeNames:   typeNamesArr,
		TargetNames: targetNamesArr,
		Envelopes:   envelopes,
	}

	batchstr, _, _ := Serialize(batch, serializerID)

	_, err := state.conn.Send(string(batchstr), 0)

	if err != nil {
		ctx.Stash()
		panic("restart it")
	}
}

func addToLookup(m map[string]int32, name string, a []string) (int32, []string) {
	max := int32(len(m))
	id, ok := m[name]
	if !ok {
		m[name] = max
		id = max
		a = append(a, name)
	}
	return id, a
}

func (state *endpointWriter) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		state.initialize()
	case *actor.Stopped:
		state.conn.Close()
	case *actor.Restarting:
		state.conn.Close()
	case []interface{}:
		state.sendEnvelopes(msg, ctx)
	case actor.SystemMessage, actor.AutoReceiveMessage:
		//ignore
	default:
		//plog.Error("EndpointWriter received unknown message", log.String("address", state.address), log.TypeOf("type", msg), log.Message(msg))
	}
}
