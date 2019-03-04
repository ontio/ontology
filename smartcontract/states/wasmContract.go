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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"io"
)

type WasmContractParam struct {
	Address common.Address
	Args    []byte
}

func (this *WasmContractParam) Serialize(w io.Writer) error {
	if err := this.Address.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractInvokeParam] Address serialize error!")
	}
	if err := serialization.WriteVarBytes(w, this.Args); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractInvokeParam] Args serialize error!")
	}
	return nil
}

func (this *WasmContractParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteAddress(this.Address)
	sink.WriteVarBytes([]byte(this.Args))
}

// Deserialize contract
func (this *WasmContractParam) Deserialize(r io.Reader) error {
	var err error

	if err := this.Address.Deserialize(r); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractInvokeParam] Address deserialize error!")
	}

	this.Args, err = serialization.ReadVarBytes(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[ContractInvokeParam] Args deserialize error!")
	}
	return nil
}

// `ContractInvokeParam.Args` has reference of `source`
func (this *WasmContractParam) Deserialization(source *common.ZeroCopySource) error {
	var irregular, eof bool
	this.Address, eof = source.NextAddress()

	this.Args, _, irregular, eof = source.NextVarBytes()
	if irregular {
		return common.ErrIrregularData
	}
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}
