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
	"io"

	ser "github.com/ontio/ontology/common/serialization"
)

type PrepareResponse struct {
	msgData   ConsensusMessageData
	Signature []byte
}

func (pres *PrepareResponse) Serialize(w io.Writer) error {
	pres.msgData.Serialize(w)
	ser.WriteVarBytes(w, pres.Signature)
	return nil
}

//read data to reader
func (pres *PrepareResponse) Deserialize(r io.Reader) error {
	err := pres.msgData.Deserialize(r)
	if err != nil {
		return err
	}
	// Fixme the 64 should be defined as a unified const
	pres.Signature, err = ser.ReadVarBytes(r)
	if err != nil {
		return err
	}
	return nil

}

func (pres *PrepareResponse) Type() ConsensusMessageType {
	return pres.ConsensusMessageData().Type
}

func (pres *PrepareResponse) ViewNumber() byte {
	return pres.msgData.ViewNumber
}

func (pres *PrepareResponse) ConsensusMessageData() *ConsensusMessageData {
	return &(pres.msgData)
}
