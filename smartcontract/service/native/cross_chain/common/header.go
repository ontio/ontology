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

package common

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
)

const (
	CURR_HEADER_VERSION = 0
)

type Header struct {
	Version          uint32
	ChainID          uint64
	PrevBlockHash    common.Uint256
	TransactionsRoot common.Uint256
	CrossStateRoot   common.Uint256
	BlockRoot        common.Uint256
	Timestamp        uint32
	Height           uint32
	ConsensusData    uint64
	ConsensusPayload []byte
	NextBookkeeper   common.Address

	Bookkeepers []keypair.PublicKey
	SigData     [][]byte

	hash *common.Uint256
}

func HeaderFromRawBytes(raw []byte) (*Header, error) {
	source := common.NewZeroCopySource(raw)
	header := &Header{}
	err := header.Deserialization(source)
	if err != nil {
		return nil, err
	}
	return header, nil
}

func (bd *Header) Serialization(sink *common.ZeroCopySink) {
	bd.serializationUnsigned(sink)
	sink.WriteVarUint(uint64(len(bd.Bookkeepers)))
	for _, v := range bd.Bookkeepers {
		sink.WriteVarBytes(keypair.SerializePublicKey(v))
	}
	sink.WriteVarUint(uint64(len(bd.SigData)))
	for _, sig := range bd.SigData {
		sink.WriteVarBytes(sig)
	}
}

func (bd *Header) serializationUnsigned(sink *common.ZeroCopySink) {
	if bd.Version > CURR_HEADER_VERSION {
		panic(fmt.Errorf("invalid header %d over max version:%d", bd.Version, CURR_HEADER_VERSION))
	}
	sink.WriteUint32(bd.Version)
	sink.WriteUint64(bd.ChainID)
	sink.WriteBytes(bd.PrevBlockHash[:])
	sink.WriteBytes(bd.TransactionsRoot[:])
	sink.WriteBytes(bd.CrossStateRoot[:])
	sink.WriteBytes(bd.BlockRoot[:])
	sink.WriteUint32(bd.Timestamp)
	sink.WriteUint32(bd.Height)
	sink.WriteUint64(bd.ConsensusData)
	sink.WriteVarBytes(bd.ConsensusPayload)
	sink.WriteBytes(bd.NextBookkeeper[:])
}

func (bd *Header) Hash() common.Uint256 {
	if bd.hash != nil {
		return *bd.hash
	}
	sink := common.NewZeroCopySink(nil)
	bd.serializationUnsigned(sink)
	temp := sha256.Sum256(sink.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))

	bd.hash = &hash
	return hash
}

func (bd *Header) Deserialization(source *common.ZeroCopySource) error {
	err := bd.deserializationUnsigned(source)
	if err != nil {
		return err
	}

	n, _, irr, eof := source.NextVarUint()
	if eof || irr {
		return errors.New("[Header] deserialize bookkeepers length error")
	}

	for i := 0; i < int(n); i++ {
		buf, _, irr, eof := source.NextVarBytes()
		if eof || irr {
			return errors.New("[Header] deserialize bookkeepers public key error")
		}
		pubkey, err := keypair.DeserializePublicKey(buf)
		if err != nil {
			return err
		}
		bd.Bookkeepers = append(bd.Bookkeepers, pubkey)
	}

	m, _, irr, eof := source.NextVarUint()
	if eof || irr {
		return errors.New("[Header] deserialize sigData length error")
	}

	for i := 0; i < int(m); i++ {
		sig, _, irr, eof := source.NextVarBytes()
		if eof || irr {
			return errors.New("[Header] deserialize sigData error")
		}
		bd.SigData = append(bd.SigData, sig)
	}

	return nil
}

func (bd *Header) deserializationUnsigned(source *common.ZeroCopySource) error {
	var eof, irr bool
	bd.Version, eof = source.NextUint32()
	if eof {
		return errors.New("[Header] read version error")
	}
	if bd.Version > CURR_HEADER_VERSION {
		return fmt.Errorf("[Header] header version %d over max version %d", bd.Version, CURR_HEADER_VERSION)
	}
	bd.ChainID, eof = source.NextUint64()
	if eof {
		return errors.New("[Header] read chainID error")
	}
	bd.PrevBlockHash, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read prevBlockHash error")
	}
	bd.TransactionsRoot, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read transactionsRoot error")
	}
	bd.CrossStateRoot, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read crossStatesRoot error")
	}
	bd.BlockRoot, eof = source.NextHash()
	if eof {
		return errors.New("[Header] read blockRoot error")
	}
	bd.Timestamp, eof = source.NextUint32()
	if eof {
		return errors.New("[Header] read timestamp error")
	}
	bd.Height, eof = source.NextUint32()
	if eof {
		return errors.New("[Header] read height error")
	}
	bd.ConsensusData, eof = source.NextUint64()
	if eof {
		return errors.New("[Header] read consensusData error")
	}
	bd.ConsensusPayload, _, irr, eof = source.NextVarBytes()
	if eof || irr {
		return errors.New("[Header] read consensusPayload eof error")
	}
	bd.NextBookkeeper, eof = source.NextAddress()
	if eof {
		return errors.New("[Header] read nextBookkeeper error")
	}
	return nil
}
