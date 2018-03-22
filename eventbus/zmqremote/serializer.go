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

var DefaultSerializerID int32 = 0
var serializers []Serializer

func init() {
	RegisterSerializer(newProtoSerializer())
	RegisterSerializer(newJsonSerializer())
}

func RegisterSerializerAsDefault(serializer Serializer) {
	serializers = append(serializers, serializer)
	DefaultSerializerID = int32(len(serializers) - 1)
}

func RegisterSerializer(serializer Serializer) {
	serializers = append(serializers, serializer)
}

type Serializer interface {
	Serialize(msg interface{}) ([]byte, error)
	Deserialize(typeName string, bytes []byte) (interface{}, error)
	GetTypeName(msg interface{}) (string, error)
}

func Serialize(message interface{}, serializerID int32) ([]byte, string, error) {
	res, err := serializers[serializerID].Serialize(message)
	typeName, err := serializers[serializerID].GetTypeName(message)
	return res, typeName, err
}

func Deserialize(message []byte, typeName string, serializerID int32) (interface{}, error) {
	return serializers[serializerID].Deserialize(typeName, message)
}
