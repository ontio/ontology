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
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
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

type PeerPoolItem struct {
	Index      uint32         //peer index
	PeerPubkey string         //peer pubkey
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

type GovernanceView struct {
	View   uint32
	Height uint32
	TxHash common.Uint256
}

func (this *GovernanceView) Serialize(w io.Writer) error {
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

func (this *GovernanceView) Deserialize(r io.Reader) error {
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
	if err := utils.WriteVarUint(w, uint64(this.N)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize n error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.C)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize c error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.K)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize k error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.L)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize l error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.BlockMsgDelay)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize block_msg_delay error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.HashMsgDelay)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize hash_msg_delay error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.PeerHandshakeTimeout)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize peer_handshake_timeout error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.MaxBlockChangeView)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize max_block_change_view error: %v", err)
	}
	return nil
}

func (this *Configuration) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize n error: %v", err)
	}
	c, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize c error: %v", err)
	}
	k, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize k error: %v", err)
	}
	l, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize l error: %v", err)
	}
	blockMsgDelay, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize blockMsgDelay error: %v", err)
	}
	hashMsgDelay, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize hashMsgDelay error: %v", err)
	}
	peerHandshakeTimeout, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize peerHandshakeTimeout error: %v", err)
	}
	maxBlockChangeView, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize maxBlockChangeView error: %v", err)
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

type SplitCurve struct {
	Yi []uint32
}

func (this *SplitCurve) Serialize(w io.Writer) error {
	if len(this.Yi) != 101 {
		return fmt.Errorf("length of split curve != 101")
	}
	if err := utils.WriteVarUint(w, uint64(len(this.Yi))); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize Yi length error: %v", err)
	}
	for _, v := range this.Yi {
		if err := utils.WriteVarUint(w, uint64(v)); err != nil {
			return fmt.Errorf("utils.WriteVarUint, serialize splitCurve error: %v", err)
		}
	}
	return nil
}

func (this *SplitCurve) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize Yi length error: %v", err)
	}
	yi := make([]uint32, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("utils.ReadVarUint, deserialize splitCurve error: %v", err)
		}
		if k > math.MaxUint32 {
			return fmt.Errorf("yi larger than max of uint32")
		}
		yi = append(yi, uint32(k))
	}
	this.Yi = yi
	return nil
}

type GlobalParam struct {
	CandidateFeeSplitNum uint32 //num of peer can receive motivation(include consensus and candidate)
	A                    uint32 //fee split to all consensus node
	B                    uint32 //fee split to all candidate node
	Yita                 uint32 //split curve coefficient
}

func (this *GlobalParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.CandidateFeeSplitNum)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize candidateFeeSplitNum error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.A)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize a error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.B)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize b error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Yita)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize yita error: %v", err)
	}
	return nil
}

func (this *GlobalParam) Deserialize(r io.Reader) error {
	candidateFeeSplitNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateFeeSplitNum error: %v", err)
	}
	a, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize a error: %v", err)
	}
	b, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize b error: %v", err)
	}
	yita, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize yita error: %v", err)
	}
	if candidateFeeSplitNum > math.MaxUint32 {
		return fmt.Errorf("minInitStake larger than max of uint32")
	}
	if a > math.MaxUint32 {
		return fmt.Errorf("a larger than max of uint32")
	}
	if b > math.MaxUint32 {
		return fmt.Errorf("b larger than max of uint32")
	}
	if yita > math.MaxUint32 {
		return fmt.Errorf("yita larger than max of uint32")
	}
	this.CandidateFeeSplitNum = uint32(candidateFeeSplitNum)
	this.A = uint32(a)
	this.B = uint32(b)
	this.Yita = uint32(yita)
	return nil
}

type CandidateSplitInfo struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint64
	Stake      uint64 //total stake, init pos + total pos
	S          uint64 //fee split weight of this peer
}

type InputPeerPoolMapParam struct {
	PeerPoolMap map[string]*PeerPoolItem
	NodeInfoMap map[string]*NodeToSideChainParams
}

func (this *InputPeerPoolMapParam) Serialize(w io.Writer) error {
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

	if err := serialization.WriteUint32(w, uint32(len(this.NodeInfoMap))); err != nil {
		return fmt.Errorf("serialization.WriteUint32, serialize PeerPoolMap length error: %v", err)
	}
	var nodeInfoMapList []*NodeToSideChainParams
	for _, v := range this.NodeInfoMap {
		nodeInfoMapList = append(nodeInfoMapList, v)
	}
	sort.SliceStable(nodeInfoMapList, func(i, j int) bool {
		return nodeInfoMapList[i].PeerPubkey > nodeInfoMapList[j].PeerPubkey
	})
	for _, v := range nodeInfoMapList {
		if err := v.Serialize(w); err != nil {
			return fmt.Errorf("serialize peerPool error: %v", err)
		}
	}
	return nil
}

func (this *InputPeerPoolMapParam) Deserialize(r io.Reader) error {
	m, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerPoolMap length error: %v", err)
	}
	peerPoolMap := make(map[string]*PeerPoolItem)
	for i := 0; uint32(i) < m; i++ {
		peerPoolItem := new(PeerPoolItem)
		if err := peerPoolItem.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		peerPoolMap[peerPoolItem.PeerPubkey] = peerPoolItem
	}

	n, err := serialization.ReadUint32(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize PeerPoolMap length error: %v", err)
	}
	nodeInfoMap := make(map[string]*NodeToSideChainParams)
	for i := 0; uint32(i) < n; i++ {
		nodeInfo := new(NodeToSideChainParams)
		if err := nodeInfo.Deserialize(r); err != nil {
			return fmt.Errorf("deserialize peerPool error: %v", err)
		}
		nodeInfoMap[nodeInfo.PeerPubkey] = nodeInfo
	}
	this.PeerPoolMap = peerPoolMap
	this.NodeInfoMap = nodeInfoMap
	return nil
}

type NodeToSideChainParams struct {
	PeerPubkey  string
	Address     common.Address
	SideChainID string
}

func (this *NodeToSideChainParams) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	if err := serialization.WriteString(w, this.SideChainID); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize sideChainID error: %v", err)
	}
	return nil
}

func (this *NodeToSideChainParams) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	sideChainID, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize sideChainID error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.SideChainID = sideChainID
	return nil
}

type SideChainID struct {
	SideChainID string
}

func (this *SideChainID) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.SideChainID); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize sideChainID error: %v", err)
	}
	return nil
}

func (this *SideChainID) Deserialize(r io.Reader) error {
	sideChainID, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize sideChainID error: %v", err)
	}
	this.SideChainID = sideChainID
	return nil
}
