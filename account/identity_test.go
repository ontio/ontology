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

package account

import (
	"encoding/hex"
	"testing"
)

var id = "did:ont:TSS6S4Xhzt5wtvRBTm4y3QCTRqB4BnU7vT"

func TestCreate(t *testing.T) {
	nonce, _ := hex.DecodeString("4c6b58adc6b8c6774eee0eb07dac4e198df87aae28f8932db3982edf3ff026e4")
	id1, err := CreateID(nonce)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("result ID:", id1)
	if id != id1 {
		t.Fatal("expected ID:", id)
	}
}

func TestVerify(t *testing.T) {
	t.Log("verify", id)
	if !VerifyID(id) {
		t.Error("error: failed")
	}

	invalid := []string{
		"did:ont:",
		"did:else:TSS6S4Xhzt5wtvRBTm4y3QCTRqB4BnU7vT",
		"TSS6S4Xhzt5wtvRBTm4y3QCTRqB4BnU7vT",
		"did:else:TSS6S4Xhzt5wtvRBTm4y3QCT",
		"did:ont:TSS6S4Xhzt5wtvRBTm4y3QCTRqB4BnU7vt",
	}

	for _, v := range invalid {
		t.Log("verify", v)
		if VerifyID(v) {
			t.Error("error: passed")
		}
	}
}
