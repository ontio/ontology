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

// package kbucket implements a kademlia 'k-bucket' routing table.
package kbucket

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ontio/ontology/p2pserver/common"
)

var ErrPeerRejectedHighLatency = errors.New("peer rejected; latency too high")
var ErrPeerRejectedNoCapacity = errors.New("peer rejected; insufficient capacity")

// maxCplForRefresh is the maximum cpl we support for refresh.
// This limit exists because we can only generate 'maxCplForRefresh' bit prefixes for now.
const maxCplForRefresh uint = 15

// CplRefresh contains a CPL(common prefix length) with the host & the last time
// we refreshed that cpl/searched for an ID which has that cpl with the host.
type CplRefresh struct {
	Cpl           uint
	LastRefreshAt time.Time
}

// RouteTable defines the routing table.
type RouteTable struct {
	// ID of the local peer
	local common.PeerId

	tabLock sync.RWMutex

	// kBuckets define all the fingers to other nodes.
	Buckets    []*Bucket
	bucketsize int

	cplRefreshLk   sync.RWMutex
	cplRefreshedAt map[uint]time.Time

	// notification callback functions
	PeerRemoved func(id common.PeerId)
	PeerAdded   func(id common.PeerId)
}

func noop(_ common.PeerId) {}

// NewRoutingTable creates a new routing table with a given bucketsize, local ID, and latency tolerance.
func NewRoutingTable(bucketsize int, localID common.PeerId) *RouteTable {
	rt := &RouteTable{
		Buckets:        []*Bucket{newBucket()},
		bucketsize:     bucketsize,
		local:          localID,
		cplRefreshedAt: make(map[uint]time.Time),
		PeerRemoved:    noop,
		PeerAdded:      noop,
	}

	return rt
}

// GetTrackedCplsForRefresh returns the Cpl's we are tracking for refresh.
// Caller is free to modify the returned slice as it is a defensive copy.
func (rt *RouteTable) GetTrackedCplsForRefresh() []CplRefresh {
	rt.cplRefreshLk.RLock()
	defer rt.cplRefreshLk.RUnlock()

	cpls := make([]CplRefresh, 0, len(rt.cplRefreshedAt))

	for c, t := range rt.cplRefreshedAt {
		cpls = append(cpls, CplRefresh{c, t})
	}

	return cpls
}

// GenRandPeerID generates a random peerID for a given Cpl
func (rt *RouteTable) GenRandKadId(targetCpl uint) common.PeerId {
	if targetCpl > maxCplForRefresh {
		targetCpl = maxCplForRefresh
	}

	return rt.local.GenRandPeerId(targetCpl)
}

// ResetCplRefreshedAtForID resets the refresh time for the Cpl of the given ID.
func (rt *RouteTable) ResetCplRefreshedAtForID(id common.PeerId, newTime time.Time) {
	cpl := common.CommonPrefixLen(id, rt.local)
	if uint(cpl) > maxCplForRefresh {
		return
	}

	rt.cplRefreshLk.Lock()
	defer rt.cplRefreshLk.Unlock()

	rt.cplRefreshedAt[uint(cpl)] = newTime
}

// Update adds or moves the given peer to the front of its respective bucket
func (rt *RouteTable) Update(peerID common.PeerId, addr string) error {
	pair := common.PeerIDAddressPair{
		ID:      peerID,
		Address: addr,
	}
	cpl := common.CommonPrefixLen(peerID, rt.local)
	rt.tabLock.Lock()
	defer rt.tabLock.Unlock()
	bucketID := cpl
	if bucketID >= len(rt.Buckets) {
		bucketID = len(rt.Buckets) - 1
	}

	bucket := rt.Buckets[bucketID]
	if bucket.Has(peerID) {
		// If the peer is already in the table, move it to the front.
		// This signifies that it it "more active" and the less active nodes
		// Will as a result tend towards the back of the list

		// the paper put more active node at tail, we put more active node at head
		bucket.MoveToFront(peerID)
		return nil
	}

	// We have enough space in the bucket (whether spawned or grouped).
	if bucket.Len() < rt.bucketsize {
		bucket.PushFront(pair)
		// call back notifier
		rt.PeerAdded(peerID)
		return nil
	}

	if bucketID == len(rt.Buckets)-1 {
		// if the bucket is too large and this is the last bucket (i.e. wildcard), unfold it.
		rt.nextBucket()
		// the structure of the table has changed, so let's recheck if the peer now has a dedicated bucket.
		bucketID = cpl
		if bucketID >= len(rt.Buckets) {
			bucketID = len(rt.Buckets) - 1
		}
		bucket = rt.Buckets[bucketID]
		if bucket.Len() >= rt.bucketsize {
			// if after all the unfolding, we're unable to find room for this peer, scrap it.
			return ErrPeerRejectedNoCapacity
		}
		bucket.PushFront(pair)
		rt.PeerAdded(peerID)
		return nil
	}

	return ErrPeerRejectedNoCapacity
}

// Remove deletes a peer from the routing table. This is to be used
// when we are sure a node has disconnected completely.
func (rt *RouteTable) Remove(p common.PeerId) {
	peerID := p
	cpl := common.CommonPrefixLen(peerID, rt.local)

	rt.tabLock.Lock()
	defer rt.tabLock.Unlock()

	bucketID := cpl
	if bucketID >= len(rt.Buckets) {
		bucketID = len(rt.Buckets) - 1
	}

	bucket := rt.Buckets[bucketID]
	if bucket.Remove(p) {
		rt.PeerRemoved(p)
	}
}

func (rt *RouteTable) nextBucket() {
	// This is the last bucket, which allegedly is a mixed bag containing peers not belonging in dedicated (unfolded) buckets.
	// _allegedly_ is used here to denote that *all* peers in the last bucket might feasibly belong to another bucket.
	// This could happen if e.g. we've unfolded 4 buckets, and all peers in folded bucket 5 really belong in bucket 8.
	bucket := rt.Buckets[len(rt.Buckets)-1]
	newBucket := bucket.Split(len(rt.Buckets)-1, rt.local)
	rt.Buckets = append(rt.Buckets, newBucket)

	// The newly formed bucket still contains too many peers. We probably just unfolded a empty bucket.
	if newBucket.Len() >= rt.bucketsize {
		// Keep unfolding the table until the last bucket is not overflowing.
		rt.nextBucket()
	}
}

// Find a specific peer by ID or return nil
func (rt *RouteTable) Find(id common.PeerId) (common.PeerIDAddressPair, bool) {
	srch := rt.NearestPeers(id, 1)
	if len(srch) == 0 || srch[0].ID != id {
		return common.PeerIDAddressPair{}, false
	}

	return srch[0], true
}

func (rt *RouteTable) NearestPeers(id common.PeerId, count int) []common.PeerIDAddressPair {
	// This is the number of bits _we_ share with the key. All peers in this
	// bucket share cpl bits with us and will therefore share at least cpl+1
	// bits with the given key. +1 because both the target and all peers in
	// this bucket differ from us in the cpl bit.
	cpl := common.CommonPrefixLen(id, rt.local)

	// It's assumed that this also protects the buckets.
	rt.tabLock.RLock()

	// Get bucket index or last bucket
	if cpl >= len(rt.Buckets) {
		cpl = len(rt.Buckets) - 1
	}

	pds := peerDistanceSorter{
		peers:  make([]peerDistance, 0, count+rt.bucketsize),
		target: id,
	}

	// Add peers from the target bucket (cpl+1 shared bits).
	pds.appendPeersFromList(rt.Buckets[cpl].list)

	// If we're short, add peers from buckets to the right until we have
	// enough. All buckets to the right share exactly cpl bits (as opposed
	// to the cpl+1 bits shared by the peers in the cpl bucket).
	//
	// Unfortunately, to be completely correct, we can't just take from
	// buckets until we have enough peers because peers because _all_ of
	// these peers will be ~2**(256-cpl) from us.
	//
	// However, we're going to do that anyways as it's "good enough"

	for i := cpl + 1; i < len(rt.Buckets) && pds.Len() < count; i++ {
		pds.appendPeersFromList(rt.Buckets[i].list)
	}

	// If we're still short, add in buckets that share _fewer_ bits. We can
	// do this bucket by bucket because each bucket will share 1 fewer bit
	// than the last.
	//
	// * bucket cpl-1: cpl-1 shared bits.
	// * bucket cpl-2: cpl-2 shared bits.
	// ...
	for i := cpl - 1; i >= 0 && pds.Len() < count; i-- {
		pds.appendPeersFromList(rt.Buckets[i].list)
	}
	rt.tabLock.RUnlock()

	// Sort by distance to local peer
	pds.sort()

	if count < pds.Len() {
		pds.peers = pds.peers[:count]
	}

	out := make([]common.PeerIDAddressPair, 0, pds.Len())
	for _, p := range pds.peers {
		out = append(out, p.p)
	}

	return out
}

// Size returns the total number of peers in the routing table
func (rt *RouteTable) Size() int {
	var tot int
	rt.tabLock.RLock()
	for _, buck := range rt.Buckets {
		tot += buck.Len()
	}
	rt.tabLock.RUnlock()
	return tot
}

// ListPeers takes a RoutingTable and returns a list of all peers from all buckets in the table.
func (rt *RouteTable) ListPeers() []common.PeerIDAddressPair {
	var peers []common.PeerIDAddressPair
	rt.tabLock.RLock()
	for _, buck := range rt.Buckets {
		peers = append(peers, buck.Peers()...)
	}
	rt.tabLock.RUnlock()
	return peers
}

// Print prints a descriptive statement about the provided RouteTable
func (rt *RouteTable) Print() {
	fmt.Printf("Routing Table, bs = %d\n", rt.bucketsize)
	rt.tabLock.RLock()

	for i, b := range rt.Buckets {
		fmt.Printf("\tbucket: %d\n", i)

		b.lk.RLock()
		for e := b.list.Front(); e != nil; e = e.Next() {
			p := e.Value.(common.PeerIDAddressPair)
			fmt.Printf("\t\t- %s\n", p.ID.ToHexString())
		}
		b.lk.RUnlock()
	}
	rt.tabLock.RUnlock()
}
