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

package vconfig

import (
	"encoding/hex"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
)

func TestPubkeyID(t *testing.T) {
	bookkeeper := "120202c924ed1a67fd1719020ce599d723d09d48362376836e04b0be72dfe825e24d81"
	pubKey, err := hex.DecodeString(bookkeeper)
	if err != nil {
		t.Errorf("DecodeString failed: %v", err)
	}
	k, err := keypair.DeserializePublicKey(pubKey)
	if err != nil {
		t.Errorf("DeserializePublicKey failed: %v", err)
	}
	nodeID := PubkeyID(k)
	t.Logf("res: %v\n", nodeID)
}

func TestPubkey(t *testing.T) {
	bookkeeper := "1202027df359dff69eea8dd7d807b669dd9635292b1aae97d03ed32cb36ff30fb7e4d9"
	pubKey, err := hex.DecodeString(bookkeeper)
	if err != nil {
		t.Errorf("DecodeString failed: %v", err)
	}
	k, err := keypair.DeserializePublicKey(pubKey)
	if err != nil {
		t.Errorf("DeserializePublicKey failed: %v", err)
	}
	nodeID := PubkeyID(k)
	publickey, err := Pubkey(nodeID)
	if err != nil {
		t.Errorf("Pubkey failed: %v", err)
	}
	t.Logf("res: %v", publickey)
}
