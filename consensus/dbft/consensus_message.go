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

package dbft

import (
	"errors"
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

type ConsensusMessage interface {
	Serialization(sink *common.ZeroCopySink)
	Deserialization(source *common.ZeroCopySource) error
	Type() ConsensusMessageType
	ViewNumber() byte
	ConsensusMessageData() *ConsensusMessageData
}

type ConsensusMessageData struct {
	Type       ConsensusMessageType
	ViewNumber byte
}

func DeserializeMessage(data []byte) (ConsensusMessage, error) {
	if len(data) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	msgType := ConsensusMessageType(data[0])

	source := common.NewZeroCopySource(data)
	switch msgType {
	case PrepareRequestMsg:
		prMsg := &PrepareRequest{}
		err := prMsg.Deserialization(source)
		if err != nil {
			log.Error("[DeserializeMessage] PrepareRequestMsg Deserialize Error: ", err.Error())
			return nil, err
		}
		return prMsg, nil

	case PrepareResponseMsg:
		presMsg := &PrepareResponse{}
		err := presMsg.Deserialization(source)
		if err != nil {
			log.Error("[DeserializeMessage] PrepareResponseMsg Deserialize Error: ", err.Error())
			return nil, err
		}
		return presMsg, nil
	case ChangeViewMsg:
		cv := &ChangeView{}
		err := cv.Deserialization(source)
		if err != nil {
			log.Error("[DeserializeMessage] ChangeViewMsg Deserialize Error: ", err.Error())
			return nil, err
		}
		return cv, nil

	case BlockSignaturesMsg:
		blockSigs := &BlockSignatures{}
		err := blockSigs.Deserialization(source)
		if err != nil {
			log.Error("[DeserializeMessage] BlockSignaturesMsg Deserialize Error: ", err.Error())
			return nil, err
		}

		return blockSigs, nil
	}

	return nil, errors.New("The message is invalid.")
}

func (cd *ConsensusMessageData) Serialization(sink *common.ZeroCopySink) {
	sink.WriteByte(byte(cd.Type))
	sink.WriteByte(byte(cd.ViewNumber))
}

//read data to reader
func (cd *ConsensusMessageData) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	var temp byte
	temp, eof = source.NextByte()
	cd.Type = ConsensusMessageType(temp)
	cd.ViewNumber, eof = source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}
