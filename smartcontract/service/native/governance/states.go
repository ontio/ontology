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

package governance

import (
	"math/big"
)

type Status int

type RegisterSyncNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
	InitPos    uint64 `json:"initPos"`
}

type ApproveSyncNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
}

type QuitNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

type RegisterCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

type ApproveCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
}

type PeerPoolList struct {
	Peers []*PeerPool `json:"peers"`
}

type PeerPoolMap struct {
	PeerPoolMap map[string]*PeerPool `json:"peerPoolMap"`
}

type PeerPool struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Status     Status `json:"status"`
	Address    string `json:"address"`
	InitPos    uint64 `json:"initPos"`
	TotalPos   uint64 `json:"totalPos"`
}

type QuitCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

type VoteForPeerParam struct {
	Address   string           `json:"address"`
	VoteTable map[string]int64 `json:"voteTable"`
}

type VoteInfoPool struct {
	PeerPubkey   string `json:"peerPubkey"`
	Address      string `json:"address"`
	PrePos       uint64 `json:"prePos"`
	FreezePos    uint64 `json:"freezePos"`
	NewPos       uint64 `json:"newPos"`
	PreFreezePos uint64 `json:"preFreezePos"`
}

type PeerStakeInfo struct {
	Index      uint32 `json:"index"`
	PeerPubkey string `json:"peerPubkey"`
	Stake      uint64 `json:"stake"`
}

type Configuration struct {
	N                    uint32 `json:"n"`
	C                    uint32 `json:"c"`
	K                    uint32 `json:"k"`
	L                    uint32 `json:"l"`
	BlockMsgDelay        uint32 `json:"block_msg_delay"`
	HashMsgDelay         uint32 `json:"hash_msg_delay"`
	PeerHandshakeTimeout uint32 `json:"peer_handshake_timeout"`
	MaxBlockChangeView   uint32 `json:"MaxBlockChangeView"`
}

type VoteCommitDposParam struct {
	Address string `json:"address"`
	Pos     int64  `json:"pos"`
}

type VoteCommitInfoPool struct {
	Address string `json:"address"`
	Pos     uint64 `json:"pos"`
}

type GovernanceView struct {
	View       *big.Int `json:"view"`
	VoteCommit bool     `json:"voteCommit"`
}
