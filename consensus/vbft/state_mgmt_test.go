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

package vbft

import (
	"sort"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/stretchr/testify/assert"
)

func FuzzMaxCommitBlock(f *testing.F) {
	f.Fuzz(func(t *testing.T, peersBytes []byte) {
		n := len(peersBytes) / 4
		if n < 4 {
			t.Skip()
		}
		source := common.NewZeroCopySource(peersBytes)
		var peers []uint32
		for i := 0; i < n; i += 1 {
			v, _ := source.ReadUint32()
			peers = append(peers, v)
		}
		c := len(peers) / 4
		start := peers[0]
		cmt := getMaxCommit(peers, start, c)
		sort.Slice(peers, func(i, j int) bool {
			return peers[j] < peers[i]
		})
		var maxCommitted uint32
		if len(peers) >= c+1 && peers[c] > start {
			maxCommitted = peers[c]
		}

		assert.Equal(t, cmt, maxCommitted)
	})
}

func getMaxCommit(peers []uint32, startBlkNum uint32, C int) uint32 {
	var maxCommitted uint32

	for _, n := range peers {
		if n > startBlkNum && n > maxCommitted {
			peerCount := 0
			for _, k := range peers {
				if k >= n {
					peerCount++
				}
			}
			if peerCount > C {
				maxCommitted = n
			}
		}
	}

	return maxCommitted
}
