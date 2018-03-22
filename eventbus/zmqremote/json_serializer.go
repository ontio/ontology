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
	"bytes"
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"reflect"
)

type jsonSerializer struct {
	jsonpb.Marshaler
	jsonpb.Unmarshaler
}

func newJsonSerializer() Serializer {
	return &jsonSerializer{
		Marshaler: jsonpb.Marshaler{},
		Unmarshaler: jsonpb.Unmarshaler{
			AllowUnknownFields: true,
		},
	}
}

func (j *jsonSerializer) Serialize(msg interface{}) ([]byte, error) {
	if message, ok := msg.(*JsonMessage); ok {
		return []byte(message.Json), nil
	} else if message, ok := msg.(proto.Message); ok {

		str, err := j.Marshaler.MarshalToString(message)
		if err != nil {
			return nil, err
		}

		return []byte(str), nil
	}
	return nil, fmt.Errorf("msg must be proto.Message")
}

func (j *jsonSerializer) Deserialize(typeName string, b []byte) (interface{}, error) {
	protoType := proto.MessageType(typeName)
	if protoType == nil {
		m := &JsonMessage{
			TypeName: typeName,
			Json:     string(b),
		}
		return m, nil
	}
	t := protoType.Elem()

	intPtr := reflect.New(t)
	instance, ok := intPtr.Interface().(proto.Message)
	if ok {
		r := bytes.NewReader(b)
		j.Unmarshaler.Unmarshal(r, instance)

		return instance, nil
	} else {
		return nil, fmt.Errorf("msg must be proto.Message")
	}
}

func (j *jsonSerializer) GetTypeName(msg interface{}) (string, error) {
	if message, ok := msg.(*JsonMessage); ok {
		return message.TypeName, nil
	} else if message, ok := msg.(proto.Message); ok {
		typeName := proto.MessageName(message)

		return typeName, nil
	}

	return "", fmt.Errorf("msg must be proto.Message")
}
