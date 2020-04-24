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
	"crypto/sha256"
	"fmt"

	"github.com/ontio/ontology/common"
)

const (
	CURR_CROSS_STATES_VERSION = 0
)

type CrossChainMsg struct {
	Version    byte
	Height     uint32
	StatesRoot common.Uint256

	SigData [][]byte

	hash *common.Uint256
}

func (this *CrossChainMsg) serializationUnsigned(sink *common.ZeroCopySink) {
	sink.WriteByte(this.Version)
	sink.WriteUint32(this.Height)
	sink.WriteBytes(this.StatesRoot[:])
}

func (this *CrossChainMsg) Serialization(sink *common.ZeroCopySink) {
	this.serializationUnsigned(sink)
	sink.WriteVarUint(uint64(len(this.SigData)))
	for _, sig := range this.SigData {
		sink.WriteVarBytes(sig)
	}
}

func (this *CrossChainMsg) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.Version, eof = source.NextByte()
	if eof {
		return fmt.Errorf("CrossChainMsg, deserialization read version error")
	}
	this.Height, eof = source.NextUint32()
	if eof {
		return fmt.Errorf("CrossChainMsg, deserialization read height error")
	}
	this.StatesRoot, eof = source.NextHash()
	if eof {
		return fmt.Errorf("CrossChainMsg, deserialization read statesRoot error")
	}
	sigLen, _, irr, eof := source.NextVarUint()
	if irr || eof {
		return fmt.Errorf("CrossChainMsg, deserialization read sigData lenght error")
	}
	sigData := make([][]byte, 0, sigLen)
	for i := 0; i < int(sigLen); i++ {
		v, _, irr, eof := source.NextVarBytes()
		if irr || eof {
			return fmt.Errorf("CrossChainMsg, deserialization read sigData value error")
		}
		sigData = append(sigData, v)
	}
	this.SigData = sigData
	return nil
}

func (this *CrossChainMsg) Hash() common.Uint256 {
	if this.hash != nil {
		return *this.hash
	}
	sink := common.NewZeroCopySink(nil)
	this.serializationUnsigned(sink)
	temp := sha256.Sum256(sink.Bytes())
	hash := common.Uint256(sha256.Sum256(temp[:]))
	this.hash = &hash
	return hash
}

func (this *CrossChainMsg) SetHash(hash common.Uint256) {
	this.hash = &hash
}
