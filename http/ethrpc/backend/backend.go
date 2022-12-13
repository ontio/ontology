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

package backend

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common/bitutil"
	"github.com/ethereum/go-ethereum/core/bloombits"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/http/base/actor"
)

const (
	// bloomServiceThreads is the number of goroutines used globally by an Ethereum
	// instance to service bloombits lookups for all running filters.
	BloomServiceThreads = 16

	// bloomFilterThreads is the number of goroutines used locally per filter to
	// multiplex requests onto the global servicing goroutines.
	BloomFilterThreads = 3

	// bloomRetrievalBatch is the maximum number of bloom bit retrievals to service
	// in a single batch.
	BloomRetrievalBatch = 16

	// bloomRetrievalWait is the maximum time to wait for enough bloom bit requests
	// to accumulate request an entire batch (avoiding hysteresis).
	BloomRetrievalWait = time.Duration(0)
)

type BloomBackend struct {
	bloomRequests     chan chan *bloombits.Retrieval
	closeBloomHandler chan struct{}
}

func NewBloomBackend() *BloomBackend {
	return &BloomBackend{
		bloomRequests:     make(chan chan *bloombits.Retrieval),
		closeBloomHandler: make(chan struct{}),
	}
}

// Close
func (b *BloomBackend) Close() {
	close(b.closeBloomHandler)
}

func (b *BloomBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	for i := 0; i < BloomFilterThreads; i++ {
		go session.Multiplex(BloomRetrievalBatch, BloomRetrievalWait, b.bloomRequests)
	}
}

func (b *BloomBackend) BloomStatus() (uint32, uint32) {
	return actor.BloomStatus()
}

// startBloomHandlers starts a batch of goroutines to accept bloom bit database
// retrievals from possibly a range of filters and serving the data to satisfy.
func (b *BloomBackend) StartBloomHandlers(sectionSize uint32, db *leveldbstore.LevelDBStore) error {
	for i := 0; i < BloomServiceThreads; i++ {
		go func() {
			for {
				select {
				case <-b.closeBloomHandler:
					return

				case request := <-b.bloomRequests:
					task := <-request
					task.Bitsets = make([][]byte, len(task.Sections))
					for i, section := range task.Sections {
						if compVector, err := ledgerstore.ReadBloomBits(db, task.Bit, uint32(section)); err == nil {
							if blob, err := bitutil.DecompressBytes(compVector, int(sectionSize/8)); err == nil {
								task.Bitsets[i] = blob
							} else {
								task.Error = err
							}
						} else {
							task.Error = err
						}
					}
					request <- task
				}
			}
		}()
	}
	return nil
}
