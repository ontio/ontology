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
func (bd *Header) Serialize(w io.Writer) error {
	bd.SerializeUnsigned(w)

	err := serialization.WriteVarUint(w, uint64(len(bd.Bookkeepers)))
	if err != nil {
		return errors.New("serialize sig pubkey length failed")
	}
	for _, pubkey := range bd.Bookkeepers {
		err := serialization.WriteVarBytes(w, keypair.SerializePublicKey(pubkey))
		if err != nil {
			return err
		}
	}

	err = serialization.WriteVarUint(w, uint64(len(bd.SigData)))
	if err != nil {
		return errors.New("serialize sig pubkey length failed")
	}

	for _, sig := range bd.SigData {
		err = serialization.WriteVarBytes(w, sig)
		if err != nil {
			return err
		}
	}

	return nil
}

//Serialize the blockheader data without program
func (bd *Header) SerializeUnsigned(w io.Writer) error {
	err := serialization.WriteUint32(w, bd.Version)
	if err != nil {
		return err
	}
	err = bd.PrevBlockHash.Serialize(w)
	if err != nil {
		return err
	}
	err = bd.TransactionsRoot.Serialize(w)
	if err != nil {
		return err
	}
	err = bd.BlockRoot.Serialize(w)
	if err != nil {
		return err
	}
	err = serialization.WriteUint32(w, bd.Timestamp)
	if err != nil {
		return err
	}
	err = serialization.WriteUint32(w, bd.Height)
	if err != nil {
		return err
	}
	err = serialization.WriteUint64(w, bd.ConsensusData)
	if err != nil {
		return err
	}
	err = serialization.WriteVarBytes(w, bd.ConsensusPayload)
	if err != nil {
		return err
	}
	err = bd.NextBookkeeper.Serialize(w)
	if err != nil {
		return err
	}
	return nil
}

func (bd *Header) Deserialize(r io.Reader) error {
	err := bd.DeserializeUnsigned(r)
	if err != nil {
		return err
	}

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
		bd.Bookkeepers = append(bd.Bookkeepers, pubkey)
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
		bd.SigData = append(bd.SigData, sig)
	}

	return nil
}

func (bd *Header) DeserializeUnsigned(r io.Reader) error {
	var err error
	bd.Version, err = serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("Header item Version Deserialize failed: %s", err)
	}

	err = bd.PrevBlockHash.Deserialize(r)
	if err != nil {
		return fmt.Errorf("Header item preBlock Deserialize failed: %s", err)
	}

	err = bd.TransactionsRoot.Deserialize(r)
	if err != nil {
		return err
	}

	err = bd.BlockRoot.Deserialize(r)
	if err != nil {
		return err
	}

	bd.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}

	bd.Height, err = serialization.ReadUint32(r)
	if err != nil {
		return err
	}

	bd.ConsensusData, err = serialization.ReadUint64(r)
	if err != nil {
		return err
	}

	bd.ConsensusPayload, err = serialization.ReadVarBytes(r)
	if err != nil {
		return err
	}

	err = bd.NextBookkeeper.Deserialize(r)

	return err
}

func (bd *Header) Hash() common.Uint256 {
	if bd.hash != nil {
		return *bd.hash
	}
	buf := new(bytes.Buffer)
	bd.SerializeUnsigned(buf)
	temp := sha256.Sum256(buf.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	bd.hash = &hash
	return hash
}

func (bd *Header) GetMessage() []byte {
	bf := new(bytes.Buffer)
	bd.SerializeUnsigned(bf)
	return bf.Bytes()
}

func (bd *Header) ToArray() []byte {
	bf := new(bytes.Buffer)
	bd.Serialize(bf)
	return bf.Bytes()
}
