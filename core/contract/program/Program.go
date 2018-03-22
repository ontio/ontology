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

package program

import (
	"github.com/Ontology/common/serialization"
	. "github.com/Ontology/errors"
	"io"
)

type Program struct {
	//the contract program code,which will be run on VM or specific envrionment
	Code      []byte

	//the program code's parameter
	Parameter []byte
}

//Serialize the Program
func (p *Program) Serialize(w io.Writer) error {
	err := serialization.WriteVarBytes(w, p.Parameter)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Serialize Code failed.")
	}
	err = serialization.WriteVarBytes(w, p.Code)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Serialize Parameter failed.")
	}

	return nil
}

//Deserialize the Program
func (p *Program) Deserialize(w io.Reader) error {
	val, err := serialization.ReadVarBytes(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Deserialize Parameter failed.")
	}
	p.Parameter = val
	p.Code, err = serialization.ReadVarBytes(w)
	if err != nil {
		return NewDetailErr(err, ErrNoCode, "Execute Program Deserialize Code failed.")
	}
	return nil
}
