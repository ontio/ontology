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

func TestSer_roleFuncs(t *testing.T) {
	param := &roleFuncs{
		[]string{"foo1", "foo2"},
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

func TestSer_AuthToken(t *testing.T) {
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

func TestSer_DelegateStatus(t *testing.T) {
	/*
		s1 := &DelegateStatus {
			root: []byte{0x01, 0x02, 0x03, 0x04, 0x05},

		}
	*/
}
func TestSer_roleAuthTokens(t *testing.T) {
	token1 := &AuthToken{
		role:       []byte("role"),
		expireTime: 1000000,
		level:      2,
	}
	token2 := &AuthToken{
		role:       []byte("role2"),
		expireTime: 10000,
		level:      2,
	}

	tokens := &roleTokens{
		tokens: []*AuthToken{token1, token2},
	}
	bf := new(bytes.Buffer)
	if err := tokens.Serialize(bf); err != nil {
		t.Fatal(err)
	}

	tokens2 := new(roleTokens)
	rd := bytes.NewReader(bf.Bytes())
	if err := tokens2.Deserialize(rd); err != nil {
		t.Fatal(err)
	}

	if len(tokens.tokens) != len(tokens2.tokens) {
		t.Fatalf("failed")
	}
}

/*
func BenchmarkDes(b *testing.B) {
	token1 := &AuthToken{
		role:       []byte("role"),
		expireTime: 1000000,
		level:      2,
	}
	token2 := &AuthToken{
		role:       []byte("role2"),
		expireTime: 10000,
		level:      2,
	}
	tokens := &roleTokens{
		tokens: []*AuthToken{token1, token2},
	}
	bf := new(bytes.Buffer)
	if err := tokens.Serialize(bf); err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		tokens2 := new(roleTokens)
		rd := bytes.NewReader(bf.Bytes())
		if err := tokens2.Deserialize(rd); err != nil {
			b.Fatal(err)
		}
	}
}
*/
