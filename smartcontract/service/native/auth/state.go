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
	"strings"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

/*
 * each role is assigned an array of funcNames
 */
type roleFuncs struct {
	funcNames []string
}

func (this *roleFuncs) AppendFuncs(fns []string) {
	funcNames := append(this.funcNames, fns...)
	this.funcNames = StringsDedupAndSort(funcNames)
}

func (this *roleFuncs) ContainsFunc(fn string) bool {
	for _, f := range this.funcNames {
		if strings.Compare(fn, f) == 0 {
			return true
		}
	}

	return false
}

func (this *roleFuncs) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.funcNames)))

	this.funcNames = StringsDedupAndSort(this.funcNames)
	for _, fn := range this.funcNames {
		sink.WriteString(fn)
	}
}

func (this *roleFuncs) Deserialization(source *common.ZeroCopySource) error {
	fnLen, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	funcNames := make([]string, 0)
	for i := uint32(0); i < fnLen; i++ {
		fn, err := utils.DecodeString(source)
		if err != nil {
			return err
		}
		funcNames = append(funcNames, fn)
	}

	this.funcNames = StringsDedupAndSort(funcNames)

	return nil
}

type AuthToken struct {
	role       []byte
	expireTime uint32
	level      uint8
}

func (this *AuthToken) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.role)
	sink.WriteUint32(this.expireTime)
	sink.WriteUint8(this.level)
}

func (this *AuthToken) Deserialization(source *common.ZeroCopySource) error {
	//rd := bytes.NewReader(data)
	var err error
	this.role, err = utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	var eof bool
	this.expireTime, eof = source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.level, eof = source.NextUint8()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type DelegateStatus struct {
	root []byte
	AuthToken
}

func (this *DelegateStatus) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.root)
	this.AuthToken.Serialization(sink)
}

func (this *DelegateStatus) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.root, err = utils.DecodeVarBytes(source)
	if err != nil {
		return err
	}
	err = this.AuthToken.Deserialization(source)
	return err
}

type Status struct {
	status []*DelegateStatus
}

func (this *Status) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.status)))
	for _, s := range this.status {
		s.Serialization(sink)
	}
}

func (this *Status) Deserialization(source *common.ZeroCopySource) error {
	sLen, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.status = make([]*DelegateStatus, 0)
	for i := uint32(0); i < sLen; i++ {
		s := new(DelegateStatus)
		err := s.Deserialization(source)
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

func (this *roleTokens) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(uint32(len(this.tokens)))
	for _, token := range this.tokens {
		token.Serialization(sink)
	}
}

func (this *roleTokens) Deserialization(source *common.ZeroCopySource) error {
	tLen, eof := source.NextUint32()
	if eof {
		return io.ErrUnexpectedEOF
	}
	this.tokens = make([]*AuthToken, 0)
	for i := uint32(0); i < tLen; i++ {
		tok := new(AuthToken)
		err := tok.Deserialization(source)
		if err != nil {
			return err
		}
		this.tokens = append(this.tokens, tok)
	}
	return nil
}
