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

package vbft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Ontology/common/serialization"
)

var (
	makeProposalTimeout    = 300 * time.Millisecond
	make2ndProposalTimeout = 300 * time.Millisecond
	endorseBlockTimeout    = 100 * time.Millisecond
	commitBlockTimeout     = 200 * time.Millisecond
	peerHandshakeTimeout   = 10 * time.Second

	Version uint32 = 1
)

type PeerConfig struct {
	Index uint32 `json:"index"`
	ID    NodeID `json:"id"`
}

type ChainConfig struct {
	Version       uint32        `json:"version"` // software version
	View          uint32        `json:"view"`    // config-updated version
	N             uint32        `json:"n"`       // network size
	F             uint32        `json:"f"`       // tolerated fault peers
	BlockMsgDelay time.Duration `json:"block_msg_delay"`
	HashMsgDelay  time.Duration `json:"hash_msg_delay"`
	SyncDelay     time.Duration `json:"sync_delay"`
	Peers         []*PeerConfig `json:"peers"`
	DposTable     []uint32      `json:"dpos_table"`
}

const (
	VrfSize           = 64 // bytes
	maxProposerCount  = 32
	maxEndorserCount  = 240
	maxCommitterCount = 240
)

type VRFValue [VrfSize]byte

var NilVRF = VRFValue{}

func (v VRFValue) Bytes() []byte {
	return v[:]
}

func (v VRFValue) IsNil() bool {
	return bytes.Compare(v.Bytes(), NilVRF.Bytes()) == 0
}

type BlockParticipantConfig struct {
	BlockNum    uint64
	L           uint32
	Vrf         VRFValue
	ChainConfig *ChainConfig
	Proposers   []uint32
	Endorsers   []uint32
	Committers  []uint32
}

func VerifyChainConfig(cfg *ChainConfig) error {

	// TODO

	return nil
}

//Serialize the ChainConfig
func (cc *ChainConfig) Serialize(w io.Writer) error {

	data, err := json.Marshal(cc)
	if err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func (cc *ChainConfig) Deserialize(r io.Reader) error {
	return fmt.Errorf("not implemented")
}

func (pc *PeerConfig) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, pc.Index); err != nil {
		return fmt.Errorf("ChainConfig peer index length serialization failed %s", err)
	}
	if err := serialization.WriteVarString(w, pc.ID.String()); err != nil {
		return fmt.Errorf("ChainConfig peer ID length serialization failed %s", err)
	}
	return nil
}

func (pc *PeerConfig) Deserialize(r io.Reader) error {
	index, _ := serialization.ReadUint32(r)
	pc.Index = index

	nodeinfo, _ := serialization.ReadVarString(r)
	nodeid, _ := StringID(nodeinfo)
	pc.ID = nodeid
	return nil
}
