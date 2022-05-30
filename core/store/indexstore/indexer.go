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
	"fmt"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ethereum/go-ethereum/core/types"
	common2 "github.com/ontio/ontology/common"
	common3 "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/leveldbstore"
)

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
	backend        bloomIndexer // Background processor generating the index data content
	start          uint32       // Start is the start height of bloom filter supported
	storedSections uint32       // Number of sections successfully indexed into the database
}

func New(dataDir string, curBlockHeight uint32) (*Indexer, error) {
	db, err := leveldbstore.NewLevelDBStore(filepath.Join(dataDir, bloomIdxDir))
	if err != nil {
		return nil, err
	}

	backend := initBloomIndexer(db)

	start, err := backend.GetFilterStart()
	if err != nil {
		if err != common3.ErrNotFound {
			return nil, fmt.Errorf("new indexer: %s", err.Error())
		}
		err = backend.PutFilterStart(curBlockHeight)
		if err != nil {
			return nil, fmt.Errorf("new indexer: %s", err.Error())
		}
		start = curBlockHeight
	}

	indexer := &Indexer{
		start:   start,
		backend: backend,
	}

	if err := indexer.setValidSections(indexer.getValidSections()); err != nil {
		return nil, err
	}
	return indexer, nil
}

func (i *Indexer) StoredSection() uint32 {
	if i != nil {
		return i.storedSections
	}
	return 0
}

type LedgerStore interface {
	GetBlockHash(height uint32) common2.Uint256
	GetBloomData(height uint32) (types2.Bloom, error)
	ClearBloomCache(begins, end uint32)
}

func (i *Indexer) ProcessSection(k LedgerStore, blockHeight uint32) error {

	knownSection := (blockHeight - i.start) / BloomBitsBlocks
	for i.storedSections < knownSection {
		section := i.storedSections

		var err error
		// Reset and partial processing
		i.backend.Reset(section)

		begin := section*BloomBitsBlocks + i.start
		end := (section+1)*BloomBitsBlocks + i.start

		for number := begin; number < end; number++ {
			blockHash := k.GetBlockHash(number)
			hash := common.Hash(blockHash)
			if hash == (common.Hash{}) {
				return fmt.Errorf("canonical block %d unknown", number)
			}

			bloom, err := k.GetBloomData(number)
			if err != nil {
				return fmt.Errorf("get bloom data height: %d %s", number, err.Error())
			}
			i.backend.Process(hash, number, bloom, i.start)
		}

		err = i.backend.Commit()
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		err = i.setValidSections(section + 1)
		if err != nil {
			return fmt.Errorf(err.Error())
		}

		k.ClearBloomCache(begin, end)
	}
	return nil
}

func (i *Indexer) Close() error {
	if i != nil && i.backend.store != nil {
		return i.backend.store.Close()
	}
	return nil
}

// GetDB get db of bloomIndexer
func (i *Indexer) GetDB() *leveldbstore.LevelDBStore {
	if i != nil {
		return i.backend.store
	}
	return nil
}

func (i *Indexer) GetFilterStart() uint32 {
	return i.start
}

// setValidSections writes the number of valid sections to the index database
func (i *Indexer) setValidSections(sections uint32) error {
	// Set the current number of valid sections in the database
	err := i.backend.SetValidSections(sections)
	if err != nil {
		return err
	}
	i.storedSections = sections // needed if new > old
	return nil
}

// GetValidSections reads the number of valid sections from the index database
// and caches is into the local state.
func (i *Indexer) getValidSections() uint32 {
	return i.backend.GetValidSections()
}
