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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
)

// DeployCode is an implementation of transaction payload for deploy smartcontract
type DeployCode struct {
	Code        []byte
	NeedStorage bool
	Name        string
	Version     string
	Author      string
	Email       string
	Description string

	OntVersion uint64
	Owner      common.Address
	AllShard   bool
	IsFrozen   bool
	ShardId    uint64

	address common.Address
}

func (dc *DeployCode) Address() common.Address {
	if dc.address == common.ADDRESS_EMPTY {
		dc.address = common.AddressFromVmCode(dc.Code)
	}
	return dc.address
}

func (dc *DeployCode) Serialize(w io.Writer) error {
	var err error

	err = serialization.WriteVarBytes(w, dc.Code)
	if err != nil {
		return fmt.Errorf("DeployCode Code Serialize failed: %s", err)
	}

	err = serialization.WriteBool(w, dc.NeedStorage)
	if err != nil {
		return fmt.Errorf("DeployCode NeedStorage Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Name)
	if err != nil {
		return fmt.Errorf("DeployCode Name Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Version)
	if err != nil {
		return fmt.Errorf("DeployCode Version Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Author)
	if err != nil {
		return fmt.Errorf("DeployCode Author Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Email)
	if err != nil {
		return fmt.Errorf("DeployCode Email Serialize failed: %s", err)
	}

	err = serialization.WriteString(w, dc.Description)
	if err != nil {
		return fmt.Errorf("DeployCode Description Serialize failed: %s", err)
	}

	if err = serialization.WriteUint64(w, dc.OntVersion); err != nil {
		return fmt.Errorf("DeployCode OntVersion Serialize failed: %s", err)
	}
	if err = dc.Owner.Serialize(w); err != nil {
		return fmt.Errorf("DeployCode Owner Serialize failed: %s", err)
	}
	if err = serialization.WriteBool(w, dc.AllShard); err != nil {
		return fmt.Errorf("DeployCode AllShard Serialize failed: %s", err)
	}
	if err = serialization.WriteBool(w, dc.IsFrozen); err != nil {
		return fmt.Errorf("DeployCode IsFrozen Serialize failed: %s", err)
	}
	if err = serialization.WriteUint64(w, dc.ShardId); err != nil {
		return fmt.Errorf("DeployCode shardId Serialize failed: %s", err)
	}

	return nil
}

func (dc *DeployCode) Deserialize(r io.Reader) error {
	code, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("DeployCode Code Deserialize failed: %s", err)
	}
	dc.Code = code

	dc.NeedStorage, err = serialization.ReadBool(r)
	if err != nil {
		return fmt.Errorf("DeployCode NeedStorage Deserialize failed: %s", err)
	}

	dc.Name, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Name Deserialize failed: %s", err)
	}

	dc.Version, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode CodeVersion Deserialize failed: %s", err)
	}

	dc.Author, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Author Deserialize failed: %s", err)
	}

	dc.Email, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Email Deserialize failed: %s", err)
	}

	dc.Description, err = serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("DeployCode Description Deserialize failed: %s", err)
	}

	return nil
}

func (dc *DeployCode) DeserializeForShard(r io.Reader) error {
	err := dc.Deserialize(r)
	if err != nil {
		return err
	}
	dc.OntVersion, err = serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("DeployCode OntVersion Deserialize failed: %s", err)
	}
	addr := &common.Address{}
	err = addr.Deserialize(r)
	if err != nil {
		return fmt.Errorf("DeployCode Owner Deserialize failed: %s", err)
	}
	dc.Owner = *addr
	dc.AllShard, err = serialization.ReadBool(r)
	if err != nil {
		return fmt.Errorf("DeployCode AllShard Deserialize failed: %s", err)
	}
	dc.IsFrozen, err = serialization.ReadBool(r)
	if err != nil {
		return fmt.Errorf("DeployCode IsFrozen Deserialize failed: %s", err)
	}
	dc.ShardId, err = serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("DeployCode ShardId Deserialize failed: %s", err)
	}

	return nil
}

func (dc *DeployCode) ToArray() []byte {
	sink := common.NewZeroCopySink(0)
	dc.Serialization(sink)
	return sink.Bytes()
}

func (dc *DeployCode) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(dc.Code)
	sink.WriteBool(dc.NeedStorage)
	sink.WriteString(dc.Name)
	sink.WriteString(dc.Version)
	sink.WriteString(dc.Author)
	sink.WriteString(dc.Email)
	sink.WriteString(dc.Description)
	sink.WriteUint64(dc.OntVersion)
	sink.WriteAddress(dc.Owner)
	sink.WriteBool(dc.AllShard)
	sink.WriteBool(dc.IsFrozen)
	sink.WriteUint64(dc.ShardId)
}

//note: DeployCode.Code has data reference of param source
func (dc *DeployCode) Deserialization(source *common.ZeroCopySource) error {
	var eof, irregular bool
	dc.Code, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}

	dc.NeedStorage, irregular, eof = source.NextBool()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Name, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Version, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Author, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Email, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	dc.Description, _, irregular, eof = source.NextString()
	if irregular {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}

	return nil
}

func (dc *DeployCode) DeserializationForShard(source *common.ZeroCopySource) error {
	err := dc.Deserialization(source)
	if err != nil {
		return err
	}
	var irr, eof bool
	dc.OntVersion, eof = source.NextUint64()
	dc.Owner, eof = source.NextAddress()
	if eof {
		return io.ErrUnexpectedEOF
	}
	dc.AllShard, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}
	dc.IsFrozen, irr, eof = source.NextBool()
	if irr {
		return common.ErrIrregularData
	}

	if eof {
		return io.ErrUnexpectedEOF
	}
	dc.ShardId, eof = source.NextUint64()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
