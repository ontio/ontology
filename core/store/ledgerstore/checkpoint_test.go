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

package ledgerstore

import (
	"strings"
	"testing"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
)

func checkpointTestHash(b byte) common.Uint256 {
	var h common.Uint256
	h[0] = b
	return h
}

func checkpointTestBlock(height uint32, blockRoot common.Uint256) *types.Block {
	return &types.Block{
		Header: &types.Header{
			PrevBlockHash:    checkpointTestHash(0x01),
			TransactionsRoot: checkpointTestHash(0x02),
			BlockRoot:        blockRoot,
			Timestamp:        100,
			Height:           height,
			ConsensusData:    1,
		},
	}
}

// TestSyncCheckpointListsSorted is the CI guard for the checkpoint lists:
// adding an out-of-order or duplicate height fails the build.
func TestSyncCheckpointListsSorted(t *testing.T) {
	lists := []struct {
		name        string
		checkpoints []SyncCheckpoint
	}{
		{"mainnet", mainNetSyncCheckpoints},
		{"polaris", polarisNetSyncCheckpoints},
	}

	for _, network := range lists {
		t.Run(network.name, func(t *testing.T) {
			for i := 1; i < len(network.checkpoints); i++ {
				prev, cur := network.checkpoints[i-1], network.checkpoints[i]
				if prev.Height >= cur.Height {
					t.Fatalf("checkpoints must be strictly increasing by height: index %d has height %d after %d",
						i, cur.Height, prev.Height)
				}
			}
		})
	}
}

func TestMustHexHashRoundTrip(t *testing.T) {
	h := checkpointTestHash(0x7f)
	if got := mustHexHash(h.ToHexString()); got != h {
		t.Fatalf("hex round trip mismatch: expected %s, got %s", h.ToHexString(), got.ToHexString())
	}
}

func TestCheckSyncCheckpoint(t *testing.T) {
	const height = uint32(100)

	blockRoot := checkpointTestHash(0x03)
	merkleRoot := checkpointTestHash(0x04)
	block := checkpointTestBlock(height, blockRoot)
	result := store.ExecuteResult{MerkleRoot: merkleRoot}
	matching := SyncCheckpoint{
		Height:          height,
		BlockHash:       block.Hash(),
		BlockRoot:       blockRoot,
		StateMerkleRoot: merkleRoot,
	}

	t.Run("matching checkpoint passes and is consumed", func(t *testing.T) {
		tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{matching}}
		if err := tracker.check(block, result); err != nil {
			t.Fatalf("expected checkpoint to pass, got %v", err)
		}
		if tracker.next != 1 {
			t.Fatalf("expected checkpoint to be removed, next=%d", tracker.next)
		}
	})

	t.Run("block hash mismatch fails and keeps the checkpoint", func(t *testing.T) {
		cp := matching
		cp.BlockHash = checkpointTestHash(0x05)
		tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{cp}}
		err := tracker.check(block, result)
		if err == nil || !strings.Contains(err.Error(), "block hash") {
			t.Fatalf("expected block hash mismatch, got %v", err)
		}
		if tracker.next != 0 {
			t.Fatalf("checkpoint must not be consumed on failure, next=%d", tracker.next)
		}
	})

	t.Run("block root mismatch fails", func(t *testing.T) {
		cp := matching
		cp.BlockRoot = checkpointTestHash(0x06)
		tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{cp}}
		err := tracker.check(block, result)
		if err == nil || !strings.Contains(err.Error(), "block root") {
			t.Fatalf("expected block root mismatch, got %v", err)
		}
	})

	t.Run("state merkle root mismatch fails", func(t *testing.T) {
		cp := matching
		cp.StateMerkleRoot = checkpointTestHash(0x07)
		tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{cp}}
		err := tracker.check(block, result)
		if err == nil || !strings.Contains(err.Error(), "state merkle root") {
			t.Fatalf("expected state merkle root mismatch, got %v", err)
		}
	})

	t.Run("empty state merkle root skips the state check", func(t *testing.T) {
		cp := matching
		cp.StateMerkleRoot = common.UINT256_EMPTY
		tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{cp}}
		if err := tracker.check(block, result); err != nil {
			t.Fatalf("expected state check to be skipped, got %v", err)
		}
	})

	t.Run("future checkpoint is not consumed", func(t *testing.T) {
		cp := matching
		cp.Height = height + 1
		tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{cp}}
		if err := tracker.check(block, result); err != nil {
			t.Fatalf("expected no checkpoint check, got %v", err)
		}
		if tracker.next != 0 {
			t.Fatalf("future checkpoint must not be consumed, next=%d", tracker.next)
		}
	})

	t.Run("empty checkpoint list is ignored", func(t *testing.T) {
		tracker := &checkpointTracker{}
		if err := tracker.check(block, result); err != nil {
			t.Fatalf("expected no checkpoint check, got %v", err)
		}
	})
}

func TestCheckSyncCheckpointConsumesInOrder(t *testing.T) {
	root100 := checkpointTestHash(0x11)
	root200 := checkpointTestHash(0x12)
	block100 := checkpointTestBlock(100, root100)
	block200 := checkpointTestBlock(200, root200)
	result100 := store.ExecuteResult{MerkleRoot: checkpointTestHash(0x21)}
	result200 := store.ExecuteResult{MerkleRoot: checkpointTestHash(0x22)}

	cp100 := SyncCheckpoint{
		Height:          100,
		BlockHash:       block100.Hash(),
		BlockRoot:       root100,
		StateMerkleRoot: result100.MerkleRoot,
	}
	cp200 := SyncCheckpoint{
		Height:          200,
		BlockHash:       block200.Hash(),
		BlockRoot:       root200,
		StateMerkleRoot: result200.MerkleRoot,
	}

	tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{cp100, cp200}}
	if err := tracker.check(block100, result100); err != nil {
		t.Fatalf("height 100: %v", err)
	}
	if tracker.next != 1 {
		t.Fatalf("expected height 100 checkpoint to be consumed, next=%d", tracker.next)
	}
	if err := tracker.check(block200, result200); err != nil {
		t.Fatalf("height 200: %v", err)
	}
	if tracker.next != 2 {
		t.Fatalf("expected all checkpoints to be consumed, next=%d", tracker.next)
	}
}

func TestCheckSyncCheckpointDropsAlreadyPassed(t *testing.T) {
	const height = uint32(100)
	blockRoot := checkpointTestHash(0x31)
	merkleRoot := checkpointTestHash(0x32)
	block100 := checkpointTestBlock(height, blockRoot)
	cp100 := SyncCheckpoint{
		Height:          height,
		BlockHash:       block100.Hash(),
		BlockRoot:       blockRoot,
		StateMerkleRoot: merkleRoot,
	}

	// A node resuming above a checkpoint drops it without trying to verify it.
	block150 := checkpointTestBlock(150, checkpointTestHash(0x33))
	tracker := &checkpointTracker{checkpoints: []SyncCheckpoint{cp100}}
	if err := tracker.check(block150, store.ExecuteResult{}); err != nil {
		t.Fatalf("expected passed checkpoint to be dropped, got %v", err)
	}
	if tracker.next != 1 {
		t.Fatalf("expected passed checkpoint to be dropped, next=%d", tracker.next)
	}
}
