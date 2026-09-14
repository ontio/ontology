// Copyright 2024 The ontology Authors
// This file is part of the ontology library.
//
// The ontology is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The ontology is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the ontology library. If not, see <http://www.gnu.org/licenses/>.

package tracers

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestCheckTraceCallParamsRejectsUnsupportedBlock(t *testing.T) {
	hash := common.HexToHash("0x1234")
	if err := checkTraceCallParams(rpc.BlockNumberOrHashWithHash(hash, false), nil); err == nil {
		t.Error("expected block hash to be rejected")
	}
	for _, blockNumber := range []rpc.BlockNumber{
		rpc.EarliestBlockNumber,
		rpc.SafeBlockNumber,
		rpc.FinalizedBlockNumber,
		rpc.BlockNumber(7),
	} {
		if err := checkTraceCallParams(rpc.BlockNumberOrHashWithNumber(blockNumber), nil); err == nil {
			t.Errorf("expected block %v to be rejected", blockNumber)
		}
	}
	for _, blockNumber := range []rpc.BlockNumber{rpc.LatestBlockNumber, rpc.PendingBlockNumber} {
		if err := checkTraceCallParams(rpc.BlockNumberOrHashWithNumber(blockNumber), nil); err != nil {
			t.Errorf("block %v should be accepted, got %v", blockNumber, err)
		}
	}
	if err := checkTraceCallParams(rpc.BlockNumberOrHash{}, nil); err != nil {
		t.Errorf("unspecified block should be accepted, got %v", err)
	}
}

func TestCheckTraceCallParamsRejectsUnsupportedOverrides(t *testing.T) {
	latest := rpc.BlockNumberOrHashWithNumber(rpc.LatestBlockNumber)

	txIndex := hexutil.Uint(1)
	if err := checkTraceCallParams(latest, &TraceCallConfig{TxIndex: &txIndex}); err == nil {
		t.Error("expected txIndex to be rejected")
	}

	number := hexutil.Big(*big.NewInt(5))
	blockOverrides := &BlockOverrides{Number: &number}
	if err := checkTraceCallParams(latest, &TraceCallConfig{BlockOverrides: blockOverrides}); err == nil {
		t.Error("expected block overrides to be rejected")
	}
	// An empty override object has no effect, so it is not worth rejecting.
	if err := checkTraceCallParams(latest, &TraceCallConfig{BlockOverrides: &BlockOverrides{}}); err != nil {
		t.Errorf("empty block overrides should be accepted, got %v", err)
	}

	// State overrides are applied by TraceCall and remain supported.
	stateOverrides := StateOverride{}
	if err := checkTraceCallParams(latest, &TraceCallConfig{StateOverrides: &stateOverrides}); err != nil {
		t.Errorf("state overrides should be accepted, got %v", err)
	}
}
