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
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/store"
	"github.com/ontio/ontology/core/types"
)

// SyncCheckpoint pins a historical block that a node must reproduce exactly
// while syncing. A node that derives different data for the same height is on
// a different chain (or executed the block differently) and must not persist it.
type SyncCheckpoint struct {
	Height          uint32         // block height
	BlockHash       common.Uint256 // block.Hash(), i.e. the header hash
	BlockRoot       common.Uint256 // block.Header.BlockRoot
	StateMerkleRoot common.Uint256 // ExecutionResult.MerkleRoot
}

var mainNetSyncCheckpoints = []SyncCheckpoint{
	{
		Height:          20942351,
		BlockHash:       mustHexHash("8380568b603d84e52bbfe7d014a4938bc43d539a8d17f5c271d517336f550c8f"),
		BlockRoot:       mustHexHash("107736222a32ff31536059d5ab255655894edcce2a0e50b0425c2f8408089f44"),
		StateMerkleRoot: mustHexHash("ee0e74fb6cfc4bee6495ec3e020aafbebc1c3d654326b77fba37ce9402f428a8"),
	},
	{
		Height:          20951014,
		BlockHash:       mustHexHash("fd3dfb093f2a6155ce272af5f08e577e50038c1ba4cee57bfed8e8fd6e73384a"),
		BlockRoot:       mustHexHash("723206dad9e9785b2add0a844c052d766f53dc8eeb21b57edb46620cc7729215"),
		StateMerkleRoot: mustHexHash("2cd7b6410e33c0ed61bf21fc57712560a6176b792535684b2636e116b1273937"),
	},
}

// polarisNetSyncCheckpoints must be sorted by Height, strictly increasing.
// See mainNetSyncCheckpoints for entry semantics.
var polarisNetSyncCheckpoints = []SyncCheckpoint{
	{
		Height:          2643676,
		BlockHash:       mustHexHash("3fed536527f658f60e39cc93720329899c7de960d6fa6169f294cda1691b1815"),
		BlockRoot:       mustHexHash("bf028a748386aa8bca19526706e30f42e7cc4979a5a8c799690f52563dd23545"),
		StateMerkleRoot: mustHexHash("c3a06f45d08079d397befdc8c1ccaec4bbdaae12761c6e5c47948e806e810b67"),
	},
}

// syncCheckpointsForNetwork returns the checkpoint list for a network. The
// network id is fixed at startup, so this is only called once per ledger.
func syncCheckpointsForNetwork(networkId uint32) []SyncCheckpoint {
	switch networkId {
	case config.NETWORK_ID_MAIN_NET:
		return mainNetSyncCheckpoints
	case config.NETWORK_ID_POLARIS_NET:
		return polarisNetSyncCheckpoints
	default:
		return nil
	}
}

// mustHexHash converts a hex-encoded hash to common.Uint256. It is meant for
// building the checkpoint literals above and panics on malformed input, which
// surfaces configuration mistakes at startup.
func mustHexHash(s string) common.Uint256 {
	h, err := common.Uint256FromHexString(s)
	if err != nil {
		panic(fmt.Sprintf("invalid checkpoint hash %q: %v", s, err))
	}
	return h
}

// checkpointTracker consumes a height-sorted checkpoint list as blocks are
// applied. submitBlock is single-threaded, so no lock is needed.
type checkpointTracker struct {
	checkpoints []SyncCheckpoint
	next        int
}

func newCheckpointTracker(networkId uint32) *checkpointTracker {
	return &checkpointTracker{checkpoints: syncCheckpointsForNetwork(networkId)}
}

func (t *checkpointTracker) check(block *types.Block, result store.ExecuteResult) error {
	height := block.Header.Height
	for t.next < len(t.checkpoints) && t.checkpoints[t.next].Height < height {
		t.next++
	}
	if t.next >= len(t.checkpoints) {
		return nil
	}

	cp := t.checkpoints[t.next]
	if cp.Height > height {
		return nil
	}

	if err := cp.verify(block, result); err != nil {
		return err
	}
	t.next++
	return nil
}

func (cp SyncCheckpoint) verify(block *types.Block, result store.ExecuteResult) error {
	if blockHash := block.Hash(); blockHash != cp.BlockHash {
		return fmt.Errorf("sync checkpoint mismatch at height %d: block hash expected %s, got %s",
			cp.Height, cp.BlockHash.ToHexString(), blockHash.ToHexString())
	}
	if block.Header.BlockRoot != cp.BlockRoot {
		return fmt.Errorf("sync checkpoint mismatch at height %d: block root expected %s, got %s",
			cp.Height, cp.BlockRoot.ToHexString(), block.Header.BlockRoot.ToHexString())
	}

	if cp.StateMerkleRoot != common.UINT256_EMPTY && result.MerkleRoot != cp.StateMerkleRoot {
		return fmt.Errorf("sync checkpoint mismatch at height %d: state merkle root expected %s, got %s",
			cp.Height, cp.StateMerkleRoot.ToHexString(), result.MerkleRoot.ToHexString())
	}
	return nil
}
