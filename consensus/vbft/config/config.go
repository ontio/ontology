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

package vconfig

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ontio/ontology/common/serialization"
)

var (
	Version uint32 = 1
)

type PeerConfig struct {
	Index uint32 `json:"index"`
	ID    NodeID `json:"id"`
}

type ChainConfig struct {
	Version              uint32        `json:"version"` // software version
	View                 uint32        `json:"view"`    // config-updated version
	N                    uint32        `json:"n"`       // network size
	C                    uint32        `json:"c"`       // consensus quorum
	BlockMsgDelay        time.Duration `json:"block_msg_delay"`
	HashMsgDelay         time.Duration `json:"hash_msg_delay"`
	PeerHandshakeTimeout time.Duration `json:"peer_handshake_timeout"`
	Peers                []*PeerConfig `json:"peers"`
	PosTable             []uint32      `json:"pos_table"`
	BlockCount           uint32        `json:"block_count"`
}

type VbftBlockInfo struct {
	Proposer           uint32       `json:"leader"`
	LastConfigBlockNum uint64       `json:"last_config_block_num"`
	NewChainConfig     *ChainConfig `json:"new_chain_config"`
}

const (
	VRF_SIZE            = 64 // bytes
	MAX_PROPOSER_COUNT  = 32
	MAX_ENDORSER_COUNT  = 240
	MAX_COMMITTER_COUNT = 240
)

type VRFValue [VRF_SIZE]byte

var NilVRF = VRFValue{}

func (v VRFValue) Bytes() []byte {
	return v[:]
}

func (v VRFValue) IsNil() bool {
	return bytes.Compare(v.Bytes(), NilVRF.Bytes()) == 0
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

func (cc *ChainConfig) Deserialize(r io.Reader, length int) error {
	buf := make([]byte, length)
	if _, err := r.Read(buf[:]); err != nil {
		return err
	}
	if err := json.Unmarshal(buf[:], &cc); err != nil {
		return err
	}
	return nil
}

func (pc *PeerConfig) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, pc.Index); err != nil {
		return fmt.Errorf("ChainConfig peer index length serialization failed %s", err)
	}
	if err := serialization.WriteString(w, pc.ID.String()); err != nil {
		return fmt.Errorf("ChainConfig peer ID length serialization failed %s", err)
	}
	return nil
}

func (pc *PeerConfig) Deserialize(r io.Reader) error {
	index, _ := serialization.ReadUint32(r)
	pc.Index = index

	nodeinfo, _ := serialization.ReadString(r)
	nodeid, _ := StringID(nodeinfo)
	pc.ID = nodeid
	return nil
}
