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
package types

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// BlockNumber represents decoding hex string to block values
type BlockNumber int64

const (
	// LatestBlockNumber mapping from "latest" to 0 for tm query
	LatestBlockNumber = BlockNumber(0)

	// EarliestBlockNumber mapping from "earliest" to 1 for tm query (earliest query not supported)
	EarliestBlockNumber = BlockNumber(1)

	// PendingBlockNumber mapping from "pending" to -1 for tm query
	PendingBlockNumber = BlockNumber(-1)
)

// NewBlockNumber creates a new BlockNumber instance.
func NewBlockNumber(n *big.Int) BlockNumber {
	return BlockNumber(n.Int64())
}

// UnmarshalJSON parses the given JSON fragment into a BlockNumber. It supports:
// - "latest", "earliest" or "pending" as string arguments
// - the block number
// Returned errors:
// - an invalid block number error when the given argument isn't a known strings
// - an out of range error when the given block number is either too little or too large
func (bn *BlockNumber) UnmarshalJSON(data []byte) error {
	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}

	switch input {
	case "earliest":
		*bn = EarliestBlockNumber
		return nil
	case "latest":
		*bn = LatestBlockNumber
		return nil
	case "pending":
		*bn = PendingBlockNumber
		return nil
	}

	blckNum, err := hexutil.DecodeUint64(input)
	if err != nil {
		return err
	}
	if blckNum > math.MaxInt64 {
		return fmt.Errorf("blocknumber too high")
	}

	*bn = BlockNumber(blckNum)
	return nil
}

func (bn BlockNumber) IsLatest() bool {
	if bn == 0 {
		return true
	}
	return false
}

func (bn BlockNumber) IsEarliest() bool {
	if bn == 1 {
		return true
	}
	return false
}

func (bn BlockNumber) IsPending() bool {
	if bn == -1 {
		return true
	}
	return false
}

// Int64 converts block number to primitive type
func (bn BlockNumber) Int64() int64 {
	return int64(bn)
}
