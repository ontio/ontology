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
package utils

import (
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/ontio/ontology-crypto/vrf"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/serialization"
	vbftconfig "github.com/ontio/ontology/consensus/vbft/config"
)

type Status uint8

func (this *Status) Serialize(w io.Writer) error {
	if err := serialization.WriteUint8(w, uint8(*this)); err != nil {
		return fmt.Errorf("serialization.WriteUint8, serialize status error: %v", err)
	}
	return nil
}

func (this *Status) Deserialize(r io.Reader) error {
	status, err := serialization.ReadUint8(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint8, deserialize status error: %v", err)
	}
	*this = Status(status)
	return nil
}

type ChangeView struct {
	View   uint32
	Height uint32
	TxHash common.Uint256
}

func (this *ChangeView) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.View); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize view error: %v", err)
	}
	if err := serialization.WriteUint32(w, this.Height); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize height error: %v", err)
	}
	if err := this.TxHash.Serialize(w); err != nil {
		return fmt.Errorf("txHash.Serialize, serialize txHash error: %v", err)
	}
	return nil
}

func (this *ChangeView) Deserialize(r io.Reader) error {
	view, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize view error: %v", err)
	}
	height, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize height error: %v", err)
	}
	txHash := new(common.Uint256)
	if err := txHash.Deserialize(r); err != nil {
		return fmt.Errorf("txHash.Deserialize, deserialize txHash error: %v", err)
	}
	this.View = view
	this.Height = height
	this.TxHash = *txHash
	return nil
}

func (this *ChangeView) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint32(this.View)
	sink.WriteUint32(this.Height)
	sink.WriteHash(this.TxHash)
}

func (this *ChangeView) Deserialization(source *common.ZeroCopySource) error {
	var eof bool
	this.View, eof = source.NextUint32()
	this.Height, eof = source.NextUint32()
	this.TxHash, eof = source.NextHash()
	if eof {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type PreConfig struct {
	Configuration *Configuration
	SetView       uint32
}

func (this *PreConfig) Serialize(w io.Writer) error {
	if err := this.Configuration.Serialize(w); err != nil {
		return fmt.Errorf("WriteVarUint, serialize configuration error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.SetView)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize setView error: %v", err)
	}
	return nil
}

func (this *PreConfig) Deserialize(r io.Reader) error {
	config := new(Configuration)
	err := config.Deserialize(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize configuration error: %v", err)
	}
	setView, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize setView error: %v", err)
	}
	if setView > math.MaxUint32 {
		return fmt.Errorf("setView larger than max of uint32")
	}
	this.Configuration = config
	this.SetView = uint32(setView)
	return nil
}

type Configuration struct {
	N                    uint32
	C                    uint32
	K                    uint32
	L                    uint32
	BlockMsgDelay        uint32
	HashMsgDelay         uint32
	PeerHandshakeTimeout uint32
	MaxBlockChangeView   uint32
}

func (this *Configuration) Serialize(w io.Writer) error {
	if err := WriteVarUint(w, uint64(this.N)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize n error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.C)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize c error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.K)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize k error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.L)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize l error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.BlockMsgDelay)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize block_msg_delay error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.HashMsgDelay)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize hash_msg_delay error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.PeerHandshakeTimeout)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize peer_handshake_timeout error: %v", err)
	}
	if err := WriteVarUint(w, uint64(this.MaxBlockChangeView)); err != nil {
		return fmt.Errorf("WriteVarUint, serialize max_block_change_view error: %v", err)
	}
	return nil
}

func (this *Configuration) Deserialize(r io.Reader) error {
	n, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize n error: %v", err)
	}
	c, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize c error: %v", err)
	}
	k, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize k error: %v", err)
	}
	l, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize l error: %v", err)
	}
	blockMsgDelay, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize blockMsgDelay error: %v", err)
	}
	hashMsgDelay, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize hashMsgDelay error: %v", err)
	}
	peerHandshakeTimeout, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize peerHandshakeTimeout error: %v", err)
	}
	maxBlockChangeView, err := ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("ReadVarUint, deserialize maxBlockChangeView error: %v", err)
	}
	if n > math.MaxUint32 {
		return fmt.Errorf("n larger than max of uint32")
	}
	if c > math.MaxUint32 {
		return fmt.Errorf("c larger than max of uint32")
	}
	if k > math.MaxUint32 {
		return fmt.Errorf("k larger than max of uint32")
	}
	if l > math.MaxUint32 {
		return fmt.Errorf("l larger than max of uint32")
	}
	if blockMsgDelay > math.MaxUint32 {
		return fmt.Errorf("blockMsgDelay larger than max of uint32")
	}
	if hashMsgDelay > math.MaxUint32 {
		return fmt.Errorf("hashMsgDelay larger than max of uint32")
	}
	if peerHandshakeTimeout > math.MaxUint32 {
		return fmt.Errorf("peerHandshakeTimeout larger than max of uint32")
	}
	if maxBlockChangeView > math.MaxUint32 {
		return fmt.Errorf("maxBlockChangeView larger than max of uint32")
	}
	this.N = uint32(n)
	this.C = uint32(c)
	this.K = uint32(k)
	this.L = uint32(l)
	this.BlockMsgDelay = uint32(blockMsgDelay)
	this.HashMsgDelay = uint32(hashMsgDelay)
	this.PeerHandshakeTimeout = uint32(peerHandshakeTimeout)
	this.MaxBlockChangeView = uint32(maxBlockChangeView)
	return nil
}

type PeerPoolItem struct {
	Index      uint32         //peer index
	PeerPubkey string         //peer pubkey, run ontology wallet account
	Address    common.Address //peer owner
	Status     Status         //peer status
	InitPos    uint64         //peer initPos
	TotalPos   uint64         //total authorize pos this peer received
}

func (this *PeerPoolItem) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, this.Index); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize address error: %v", err)
	}
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := this.Address.Serialize(w); err != nil {
		return fmt.Errorf("address.Serialize, serialize address error: %v", err)
	}
	if err := this.Status.Serialize(w); err != nil {
		return fmt.Errorf("this.Status.Serialize, serialize Status error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.InitPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize initPos error: %v", err)
	}
	if err := serialization.WriteUint64(w, this.TotalPos); err != nil {
		return fmt.Errorf("serialization.WriteUint64, serialize totalPos error: %v", err)
	}
	return nil
}

func (this *PeerPoolItem) Deserialize(r io.Reader) error {
	index, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize index error: %v", err)
	}
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address := new(common.Address)
	err = address.Deserialize(r)
	if err != nil {
		return fmt.Errorf("address.Deserialize, deserialize address error: %v", err)
	}
	status := new(Status)
	err = status.Deserialize(r)
	if err != nil {
		return fmt.Errorf("status.Deserialize. deserialize status error: %v", err)
	}
	initPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize initPos error: %v", err)
	}
	totalPos, err := serialization.ReadUint64(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint64, deserialize totalPos error: %v", err)
	}
	this.Index = index
	this.PeerPubkey = peerPubkey
	this.Address = *address
	this.Status = *status
	this.InitPos = initPos
	this.TotalPos = totalPos
	return nil
}

type PeerPoolMap struct {
	PeerPoolMap map[string]*PeerPoolItem
}

func (this *PeerPoolMap) Serialize(w io.Writer) error {
	if err := serialization.WriteUint32(w, uint32(len(this.PeerPoolMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize PeerPoolMap length error: %v", err)
	}
	var peerPoolItemList []*PeerPoolItem
	for _, v := range this.PeerPoolMap {
		peerPoolItemList = append(peerPoolItemList, v)
	}
	sort.SliceStable(peerPoolItemList, func(i, j int) bool {
		return peerPoolItemList[i].PeerPubkey > peerPoolItemList[j].PeerPubkey
	})
	for _, v := range peerPoolItemList {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize peerPool error: %v", err)
		}
	}
	return nil
}

func (this *PeerPoolMap) Deserialize(r io.Reader) error {
	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerPoolMap length error: %v", err)
	}
	peerPoolMap := make(map[string]*PeerPoolItem)
	for i := 0; uint32(i) < n; i++ {
		peerPoolItem := new(PeerPoolItem)
		if err := peerPoolItem.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		peerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
	}
	this.PeerPoolMap = peerPoolMap
	return nil
}

func CheckVBFTConfig(configuration *config.VBFTConfig) error {
	if configuration.C == 0 {
		return fmt.Errorf("CheckVBFTConfig: C can not be 0 in config")
	}
	if len(configuration.Peers) > 0 && int(configuration.K) != len(configuration.Peers) {
		return fmt.Errorf("CheckVBFTConfig: K must equal to length of peer in config")
	}
	if configuration.L < 16*configuration.K || configuration.L%configuration.K != 0 {
		return fmt.Errorf("CheckVBFTConfig: L can not be less than 16*K and K must be times of L in config")
	}
	if configuration.K < 2*configuration.C+1 {
		return fmt.Errorf("CheckVBFTConfig: K can not be less than 2*C+1 in config")
	}
	if configuration.N < configuration.K || configuration.K < 7 {
		return fmt.Errorf("CheckVBFTConfig: config not match N >= K >= 7")
	}
	if configuration.BlockMsgDelay < 5000 {
		return fmt.Errorf("CheckVBFTConfig: BlockMsgDelay must >= 5000")
	}
	if configuration.HashMsgDelay < 5000 {
		return fmt.Errorf("CheckVBFTConfig: HashMsgDelay must >= 5000")
	}
	if configuration.PeerHandshakeTimeout < 10 {
		return fmt.Errorf("CheckVBFTConfig: PeerHandshakeTimeout must >= 10")
	}
	if configuration.MinInitStake < 10000 {
		return fmt.Errorf("CheckVBFTConfig: MinInitStake must >= 10000")
	}
	if len(configuration.VrfProof) < 128 {
		return fmt.Errorf("CheckVBFTConfig: VrfProof must >= 128")
	}
	if len(configuration.VrfValue) < 128 {
		return fmt.Errorf("CheckVBFTConfig: VrfValue must >= 128")
	}

	indexMap := make(map[uint32]struct{})
	peerPubkeyMap := make(map[string]struct{})
	for _, peer := range configuration.Peers {
		_, ok := indexMap[peer.Index]
		if ok {
			return fmt.Errorf("CheckVBFTConfig: peer index is duplicated")
		}
		indexMap[peer.Index] = struct{}{}

		_, ok = peerPubkeyMap[peer.PeerPubkey]
		if ok {
			return fmt.Errorf("CheckVBFTConfig: peerPubkey is duplicated")
		}
		peerPubkeyMap[peer.PeerPubkey] = struct{}{}

		if peer.Index <= 0 {
			return fmt.Errorf("CheckVBFTConfig: peer index in config must > 0")
		}
		//check peerPubkey
		if err := ValidatePeerPubKeyFormat(peer.PeerPubkey); err != nil {
			return fmt.Errorf("CheckVBFTConfig: invalid peer pubkey")
		}
		_, err := common.AddressFromBase58(peer.Address)
		if err != nil {
			return fmt.Errorf("CheckVBFTConfig: address format error: %v", err)
		}
	}
	return nil
}

func ValidatePeerPubKeyFormat(pubkey string) error {
	pk, err := vbftconfig.Pubkey(pubkey)
	if err != nil {
		return fmt.Errorf("failed to parse pubkey")
	}
	if !vrf.ValidatePublicKey(pk) {
		return fmt.Errorf("invalid for VRF")
	}
	return nil
}
