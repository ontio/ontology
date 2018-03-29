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
	"fmt"
	"io"

	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	ser "github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
)

type PrepareRequest struct {
	msgData        ConsensusMessageData
	Nonce          uint64
	NextBookkeeper common.Address
	Transactions   []*types.Transaction
	Signature      []byte
}

func (pr *PrepareRequest) Serialize(w io.Writer) error {
	log.Debug()

	pr.msgData.Serialize(w)
	if err := ser.WriteVarUint(w, pr.Nonce); err != nil {
		return fmt.Errorf("[PrepareRequest] nonce serialization failed: %s", err)
	}
	if err := pr.NextBookkeeper.Serialize(w); err != nil {
		return fmt.Errorf("[PrepareRequest] nextbookkeeper serialization failed: %s", err)
	}
	if err := ser.WriteVarUint(w, uint64(len(pr.Transactions))); err != nil {
		return fmt.Errorf("[PrepareRequest] length serialization failed: %s", err)
	}
	for _, t := range pr.Transactions {
		if err := t.Serialize(w); err != nil {
			return fmt.Errorf("[PrepareRequest] transactions serialization failed: %s", err)
		}
	}
	if err := ser.WriteVarBytes(w, pr.Signature); err != nil {
		return fmt.Errorf("[PrepareRequest] signature serialization failed: %s", err)
	}
	return nil
}

func (pr *PrepareRequest) Deserialize(r io.Reader) error {
	pr.msgData = ConsensusMessageData{}
	pr.msgData.Deserialize(r)
	pr.Nonce, _ = ser.ReadVarUint(r, 0)

	if err := pr.NextBookkeeper.Deserialize(r); err != nil {
		return fmt.Errorf("[PrepareRequest] nextbookkeeper deserialization failed: %s", err)
	}

	length, err := ser.ReadVarUint(r, 0)
	if err != nil {
		return fmt.Errorf("[PrepareRequest] length deserialization failed: %s", err)
	}

	pr.Transactions = make([]*types.Transaction, length)
	for i := 0; i < len(pr.Transactions); i++ {
		var t types.Transaction
		if err := t.Deserialize(r); err != nil {
			return fmt.Errorf("[PrepareRequest] transactions deserialization failed: %s", err)
		}
		pr.Transactions[i] = &t
	}

	pr.Signature, err = ser.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("[PrepareRequest] signature deserialization failed: %s", err)
	}

	return nil
}

func (pr *PrepareRequest) Type() ConsensusMessageType {
	log.Debug()
	return pr.ConsensusMessageData().Type
}

func (pr *PrepareRequest) ViewNumber() byte {
	log.Debug()
	return pr.msgData.ViewNumber
}

func (pr *PrepareRequest) ConsensusMessageData() *ConsensusMessageData {
	log.Debug()
	return &(pr.msgData)
}
