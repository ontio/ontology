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

package common

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertPeerID(t *testing.T) {
	start := time.Now().Unix()
	fmt.Println("start:", start)
	RandPeerKeyId()

	end := time.Now().Unix()
	fmt.Println("end:", end)
	fmt.Println(end - start)
}

func TestKIdToUint64(t *testing.T) {
	for i := 0; i < 100; i++ {
		data := rand.Uint64()
		id := PseudoPeerIdFromUint64(data)
		data2 := id.ToUint64()
		assert.Equal(t, data, data2)
	}
}

func TestKadId_IsEmpty(t *testing.T) {
	id := PeerId{}
	assert.True(t, id.IsEmpty())
	kid := RandPeerKeyId()
	assert.False(t, kid.Id.IsEmpty())
}
