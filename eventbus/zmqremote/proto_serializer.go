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
	"fmt"
	"github.com/gogo/protobuf/proto"
	"reflect"
)

type protoSerializer struct{}

func newProtoSerializer() Serializer {
	return &protoSerializer{}
}

func (protoSerializer) Serialize(msg interface{}) ([]byte, error) {
	if message, ok := msg.(proto.Message); ok {
		bytes, err := proto.Marshal(message)
		if err != nil {
			return nil, err
		}

		return bytes, nil
	}
	return nil, fmt.Errorf("msg must be proto.Message")
}

func (protoSerializer) Deserialize(typeName string, bytes []byte) (interface{}, error) {
	protoType := proto.MessageType(typeName)
	if protoType == nil {
		return nil, fmt.Errorf("Unknown message type %v", typeName)
	}
	t := protoType.Elem()

	intPtr := reflect.New(t)
	instance := intPtr.Interface().(proto.Message)
	proto.Unmarshal(bytes, instance)

	return instance, nil
}

func (protoSerializer) GetTypeName(msg interface{}) (string, error) {
	if message, ok := msg.(proto.Message); ok {
		typeName := proto.MessageName(message)

		return typeName, nil
	}
	return "", fmt.Errorf("msg must be proto.Message")
}
