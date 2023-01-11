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

package filters

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/bloombits"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/filters"
	common2 "github.com/ontio/ontology/common"
	common4 "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/core/store/ledgerstore"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/http/base/actor"
	utils2 "github.com/ontio/ontology/http/ethrpc/utils"
	"github.com/ontio/ontology/smartcontract/event"
)

const MAX_SEARCH_RANGE = 100000

type Backend interface {
	BloomStatus() (uint32, uint32)
	ServiceFilter(ctx context.Context, session *bloombits.MatcherSession)
}

// Filter can be used to retrieve and filter logs.
type Filter struct {
	backend  Backend
	criteria filters.FilterCriteria
	matcher  *bloombits.Matcher
}

// NewBlockFilter creates a new filter which directly inspects the contents of
// a block to figure out whether it is interesting or not.
func NewBlockFilter(backend Backend, criteria filters.FilterCriteria) *Filter {
	// Create a generic filter and convert it into a block filter
	return newFilter(backend, criteria, nil)
}

// NewRangeFilter creates a new filter which uses a bloom filter on blocks to
// figure out whether a particular block is interesting or not.
func NewRangeFilter(backend Backend, begin, end int64, addresses []common.Address, topics [][]common.Hash) *Filter {
	// Flatten the address and topic filter clauses into a single bloombits filter
	// system. Since the bloombits are not positional, nil topics are permitted,
	// which get flattened into a nil byte slice.
	var filtersBz [][][]byte // nolint: prealloc
	if len(addresses) > 0 {
		filter := make([][]byte, len(addresses))
		for i, address := range addresses {
			filter[i] = address.Bytes()
		}
		filtersBz = append(filtersBz, filter)
	}

	for _, topicList := range topics {
		filter := make([][]byte, len(topicList))
		for i, topic := range topicList {
			filter[i] = topic.Bytes()
		}
		filtersBz = append(filtersBz, filter)
	}

	size, _ := actor.BloomStatus()

	// Create a generic filter and convert it into a range filter
	criteria := filters.FilterCriteria{
		FromBlock: big.NewInt(begin),
		ToBlock:   big.NewInt(end),
		Addresses: addresses,
		Topics:    topics,
	}

	return newFilter(backend, criteria, bloombits.NewMatcher(uint64(size), filtersBz))
}

// newFilter returns a new Filter
func newFilter(backend Backend, criteria filters.FilterCriteria, matcher *bloombits.Matcher) *Filter {
	return &Filter{
		backend:  backend,
		criteria: criteria,
		matcher:  matcher,
	}
}

// Logs searches the blockchain for matching log entries, returning all from the
// first block that contains matches, updating the start of the filter accordingly.
func (f *Filter) Logs(ctx context.Context) ([]*ethtypes.Log, error) {
	var logs []*ethtypes.Log
	var err error

	// If we're doing singleton block filtering, execute and return
	if f.criteria.BlockHash != nil && f.criteria.BlockHash != (&common.Hash{}) {
		block, err := actor.GetBlockFromStore(common2.Uint256(*f.criteria.BlockHash))
		if err != nil {
			return nil, err
		}
		if block.Header == nil {
			return nil, fmt.Errorf("unknown block header %s", f.criteria.BlockHash.String())
		}
		bloom, err := actor.GetBloomData(block.Header.Height)
		if err != nil {
			return nil, err
		}
		return f.blockLogs(bloom, common2.Uint256(*f.criteria.BlockHash))
	}

	// Figure out the limits of the filter range
	curHeight := actor.GetCurrentBlockHeight()
	block, err := actor.GetBlockByHeight(curHeight)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, nil
	}

	if f.criteria.FromBlock.Int64() == -1 {
		f.criteria.FromBlock = big.NewInt(int64(curHeight))
	}
	if f.criteria.ToBlock.Int64() == -1 {
		f.criteria.ToBlock = big.NewInt(int64(curHeight))
	}

	start := actor.GetFilterStart()
	minFilterStart := ledgerstore.MinFilterStart()
	if f.criteria.ToBlock.Int64() < int64(minFilterStart) {
		return nil, nil
	}

	if f.criteria.FromBlock.Int64() < int64(minFilterStart) {
		f.criteria.FromBlock = big.NewInt(int64(minFilterStart))
	}

	if f.criteria.FromBlock.Int64() < int64(start) ||
		f.criteria.ToBlock.Int64() < int64(start) {
		return nil, fmt.Errorf("from and to block height must greater than %d", int64(start))
	}

	if f.criteria.ToBlock.Int64()-f.criteria.FromBlock.Int64() > MAX_SEARCH_RANGE {
		return nil, fmt.Errorf("the span between fromBlock and toBlock must be less than or equal to %d", MAX_SEARCH_RANGE)
	}

	begin := f.criteria.FromBlock.Uint64()
	end := f.criteria.ToBlock.Uint64()
	size, sections := actor.BloomStatus()

	if indexed := uint64(sections * size); indexed > begin {
		if indexed > end {
			logs, err = f.indexedLogs(ctx, end)
		} else {
			logs, err = f.indexedLogs(ctx, indexed-1)
		}
		if err != nil {
			return logs, err
		}
	}
	rest, err := f.unindexedLogs(ctx, end)
	logs = append(logs, rest...)
	return logs, err
}

// blockLogs returns the logs matching the filter criteria within a single block.
func (f *Filter) blockLogs(bloom ethtypes.Bloom, hash common2.Uint256) ([]*ethtypes.Log, error) {
	if !bloomFilter(bloom, f.criteria.Addresses, f.criteria.Topics) {
		return []*ethtypes.Log{}, nil
	}

	logsList, err := getLogs(hash)
	if err != nil {
		return []*ethtypes.Log{}, err
	}

	var unfiltered []*ethtypes.Log // nolint: prealloc
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs...)
	}
	logs := FilterLogs(unfiltered, nil, nil, f.criteria.Addresses, f.criteria.Topics)
	if len(logs) == 0 {
		return []*ethtypes.Log{}, nil
	}
	return logs, nil
}

func getLogs(hash common2.Uint256) ([][]*ethtypes.Log, error) {
	block, err := actor.GetBlockFromStore(hash)
	if err != nil {
		if err == common4.ErrNotFound {
			return nil, err
		}
		return nil, err
	}

	var res [][]*ethtypes.Log
	for _, tx := range block.Transactions {
		if tx.TxType != types.EIP155 {
			continue
		}
		notify, err := actor.GetEventNotifyByTxHash(tx.Hash())
		if err != nil {
			if err == common4.ErrNotFound {
				continue
			}
			return nil, err
		}
		if notify != nil {
			txLogs, err := generateLog(notify)
			if err != nil {
				return nil, err
			}
			if txLogs != nil && len(txLogs) != 0 {
				res = append(res, txLogs)
			}
		}
	}
	return res, nil
}

func generateLog(rawNotify *event.ExecuteNotify) ([]*ethtypes.Log, error) {
	var res []*ethtypes.Log
	txHash := rawNotify.TxHash
	height, _, err := actor.GetTxnWithHeightByTxHash(txHash)
	if err != nil {
		return nil, err
	}
	hash := actor.GetBlockHashFromStore(height)
	ethHash := utils2.OntToEthHash(hash)
	for idx, n := range rawNotify.Notify {
		storageLog, err := event.NotifyEventInfoToEvmLog(n)
		if err != nil {
			return nil, err
		}
		res = append(res,
			&ethtypes.Log{
				Address:     storageLog.Address,
				Topics:      storageLog.Topics,
				Data:        storageLog.Data,
				BlockNumber: uint64(height),
				TxHash:      utils2.OntToEthHash(txHash),
				TxIndex:     uint(rawNotify.TxIndex),
				BlockHash:   ethHash,
				Index:       uint(idx),
				Removed:     false,
			})
	}

	return res, nil
}

// checkMatches checks if the receipts belonging to the given header contain any log events that
// match the filter criteria. This function is called when the bloom filter signals a potential match.
func (f *Filter) checkMatches(hash common2.Uint256) (logs []*ethtypes.Log, err error) {
	// Get the logs of the block
	logsList, err := getLogs(hash)
	if err != nil {
		return nil, err
	}
	var unfiltered []*ethtypes.Log
	for _, logs := range logsList {
		unfiltered = append(unfiltered, logs...)
	}
	logs = filterLogs(unfiltered, nil, nil, f.criteria.Addresses, f.criteria.Topics)
	return logs, nil
}

// indexedLogs returns the logs matching the filter criteria based on the bloom
// bits indexed available locally or via the network.
func (f *Filter) indexedLogs(ctx context.Context, end uint64) ([]*ethtypes.Log, error) {
	// Create a matcher session and request servicing from the backend
	matches := make(chan uint64, 64)

	session, err := f.matcher.Start(ctx, f.criteria.FromBlock.Uint64(), end, matches)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	f.backend.ServiceFilter(ctx, session)

	// Iterate over the matches until exhausted or context closed
	var logs []*ethtypes.Log

	bigEnd := big.NewInt(int64(end))
	for {
		select {
		case number, ok := <-matches:

			// Abort if all matches have been fulfilled
			if !ok {
				err := session.Error()
				if err == nil {
					f.criteria.FromBlock = bigEnd.Add(bigEnd, big.NewInt(1))
				}
				return logs, err
			}
			f.criteria.FromBlock = big.NewInt(int64(number)).Add(big.NewInt(int64(number)), big.NewInt(1))

			// Retrieve the suggested block and pull any truly matching logs
			block, err := actor.GetBlockByHeight(uint32(number))
			if err != nil {
				return nil, err
			}
			if block == nil {
				return nil, fmt.Errorf("block %v not found", number)
			}
			found, err := f.checkMatches(block.Hash())
			if err != nil {
				return logs, err
			}
			logs = append(logs, found...)

		case <-ctx.Done():
			return logs, ctx.Err()
		}
	}
}

// unindexedLogs returns the logs matching the filter criteria based on raw block
// iteration and bloom matching.
func (f *Filter) unindexedLogs(ctx context.Context, end uint64) ([]*ethtypes.Log, error) {
	var logs []*ethtypes.Log
	begin := f.criteria.FromBlock.Int64()
	beginPtr := &begin
	defer f.criteria.FromBlock.SetInt64(*beginPtr)

	for ; begin <= int64(end); begin++ {
		block, err := actor.GetBlockByHeight(uint32(begin))
		if err != nil {
			return nil, err
		}
		if block == nil {
			return logs, nil
		}
		if block.Header == nil {
			return nil, fmt.Errorf("unknown block header %s", f.criteria.BlockHash.String())
		}
		bloom, err := actor.GetBloomData(block.Header.Height)
		if err != nil {
			return nil, err
		}
		found, err := f.blockLogs(bloom, block.Hash())
		if err != nil {
			return logs, err
		}
		logs = append(logs, found...)
	}
	return logs, nil
}

// filterLogs creates a slice of logs matching the given criteria.
func filterLogs(logs []*ethtypes.Log, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) []*ethtypes.Log {
	var ret []*ethtypes.Log
Logs:
	for _, log := range logs {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
			continue
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
			continue
		}

		if len(addresses) > 0 && !includes(addresses, log.Address) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue Logs
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}

// filterLogs creates a slice of logs matching the given criteria.
// [] -> anything
// [A] -> A in first position of log topics, anything after
// [null, B] -> anything in first position, B in second position
// [A, B] -> A in first position and B in second position
// [[A, B], [A, B]] -> A or B in first position, A or B in second position
func FilterLogs(logs []*ethtypes.Log, fromBlock, toBlock *big.Int, addresses []common.Address, topics [][]common.Hash) []*ethtypes.Log {
	var ret []*ethtypes.Log
Logs:
	for _, log := range logs {
		if fromBlock != nil && fromBlock.Int64() >= 0 && fromBlock.Uint64() > log.BlockNumber {
			continue
		}
		if toBlock != nil && toBlock.Int64() >= 0 && toBlock.Uint64() < log.BlockNumber {
			continue
		}
		if len(addresses) > 0 && !includes(addresses, log.Address) {
			continue
		}
		// If the to filtered topics is greater than the amount of topics in logs, skip.
		if len(topics) > len(log.Topics) {
			continue
		}
		for i, sub := range topics {
			match := len(sub) == 0 // empty rule set == wildcard
			for _, topic := range sub {
				if log.Topics[i] == topic {
					match = true
					break
				}
			}
			if !match {
				continue Logs
			}
		}
		ret = append(ret, log)
	}
	return ret
}

func includes(addresses []common.Address, a common.Address) bool {
	for _, addr := range addresses {
		if addr == a {
			return true
		}
	}

	return false
}

func bloomFilter(bloom ethtypes.Bloom, addresses []common.Address, topics [][]common.Hash) bool {
	var included bool = true
	if len(addresses) > 0 {
		included = false
		for _, addr := range addresses {
			if ethtypes.BloomLookup(bloom, addr) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, sub := range topics {
		included = len(sub) == 0 // empty rule set == wildcard
		for _, topic := range sub {
			if ethtypes.BloomLookup(bloom, topic) {
				included = true
				break
			}
		}
	}
	return included
}

// returnHashes is a helper that will return an empty hash array case the given hash array is nil,
// otherwise the given hashes array is returned.
func returnHashes(hashes []common.Hash) []common.Hash {
	if hashes == nil {
		return []common.Hash{}
	}
	return hashes
}

// returnLogs is a helper that will return an empty log array in case the given logs array is nil,
// otherwise the given logs array is returned.
func returnLogs(logs []*ethtypes.Log) []*ethtypes.Log {
	if logs == nil {
		return []*ethtypes.Log{}
	}
	return logs
}
