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

package kbucket

import (
	"bytes"
	"container/list"
	"sort"

	"github.com/ontio/ontology/p2pserver/common"
)

// A helper struct to sort peers by their distance to the local node
type peerDistance struct {
	p        common.PeerIDAddressPair
	distance [20]byte
}

// peerDistanceSorter implements sort.Interface to sort peers by xor distance
type peerDistanceSorter struct {
	peers  []peerDistance
	target common.PeerId
}

func (pds *peerDistanceSorter) Len() int { return len(pds.peers) }
func (pds *peerDistanceSorter) Swap(a, b int) {
	pds.peers[a], pds.peers[b] = pds.peers[b], pds.peers[a]
}
func (pds *peerDistanceSorter) Less(a, b int) bool {
	return bytes.Compare(pds.peers[a].distance[:], pds.peers[b].distance[:]) < 0
}

// Append the peer.ID to the sorter's slice. It may no longer be sorted.
func (pds *peerDistanceSorter) appendPeer(p common.PeerIDAddressPair) {
	pds.peers = append(pds.peers, peerDistance{
		p:        p,
		distance: pds.target.Distance(p.ID),
	})
}

// Append the peer.ID values in the list to the sorter's slice. It may no longer be sorted.
func (pds *peerDistanceSorter) appendPeersFromList(l *list.List) {
	for e := l.Front(); e != nil; e = e.Next() {
		pds.appendPeer(e.Value.(common.PeerIDAddressPair))
	}
}

func (pds *peerDistanceSorter) sort() {
	sort.Sort(pds)
}
