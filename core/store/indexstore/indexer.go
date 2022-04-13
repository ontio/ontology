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
package indexstore

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/store/leveldbstore"
)

var (
	indexer *Indexer
	once    sync.Once
)

func CloseIndexer() {
	if indexer != nil && indexer.backend.store != nil {
		indexer.backend.store.Close()
	}
}

// Indexer does a post-processing job for equally sized sections of the
// canonical chain (like BlooomBits and CHT structures). A Indexer is
// connected to the blockchain through the event system by starting a
// ChainHeadEventLoop in a goroutine.
//
// Further child ChainIndexers can be added which use the output of the parent
// section indexer. These child indexers receive new head notifications only
// after an entire section has been finished or in case of rollbacks that might
// affect already finished sections.
type Indexer struct {
	backend bloomIndexer // Background processor generating the index data content

	update chan struct{} // Notification channel that headers should be processed
	quit   chan struct{} // Quit channel to tear down running goroutines

	storedSections uint32 // Number of sections successfully indexed into the database
	processing     uint32 // Atomic flag whether indexer is processing or not
}

func InitIndexer(dataDir string) error {
	store, err := leveldbstore.NewLevelDBStore(filepath.Join(dataDir, bloomIdxDir))
	if err != nil {
		return err
	}
	indexer = &Indexer{
		backend: initBloomIndexer(store),
		update:  make(chan struct{}, 1), // updata ch
		quit:    make(chan struct{}),
	}
	indexer.setValidSections(indexer.GetValidSections())
	return err
}

func GetIndexer() *Indexer {
	return indexer
}

func (i *Indexer) StoredSection() uint32 {
	if i != nil {
		return i.storedSections
	}
	return 0
}

func (i *Indexer) IsProcessing() bool {
	return atomic.LoadUint32(&i.processing) == 1
}

func (i *Indexer) ProcessSection(k store.LedgerStore, interval uint32) error {
	if atomic.SwapUint32(&i.processing, 1) == 1 {
		return fmt.Errorf("matcher is already running")
	}
	defer atomic.StoreUint32(&i.processing, 0)
	knownSection := interval / BloomBitsBlocks
	for i.storedSections < knownSection {
		section := i.storedSections
		var lastHead common.Hash
		if section > 0 {
			lastHead = i.sectionHead(section - 1)
		}

		var err error

		// Reset and partial processing
		if err = i.backend.Reset(section); err != nil {
			i.setValidSections(0)
			return fmt.Errorf(err.Error())
		}

		begin := section*BloomBitsBlocks + config.GetAddFilterHeight()
		end := (section+1)*BloomBitsBlocks + config.GetAddFilterHeight()

		for number := begin; number < end; number++ {
			var (
				bloom ethtypes.Bloom
				hash  common.Hash
			)
			i.updateBlock()
			if number == config.GetAddFilterHeight() {
				bloom = ethtypes.Bloom{}
				hash = common.Hash{}
			} else {
				blockHash := k.GetBlockHash(number)
				hash = common.BytesToHash(blockHash.ToArray())
				if hash == (common.Hash{}) {
					return fmt.Errorf("canonical block %d unknown", number)
				}

				bloom, err = k.GetBloomData(number)
				if err != nil {
					return fmt.Errorf("get bloom data height: %d %s", number, err.Error())
				}
			}
			i.backend.Process(hash, number, bloom)
			lastHead = hash
		}

		bd, err := i.backend.Commit()
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		i.setSectionHead(section, lastHead)
		i.setValidSections(section + 1)
		i.setBloomData(&bd, section, lastHead)
	}
	return nil
}

// GetDB get db of bloomIndexer
func (b *Indexer) GetDB() *leveldbstore.LevelDBStore {
	if b != nil {
		return b.backend.store
	}
	return nil
}

// setValidSections writes the number of valid sections to the index database
func (i *Indexer) setValidSections(sections uint32) {
	// Set the current number of valid sections in the database
	var data [8]byte
	binary.BigEndian.PutUint32(data[:], sections)
	i.backend.store.Put([]byte("count"), data[:])

	// Remove any reorged sections, caching the valids in the mean time
	for i.storedSections > sections {
		i.storedSections--
		i.removeSectionHead(i.storedSections)
	}
	i.storedSections = sections // needed if new > old
}

// setBloomData put SectionHead and ValidSections into watcher.bloomData
func (i *Indexer) setBloomData(bloomData *[]*KV, section uint32, hash common.Hash) {
	var data [8]byte
	binary.BigEndian.PutUint32(data[:], section)
	*bloomData = append(*bloomData, &KV{Key: append([]byte("shead"), data[:]...), Value: hash.Bytes()})
	*bloomData = append(*bloomData, &KV{Key: []byte("count"), Value: data[:]})
}

// GetValidSections reads the number of valid sections from the index database
// and caches is into the local state.
func (i *Indexer) GetValidSections() uint32 {
	data, _ := i.backend.store.Get([]byte("count"))
	if len(data) == 8 {
		return binary.BigEndian.Uint32(data)
	}
	return 0
}

// sectionHead retrieves the last block hash of a processed section from the
// index database.
func (i *Indexer) sectionHead(section uint32) common.Hash {
	var data [8]byte
	binary.BigEndian.PutUint32(data[:], section)

	hash, _ := i.backend.store.Get(append([]byte("shead"), data[:]...))
	if len(hash) == len(common.Hash{}) {
		return common.BytesToHash(hash)
	}
	return common.Hash{}
}

// setSectionHead writes the last block hash of a processed section to the index
// database.
func (i *Indexer) setSectionHead(section uint32, hash common.Hash) {
	var data [8]byte
	binary.BigEndian.PutUint32(data[:], section)

	i.backend.store.Put(append([]byte("shead"), data[:]...), hash.Bytes())
}

// removeSectionHead removes the reference to a processed section from the index
// database.
func (i *Indexer) removeSectionHead(section uint32) {
	var data [8]byte
	binary.BigEndian.PutUint32(data[:], section)

	i.backend.store.Delete(append([]byte("shead"), data[:]...))
}

func (i *Indexer) NotifyNewHeight() {
	i.update <- struct{}{}
}

func (i *Indexer) updateBlock() {
	exit := false
	for {
		select {
		case <-i.update:
		default:
			exit = true
		}
		if exit {
			break
		}
	}
}
