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

package states

import (
	"math/big"
)

type Status int

type RegisterSyncNodeParam struct {
	PeerPubkey string   `json:"peerPubkey"`
	Address    string   `json:"address"`
	InitPos    *big.Int `json:"initPos"`
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

type PeerPool struct {
	Index      *big.Int `json:"index"`
	PeerPubkey string   `json:"peerPubkey"`
	Status     Status   `json:"status"`
	QuitView   *big.Int `json:"quitView"`
	Address    string   `json:"address"`
	InitPos    *big.Int `json:"initPos"`
	TotalPos   *big.Int `json:"totalPos"`
}

type QuitCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

type VoteForPeerParam struct {
	Address   string              `json:"address"`
	VoteTable map[string]*big.Int `json:"voteTable"`
}

type VoteInfoPool struct {
	PeerPubkey   string   `json:"peerPubkey"`
	Address      string   `json:"address"`
	PrePos       *big.Int `json:"prePos"`
	FreezePos    *big.Int `json:"freezePos"`
	NewPos       *big.Int `json:"newPos"`
	PreFreezePos *big.Int `json:"preFreezePos"`
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
	MaxBlockChangeView   uint64 `json:"MaxBlockChangeView"`
}

type VoteCommitDposParam struct {
	Address string   `json:"address"`
	Pos     *big.Int `json:"pos"`
}

type VoteCommitInfoPool struct {
	Address string   `json:"address"`
	Pos     *big.Int `json:"pos"`
}

type GovernanceView struct {
	View       *big.Int `json:"view"`
	VoteCommit bool     `json:"voteCommit"`
}
