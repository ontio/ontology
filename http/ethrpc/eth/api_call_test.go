/*
 * Copyright (C) 2024 The ontology Authors
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
package eth

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	types2 "github.com/ontio/ontology/http/ethrpc/types"
)

func TestCheckCallParamsRejectsUnsupportedBlock(t *testing.T) {
	for _, blockNumber := range []types2.BlockNumber{
		types2.EarliestBlockNumber,
		types2.BlockNumber(1),
		types2.BlockNumber(12345),
	} {
		if err := checkCallParams(blockNumber, nil); err == nil {
			t.Errorf("expected block %v to be rejected", blockNumber)
		}
	}
}

func TestCheckCallParamsRejectsStateOverrides(t *testing.T) {
	overrides := map[common.Address]types2.Account{common.Address{}: {}}
	if err := checkCallParams(types2.LatestBlockNumber, &overrides); err == nil {
		t.Error("expected state overrides to be rejected")
	}
}

func TestCheckCallParamsAcceptsLatestAndPending(t *testing.T) {
	for _, blockNumber := range []types2.BlockNumber{types2.LatestBlockNumber, types2.PendingBlockNumber} {
		if err := checkCallParams(blockNumber, nil); err != nil {
			t.Errorf("block %v without overrides should be accepted, got %v", blockNumber, err)
		}
	}
	// An empty override set has no effect, so it is not worth rejecting.
	empty := map[common.Address]types2.Account{}
	if err := checkCallParams(types2.LatestBlockNumber, &empty); err != nil {
		t.Errorf("empty state overrides should be accepted, got %v", err)
	}
}
