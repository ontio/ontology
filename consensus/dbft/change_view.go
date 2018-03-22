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
	ser "github.com/Ontology/common/serialization"
	"io"
)

type ChangeView struct {
	msgData       ConsensusMessageData
	NewViewNumber byte
}

func (cv *ChangeView) Serialize(w io.Writer) error {
	cv.msgData.Serialize(w)
	w.Write([]byte{cv.NewViewNumber})
	return nil
}

//read data to reader
func (cv *ChangeView) Deserialize(r io.Reader) error {
	cv.msgData.Deserialize(r)
	viewNum, err := ser.ReadBytes(r, 1)
	if err != nil {
		return err
	}
	cv.NewViewNumber = viewNum[0]
	return nil
}

func (cv *ChangeView) Type() ConsensusMessageType {
	return cv.ConsensusMessageData().Type
}

func (cv *ChangeView) ViewNumber() byte {
	return cv.msgData.ViewNumber
}

func (cv *ChangeView) ConsensusMessageData() *ConsensusMessageData {
	return &(cv.msgData)
}
