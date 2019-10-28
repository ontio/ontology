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

package states

import (
	"io"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/event"
)

// Invoke smart contract struct
// Param Version: invoke smart contract version, default 0
// Param Address: invoke on blockchain smart contract by address
// Param Method: invoke smart contract method, default ""
// Param Args: invoke smart contract arguments
type ContractInvokeParam struct {
	Version byte
	Address common.Address
	Method  string
	Args    []byte
}

func (this *ContractInvokeParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteByte(this.Version)
	sink.WriteAddress(this.Address)
	sink.WriteVarBytes([]byte(this.Method))
	sink.WriteVarBytes([]byte(this.Args))
}

// `ContractInvokeParam.Args` has reference of `source`
func (this *ContractInvokeParam) Deserialization(source *common.ZeroCopySource) error {
	var irregular, eof bool
	this.Version, eof = source.NextByte()
	this.Address, eof = source.NextAddress()
	var method []byte
	method, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	this.Method = string(method)

	this.Args, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type PreExecResult struct {
	State  byte
	Gas    uint64
	Result interface{}
	Notify []*event.NotifyEventInfo
}
