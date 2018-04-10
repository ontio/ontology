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
	"errors"
	"io"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/signature"
)

type ConsensusPayload struct {
	Version         uint32
	PrevHash        common.Uint256
	Height          uint32
	BookkeeperIndex uint16
	Timestamp       uint32
	Data            []byte
	Owner           keypair.PublicKey
	Signature       []byte
	hash            common.Uint256
}

//get the consensus payload hash
func (cp *ConsensusPayload) Hash() common.Uint256 {
	return common.Uint256{}
}

//Check whether header is correct
func (cp *ConsensusPayload) Verify() error {
	buf := new(bytes.Buffer)
	cp.SerializeUnsigned(buf)

	err := signature.Verify(cp.Owner, buf.Bytes(), cp.Signature)

	return err
}

//serialize the consensus payload
func (cp *ConsensusPayload) ToArray() []byte {
	b := new(bytes.Buffer)
	cp.Serialize(b)
	return b.Bytes()
}

//return inventory type
func (cp *ConsensusPayload) InventoryType() common.InventoryType {
	return common.CONSENSUS
}

func (cp *ConsensusPayload) GetMessage() []byte {
	//TODO: GetMessage
	//return sig.GetHashData(cp)
	return []byte{}
}

func (cp *ConsensusPayload) Type() common.InventoryType {

	//TODO:Temporary add for Interface signature.SignableData use.
	return common.CONSENSUS
}

//Serialize message payload
func (cp *ConsensusPayload) Serialize(w io.Writer) error {
	err := cp.SerializeUnsigned(w)
	if err != nil {
		return err
	}
	buf := keypair.SerializePublicKey(cp.Owner)
	err = serialization.WriteVarBytes(w, buf)
	if err != nil {
		return err
	}

	err = serialization.WriteVarBytes(w, cp.Signature)
	if err != nil {
		return err
	}

	return err
}

//Deserialize message payload
func (cp *ConsensusPayload) Deserialize(r io.Reader) error {
	err := cp.DeserializeUnsigned(r)

	buf, err := serialization.ReadVarBytes(r)
	if err != nil {
		log.Warn("Consensus item Owner deserialize failed, " + err.Error())
		return errors.New("Consensus item Owner deserialize failed.")
	}
	cp.Owner, err = keypair.DeserializePublicKey(buf)
	if err != nil {
		log.Warn("Consensus item Owner deserialize failed, " + err.Error())
		return errors.New("Consensus item Owner deserialize failed.")
	}

	cp.Signature, err = serialization.ReadVarBytes(r)
	if err != nil {
		log.Warn("Consensus item Signature deserialize failed, " + err.Error())
		return errors.New("Consensus item Signature deserialize failed.")
	}

	return err
}

//Serialize message payload
func (cp *ConsensusPayload) SerializeUnsigned(w io.Writer) error {
	serialization.WriteUint32(w, cp.Version)
	cp.PrevHash.Serialize(w)
	serialization.WriteUint32(w, cp.Height)
	serialization.WriteUint16(w, cp.BookkeeperIndex)
	serialization.WriteUint32(w, cp.Timestamp)
	serialization.WriteVarBytes(w, cp.Data)
	return nil
}

//Deserialize message payload
func (cp *ConsensusPayload) DeserializeUnsigned(r io.Reader) error {
	var err error
	cp.Version, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Version Deserialize failed.")
		return errors.New("consensus item Version Deserialize failed. ")
	}

	preBlock := new(common.Uint256)
	err = preBlock.Deserialize(r)
	if err != nil {
		log.Warn("consensus item preHash Deserialize failed.")
		return errors.New("consensus item preHash Deserialize failed. ")
	}
	cp.PrevHash = *preBlock

	cp.Height, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Height Deserialize failed.")
		return errors.New("consensus item Height Deserialize failed. ")
	}

	cp.BookkeeperIndex, err = serialization.ReadUint16(r)
	if err != nil {
		log.Warn("consensus item BookKeeperIndex Deserialize failed.")
		return errors.New("consensus item BookKeeperIndex Deserialize failed. ")
	}

	cp.Timestamp, err = serialization.ReadUint32(r)
	if err != nil {
		log.Warn("consensus item Timestamp Deserialize failed.")
		return errors.New("consensus item Timestamp Deserialize failed. ")
	}

	cp.Data, err = serialization.ReadVarBytes(r)
	if err != nil {
		log.Warn("consensus item Data Deserialize failed.")
		return errors.New("consensus item Data Deserialize failed. ")
	}

	return nil
}
