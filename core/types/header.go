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

package types

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

type Header struct {
	Version          uint32
	PrevBlockHash    common.Uint256
	TransactionsRoot common.Uint256
	BlockRoot        common.Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload []byte
	NextBookkeeper   common.Address

	//Program *program.Program
	Bookkeepers []keypair.PublicKey
	SigData     [][]byte

	hash *common.Uint256
}

//Serialize the blockheader
func (self *Header) Serialize(w io.Writer) error {
	self.SerializeUnsigned(w)

	err := serialization.WriteVarUint(w, uint64(len(self.Bookkeepers)))
	if err != nil {
		return errors.New("serialize sig pubkey length failed")
	}
	for _, pubkey := range self.Bookkeepers {
		err := serialization.WriteVarBytes(w, keypair.SerializePublicKey(pubkey))
		if err != nil {
			return err
		}
	}

	err = serialization.WriteVarUint(w, uint64(len(self.SigData)))
	if err != nil {
		return errors.New("serialize sig pubkey length failed")
	}

	for _, sig := range self.SigData {
		err = serialization.WriteVarBytes(w, sig)
		if err != nil {
			return err
		}
	}

	return nil
}

//Serialize the blockheader data without program
func (self *Header) SerializeUnsigned(w io.Writer) error {
	err := serialization.WriteUint32(w, self.Version)
	if err != nil {
		return err
	}
	err = self.PrevBlockHash.Serialize(w)
	if err != nil {
		return err
	}
	err = self.TransactionsRoot.Serialize(w)
	if err != nil {
		return err
	}
	err = self.BlockRoot.Serialize(w)
	if err != nil {
		return err
	}
	err = serialization.WriteUint32(w, self.Timestamp)
	if err != nil {
		return err
	}
	err = serialization.WriteUint32(w, self.Height)
	if err != nil {
		return err
	}
	err = serialization.WriteUint64(w, self.ConsensusData)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, self.ConsensusPayload)
	if err != nil {
		return err
	}
	err = self.NextBookkeeper.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (self *Header) Deserialize(reader io.Reader) error {
	hasher := sha256.New()
	r := io.TeeReader(reader, hasher)

	err := self.DeserializeUnsigned(r)
	if err != nil {
		return err
	}

	temp := hasher.Sum(nil)
	f := common.Uint256(sha256.Sum256(temp[:]))
	self.hash = &f
	//reset r back to origin reader, it is ok because TeeReader has no internal bufffer
	r = reader

	n, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}

	for i := 0; i < int(n); i++ {
		buf, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		pubkey, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		self.Bookkeepers = append(self.Bookkeepers, pubkey)
	}

	m, err := serialization.ReadVarUint(r, 0)
	if err != nil {
		return err
	}

	for i := 0; i < int(m); i++ {
		sig, err := serialization.ReadVarBytes(r)
		if err != nil {
			return err
		}
		self.SigData = append(self.SigData, sig)
	}

	return nil
}

func (self *Header) DeserializeUnsigned(r io.Reader) error {
	var err error
	self.Version, err = serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("Header item Version Deserialize failed: %s", err)
	}

	err = self.PrevBlockHash.Deserialize(r)
	if err != nil {
		return fmt.Errorf("Header item preBlock Deserialize failed: %s", err)
	}

	err = self.TransactionsRoot.Deserialize(r)
	if err != nil {
		return err
	}

	err = self.BlockRoot.Deserialize(r)
	if err != nil {
		return err
	}

	self.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}

	self.Height, err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}

	self.ConsensusData, err = serialization.ReadUint64(r)
	if err != nil {
		return err
	}

	self.ConsensusPayload, err = serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}

	err = self.NextBookkeeper.Deserialize(r)

	return err
}

func (self *Header) Hash() common.Uint256 {
	if self.hash != nil {
		return *self.hash
	}
	buf := new(bytes.Buffer)
	self.SerializeUnsigned(buf)
	temp := sha256.Sum256(buf.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	self.hash = &hash
	return hash
}

func (self *Header) GetMessage() []byte {
	bf := new(bytes.Buffer)
	self.SerializeUnsigned(bf)
	return bf.Bytes()
}

func (self *Header) ToArray() []byte {
	bf := new(bytes.Buffer)
	self.Serialize(bf)
	return bf.Bytes()
}
