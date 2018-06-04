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

package auth

import (
	"io"

	"github.com/ontio/ontology/common/serialization"
)

/*
 * each role is assigned an array of funcNames
 */
type roleFuncs struct {
	funcNames []string
}

func (this *roleFuncs) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.funcNames))); err != nil {
		return err
	}
	for _, fn := range this.funcNames {
		if err := serialization.WriteString(w, fn); err != nil {
			return err
		}
	}
	return nil
}

func (this *roleFuncs) Deserialize(rd io.Reader) error {
	var err error
	fnLen, err := serialization.ReadUint32(rd)
	if err != nil {
		return err
	}
	this.funcNames = make([]string, 0)
	for i := uint32(0); i < fnLen; i++ {
		fn, err := serialization.ReadString(rd)
		if err != nil {
			return err
		}
		this.funcNames = append(this.funcNames, fn)
	}
	return nil
}

type AuthToken struct {
	role       []byte
	expireTime uint32
	level      uint8
}

func (this *AuthToken) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.role); err != nil {
		return err
	}
	if err := serialization.WriteUint32(w, this.expireTime); err != nil {
		return err
	}
	if err := serialization.WriteUint8(w, this.level); err != nil {
		return err
	}
	return nil
}

func (this *AuthToken) Deserialize(rd io.Reader) error {
	//rd := bytes.NewReader(data)
	var err error
	this.role, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	this.expireTime, err = serialization.ReadUint32(rd)
	if err != nil {
		return err
	}
	this.level, err = serialization.ReadUint8(rd)
	if err != nil {
		return err
	}
	return nil
}

type DelegateStatus struct {
	root []byte
	AuthToken
}

func (this *DelegateStatus) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.root); err != nil {
		return err
	}
	if err := this.AuthToken.Serialize(w); err != nil {
		return err
	}
	return nil
}

func (this *DelegateStatus) Deserialize(rd io.Reader) error {
	var err error
	this.root, err = serialization.ReadVarBytes(rd)
	if err != nil {
		return err
	}
	err = this.AuthToken.Deserialize(rd)
	return err
}

type Status struct {
	status []*DelegateStatus
}

func (this *Status) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.status))); err != nil {
		return err
	}
	for _, s := range this.status {
		if err := s.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

func (this *Status) Deserialize(rd io.Reader) error {
	sLen, err := serialization.ReadUint32(rd)
	if err != nil {
		return err
	}
	this.status = make([]*DelegateStatus, 0)
	for i := uint32(0); i < sLen; i++ {
		s := new(DelegateStatus)
		err = s.Deserialize(rd)
		if err != nil {
			return err
		}
		this.status = append(this.status, s)
	}
	return nil
}

type roleTokens struct {
	tokens []*AuthToken
}

func (this *roleTokens) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.tokens))); err != nil {
		return err
	}
	for _, token := range this.tokens {
		if err := token.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

func (this *roleTokens) Deserialize(rd io.Reader) error {
	tLen, err := serialization.ReadUint32(rd)
	if err != nil {
		return err
	}
	this.tokens = make([]*AuthToken, 0)
	for i := uint32(0); i < tLen; i++ {
		tok := new(AuthToken)
		err = tok.Deserialize(rd)
		if err != nil {
			return err
		}
		this.tokens = append(this.tokens, tok)
	}
	return nil
}
