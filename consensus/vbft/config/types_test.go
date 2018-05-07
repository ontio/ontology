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

func TestStringID(t *testing.T) {
	nodeid := "120202890c587f4e4a6a98b455248eabac04b733580cfe5f11acd648c675543dfbb926"
	nodeID, err := StringID(nodeid)
	if err != nil {
		t.Errorf("test failed: %v", err)
	} else {
		t.Logf("test succ: %v\n", len(nodeID))
	}
}

func TestPubkeyID(t *testing.T) {
	bookkeeper := "1202027df359dff69eea8dd7d807b669dd9635292b1aae97d03ed32cb36ff30fb7e4d9"
	pubKey, err := hex.DecodeString(bookkeeper)
	if err != nil {
		t.Errorf("DecodeString failed: %v", err)
	}
	k, err := keypair.DeserializePublicKey(pubKey)
	if err != nil {
		t.Errorf("DeserializePublicKey failed: %v", err)
	}
	nodeID, err := PubkeyID(k)
	if err != nil {
		t.Errorf("PubkeyID failed: %v", err)
	}
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
	nodeID, err := PubkeyID(k)
	if err != nil {
		t.Errorf("PubkeyID failed: %v", err)
	}
	publickey, err := nodeID.Pubkey()
	if err != nil {
		t.Errorf("Pubkey failed: %v", err)
	}
	t.Logf("res: %v", publickey)
}
