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

package dht

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/ontio/ontology/p2pserver/common"
	"github.com/stretchr/testify/assert"
)

func TestNewDHT(t *testing.T) {
	id := common.RandPeerKeyId()
	dht := NewDHT(id.Id)
	assert.NotNil(t, dht)
	assert.Equal(t, dht.AutoRefresh, true)
}

func init() {
	rand.Seed(time.Now().Unix())
	common.Difficulty = 1
}
func TestDHT_Update(t *testing.T) {
	for i := 0; i < 10; i++ {
		id := common.RandPeerKeyId()
		dht := NewDHT(id.Id)
		local := dht.localId
		prefix := rand.Int31n(15)
		kid := local.GenRandPeerId(uint(prefix))
		boo := dht.Update(kid, "127.0.0.1")
		assert.True(t, boo)
		if prefix == 0 {
			continue
		}
		kids := dht.BetterPeers(dht.localId, int(prefix))
		assert.Equal(t, len(kids), 1)
		assert.Equal(t, kids[0].ID, kid)
	}
}

func TestDHT_Remove(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := common.RandPeerKeyId()
		dht := NewDHT(id.Id)
		local := dht.localId
		prefix := rand.Int31n(15)
		kid := local.GenRandPeerId(uint(prefix))
		boo := dht.Update(kid, "127.0.0.1")
		assert.True(t, boo)
		kids := dht.BetterPeers(dht.localId, 1)
		assert.Equal(t, len(kids), 1)
		assert.Equal(t, kids[0].ID, kid)
		dht.Remove(kid)
		kids = dht.BetterPeers(dht.localId, int(prefix))
		assert.Equal(t, len(kids), 0)
	}

}

func TestDHT_BetterPeers(t *testing.T) {
	id := common.RandPeerKeyId()
	dht := NewDHT(id.Id)
	local := dht.localId
	rand.Seed(time.Now().Unix())
	prefix := rand.Int31n(15)
	for i := 0; i < 15; i++ {
		kid := local.GenRandPeerId(uint(prefix))
		boo := dht.Update(kid, "127.0.0.1")
		if !boo {
			fmt.Println(boo, prefix)
		}
		assert.True(t, boo)
	}
	kids := dht.BetterPeers(dht.localId, 3)
	assert.Equal(t, len(kids), 3)
}
