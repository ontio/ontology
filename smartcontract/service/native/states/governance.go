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

import "math/big"

type RegisterCandidateParam struct {
	PeerPubkey string   `json:"peerPubkey"`
	Address    string   `json:"address"`
	InitPos    *big.Int `json:"initPos"`
}

type RegisterPool struct {
	Address string   `json:"address"`
	InitPos *big.Int `json:"initPos"`
}

type ApproveCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
}

type CandidatePool struct {
	Index   *big.Int `json:"index"`
	Address string   `json:"address"`
	InitPos *big.Int `json:"initPos"`
}

type QuitCandidateParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

type RegisterSyncNodeParam struct {
	PeerPubkey string   `json:"peerPubkey"`
	Address    string   `json:"address"`
	InitPos    *big.Int `json:"initPos"`
}

type SyncNodePool struct {
	Address string   `json:"address"`
	InitPos *big.Int `json:"initPos"`
}

type QuitSyncNodeParam struct {
	PeerPubkey string `json:"peerPubkey"`
	Address    string `json:"address"`
}

type VoteForPeerParam struct {
	Address   string              `json:"address"`
	VoteTable map[string]*big.Int `json:"voteTable"`
}

type VoteInfoPool struct {
	Address   string              `json:"address"`
	Total     *big.Int            `json:"total"`
	VoteTable map[string]*big.Int `json:"voteTable"`
}
