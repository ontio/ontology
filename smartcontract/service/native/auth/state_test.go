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
	"bytes"
	"testing"
)

func TestSerRoleFuncs(t *testing.T) {
	param := &roleFuncs{
		[]string{"foo1", "foo2"},
		//[]string{},
	}
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(roleFuncs)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if len(param.funcNames) != len(param2.funcNames) {
		t.Fatalf("does not match")
	}
	for i := 0; i < len(param.funcNames); i++ {
		if param.funcNames[i] != param2.funcNames[i] {
			t.Fatalf("%s \t %s does not match", param.funcNames[i], param2.funcNames[i])
		}
	}
}

func TestSerAuthToken(t *testing.T) {
	param := &AuthToken{
		role:       []byte("role"),
		expireTime: 1000000,
		level:      2,
	}

	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	param2 := new(AuthToken)
	if err := param2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if param.expireTime != param2.expireTime ||
		param.level != param2.level ||
		bytes.Compare(param.role, param2.role) != 0 {
		t.Fatalf("failed")
	}
}

func TestSerDelegateStatus(t *testing.T) {
	token := &AuthToken{
		role:       []byte("role"),
		expireTime: 1000000,
		level:      2,
	}
	s1 := &DelegateStatus{
		root:      []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		AuthToken: *token,
	}
	bf := new(bytes.Buffer)
	if err := s1.Serialize(bf); err != nil {
		t.Fatal(err)
	}
	rd := bytes.NewReader(bf.Bytes())
	s2 := new(DelegateStatus)
	if err := s2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(s1.root, s2.root) != 0 ||
		bytes.Compare(s1.role, s2.role) != 0 ||
		s1.expireTime != s2.expireTime || s1.level != s2.level {
		t.Fatalf("failed")
	}
}
