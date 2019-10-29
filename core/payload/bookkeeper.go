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

package payload

import (
	"fmt"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
)

const BookkeeperPayloadVersion byte = 0x00

type BookkeeperAction byte

const (
	BookkeeperAction_ADD BookkeeperAction = 0
	BookkeeperAction_SUB BookkeeperAction = 1
)

// Bookkeeper is an implementation of transaction payload for consensus bookkeeper list modification
type Bookkeeper struct {
	PubKey keypair.PublicKey
	Action BookkeeperAction
	Cert   []byte
	Issuer keypair.PublicKey
}

func (self *Bookkeeper) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(keypair.SerializePublicKey(self.PubKey))
	sink.WriteByte(byte(self.Action))
	sink.WriteVarBytes(self.Cert)
	sink.WriteVarBytes(keypair.SerializePublicKey(self.Issuer))
}
func (self *Bookkeeper) Deserialization(source *common.ZeroCopySource) error {
	pubKey, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	var err error
	self.PubKey, err = keypair.DeserializePublicKey(pubKey)
	if err != nil {
		return fmt.Errorf("[Bookkeeper], deserializing PubKey failed: %s", err)
	}
	action, eof := source.NextByte()
	if eof {
		return io.ErrUnexpectedEOF
	}
	self.Action = BookkeeperAction(action)
	cert, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	self.Cert = cert
	issuer, _, irregular, eof := source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	self.Issuer, err = keypair.DeserializePublicKey(issuer)
	if err != nil {
		return fmt.Errorf("[Bookkeeper], deserializing Issuer failed: %s", err)
	}
	return nil
}
