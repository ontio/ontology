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

package oracle

import (
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/errors"
	"io"
)

type Status uint8

func (this *Status) Serialize(w io.Writer) error {
	if err := serialization.WriteUint8(w, uint8(*this)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize guaranty error!")
	}
	return nil
}

func (this *Status) Deserialize(r io.Reader) error {
	status, err := serialization.ReadUint8(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint. deserialize status error!")
	}
	*this = Status(status)
	return nil
}

type RegisterOracleNodeParam struct {
	Address  string `json:"address"`
	Guaranty uint64 `json:"guaranty"`
}

func (this *RegisterOracleNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize address error!")
	}
	if err := serialization.WriteUint64(w, this.Guaranty); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize guaranty error!")
	}
	return nil
}

func (this *RegisterOracleNodeParam) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	guaranty, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint. deserialize guaranty error!")
	}
	this.Guaranty = guaranty
	return nil
}

type ApproveOracleNodeParam struct {
	Address string `json:"address"`
}

func (this *ApproveOracleNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize address error!")
	}
	return nil
}

func (this *ApproveOracleNodeParam) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address
	return nil
}

type OracleNode struct {
	Address  string `json:"address"`
	Guaranty uint64 `json:"guaranty"`
	Status   Status `json:"status"`
}

func (this *OracleNode) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize address error!")
	}
	if err := serialization.WriteUint64(w, this.Guaranty); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize guaranty error!")
	}
	if err := this.Status.Serialize(w); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "this.Status.Serialize, serialize Status error!")
	}
	return nil
}

func (this *OracleNode) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	guaranty, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint. deserialize guaranty error!")
	}
	this.Guaranty = guaranty

	status := new(Status)
	err = status.Deserialize(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "status.Deserialize. deserialize status error!")
	}
	this.Status = *status
	return nil
}

type QuitOracleNodeParam struct {
	Address string `json:"address"`
}

func (this *QuitOracleNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, serialize address error!")
	}
	return nil
}

func (this *QuitOracleNodeParam) Deserialize(r io.Reader) error {
	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address
	return nil
}

type CreateOracleRequestParam struct {
	Request    string `json:"request"`
	OracleNode string `json:"oracleNode"`
	Address    string `json:"address"`
	Fee        uint64 `json:"fee"`
}

func (this *CreateOracleRequestParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.Request); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, request address error!")
	}
	if err := serialization.WriteString(w, this.OracleNode); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, oracleNode address error!")
	}
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, address address error!")
	}
	if err := serialization.WriteUint64(w, this.Fee); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteVarUint, serialize fee error!")
	}
	return nil
}

func (this *CreateOracleRequestParam) Deserialize(r io.Reader) error {
	request, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize request error!")
	}
	this.Request = request

	oracleNode, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize oracleNode error!")
	}
	this.OracleNode = oracleNode

	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	fee, err := serialization.ReadUint64(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadVarUint. deserialize fee error!")
	}
	this.Fee = fee
	return nil
}

type UndoRequests struct {
	Requests map[string]struct{} `json:"requests"`
}

type SetOracleOutcomeParam struct {
	TxHash  string `json:"txHash"`
	Address string `json:"address"`
	Outcome string `json:"outcome"`
}

func (this *SetOracleOutcomeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.TxHash); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize txHash error!")
	}
	if err := serialization.WriteString(w, this.Address); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize address error!")
	}
	if err := serialization.WriteString(w, this.Outcome); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.WriteString, deserialize outcome error!")
	}
	return nil
}

func (this *SetOracleOutcomeParam) Deserialize(r io.Reader) error {
	txHash, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize txHash error!")
	}
	this.TxHash = txHash

	address, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize address error!")
	}
	this.Address = address

	outcome, err := serialization.ReadString(r)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "serialization.ReadString, deserialize outcome error!")
	}
	this.Outcome = outcome
	return nil
}
