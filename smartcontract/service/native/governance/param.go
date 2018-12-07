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

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"sort"
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
		return fmt.Errorf("serialization.WriteUint32, serialize height error: %v", err)
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

	if err != nil {
		return fmt.Errorf("serialization.ReadUint32, deserialize height error: %v", err)
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

type CandidateSplitInfo struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint64
	Stake      uint64 //total stake, init pos + total pos
	S          uint64 //fee split weight of this peer
}

type SideChainNodeInfo struct {
	SideChainID string
	NodeInfoMap map[string]*NodeToSideChainParams
}

func (this *SideChainNodeInfo) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.SideChainID); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize sideChainID error: %v", err)
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

func (this *SideChainNodeInfo) Deserialize(r io.Reader) error {
	sideChainID, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize sideChainID error: %v", err)
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
	this.SideChainID = sideChainID
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
	if err := utils.WriteAddress(w, this.Address); err != nil {
		return fmt.Errorf("utils.WriteAddress, serialize address error: %v", err)
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

type SyncAddress struct {
	SyncAddress common.Address
}

func (this *SyncAddress) Serialize(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.SyncAddress)
}

func (this *SyncAddress) Deserialize(source *common.ZeroCopySource) error {
	var err error
	this.SyncAddress, err = utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("deserialize address error:%s", err)
	}
	return nil
}

type CommitDposParam struct {
	GovernanceView    *GovernanceView
	PeerPoolMap       *PeerPoolMap
	SideChainNodeInfo *SideChainNodeInfo
	Configuration     *Configuration
	GlobalParam       *GlobalParam
	GlobalParam2      *GlobalParam2
	SplitCurve        *SplitCurve
}

func (this *CommitDposParam) Serialize(w io.Writer) error {
	if err := this.GovernanceView.Serialize(w); err != nil {
		return fmt.Errorf("this.GovernanceView.Serialize, serialize GovernanceView error: %v", err)
	}

	if err := this.PeerPoolMap.Serialize(w); err != nil {
		return fmt.Errorf("this.PeerPoolMap.Serialize, serialize PeerPoolMap error: %v", err)
	}

	if err := this.SideChainNodeInfo.Serialize(w); err != nil {
		return fmt.Errorf("this.SideChainNodeInfo.Serialize, serialize SideChainNodeInfo error: %v", err)
	}

	if err := this.Configuration.Serialize(w); err != nil {
		return fmt.Errorf("this.Configuration.Serialize, serialize Configuration error: %v", err)
	}

	if err := this.GlobalParam.Serialize(w); err != nil {
		return fmt.Errorf("this.GlobalParam.Serialize, serialize GlobalParam error: %v", err)
	}
	if err := this.GlobalParam2.Serialize(w); err != nil {
		return fmt.Errorf("this.GlobalParam2.Serialize, serialize GlobalParam2 error: %v", err)
	}

	if err := this.SplitCurve.Serialize(w); err != nil {
		return fmt.Errorf("this.SplitCurve.Serialize, serialize SplitCurve error: %v", err)
	}
	return nil
}

func (this *CommitDposParam) Deserialize(r io.Reader) error {
	governanceView := new(GovernanceView)
	err := governanceView.Deserialize(r)
	if err != nil {
		return fmt.Errorf("governanceView.Deserialize, deserialize governanceView error: %v", err)
	}

	peerPoolMap := new(PeerPoolMap)
	err = peerPoolMap.Deserialize(r)
	if err != nil {
		return fmt.Errorf("peerPoolMap.Deserialize, deserialize peerPoolMap error: %v", err)
	}
	sideChainNodeInfo := new(SideChainNodeInfo)
	err = sideChainNodeInfo.Deserialize(r)
	if err != nil {
		return fmt.Errorf("sideChainNodeInfo.Deserialize, deserialize sideChainNodeInfo error: %v", err)
	}
	configuration := new(Configuration)
	err = configuration.Deserialize(r)
	if err != nil {
		return fmt.Errorf("configuration.Deserialize, deserialize configuration error: %v", err)
	}
	globalParam := new(GlobalParam)
	err = globalParam.Deserialize(r)
	if err != nil {
		return fmt.Errorf("globalParam.Deserialize, deserialize globalParam error: %v", err)
	}
	globalParam2 := new(GlobalParam2)
	err = globalParam2.Deserialize(r)
	if err != nil {
		return fmt.Errorf("globalParam2.Deserialize, deserialize globalParam2 error: %v", err)
	}
	splitCurve := new(SplitCurve)
	err = splitCurve.Deserialize(r)
	if err != nil {
		return fmt.Errorf("splitCurve.Deserialize, deserialize splitCurve error: %v", err)
	}
	this.GovernanceView = governanceView
	this.PeerPoolMap = peerPoolMap
	this.SideChainNodeInfo = sideChainNodeInfo
	this.Configuration = configuration
	this.GlobalParam = globalParam
	this.GlobalParam2 = globalParam2
	this.SplitCurve = splitCurve
	return nil
}

type QuitNodeParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *QuitNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, deserialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	return nil
}

func (this *QuitNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type ApproveCandidateParam struct {
	PeerPubkey string
}

func (this *ApproveCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	return nil
}

func (this *ApproveCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type RejectCandidateParam struct {
	PeerPubkey string
}

func (this *RejectCandidateParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	return nil
}

func (this *RejectCandidateParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type BlackNodeParam struct {
	PeerPubkeyList []string
}

func (this *BlackNodeParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(len(this.PeerPubkeyList))); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize peerPubkeyList length error: %v", err)
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
		}
	}
	return nil
}

func (this *BlackNodeParam) Deserialize(r io.Reader) error {
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize peerPubkeyList length error: %v", err)
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	this.PeerPubkeyList = peerPubkeyList
	return nil
}

type WhiteNodeParam struct {
	PeerPubkey string
}

func (this *WhiteNodeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	return nil
}

func (this *WhiteNodeParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type AuthorizeForPeerParam struct {
	Address        common.Address
	PeerPubkeyList []string
	PosList        []uint32
}

func (this *AuthorizeForPeerParam) Serialize(w io.Writer) error {
	if len(this.PeerPubkeyList) > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	if len(this.PeerPubkeyList) != len(this.PosList) {
		return fmt.Errorf("length of PeerPubkeyList != length of PosList")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.PeerPubkeyList))); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize peerPubkeyList length error: %v", err)
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
		}
	}
	if err := utils.WriteVarUint(w, uint64(len(this.PosList))); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize posList length error: %v", err)
	}
	for _, v := range this.PosList {
		if err := utils.WriteVarUint(w, uint64(v)); err != nil {
			return fmt.Errorf("utils.WriteVarUint, serialize pos error: %v", err)
		}
	}
	return nil
}

func (this *AuthorizeForPeerParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize peerPubkeyList length error: %v", err)
	}
	if n > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize posList length error: %v", err)
	}
	posList := make([]uint32, 0)
	for i := 0; uint64(i) < m; i++ {
		k, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("utils.ReadVarUint, deserialize pos error: %v", err)
		}
		if k > math.MaxUint32 {
			return fmt.Errorf("pos larger than max of uint32")
		}
		posList = append(posList, uint32(k))
	}
	if m != n {
		return fmt.Errorf("length of PeerPubkeyList != length of PosList")
	}
	this.Address = address
	this.PeerPubkeyList = peerPubkeyList
	this.PosList = posList
	return nil
}

type WithdrawParam struct {
	Address        common.Address
	PeerPubkeyList []string
	WithdrawList   []uint32
}

func (this *WithdrawParam) Serialize(w io.Writer) error {
	if len(this.PeerPubkeyList) > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	if len(this.PeerPubkeyList) != len(this.WithdrawList) {
		return fmt.Errorf("length of PeerPubkeyList != length of WithdrawList, contract params error")
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, address address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(len(this.PeerPubkeyList))); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize peerPubkeyList length error: %v", err)
	}
	for _, v := range this.PeerPubkeyList {
		if err := serialization.WriteString(w, v); err != nil {
			return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
		}
	}
	if err := utils.WriteVarUint(w, uint64(len(this.WithdrawList))); err != nil {
		return fmt.Errorf("serialization.WriteVarUint, serialize withdrawList length error: %v", err)
	}
	for _, v := range this.WithdrawList {
		if err := utils.WriteVarUint(w, uint64(v)); err != nil {
			return fmt.Errorf("utils.WriteVarUint, serialize withdraw error: %v", err)
		}
	}
	return nil
}

func (this *WithdrawParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	n, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize peerPubkeyList length error: %v", err)
	}
	if n > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := serialization.ReadString(r)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize withdrawList length error: %v", err)
	}
	withdrawList := make([]uint32, 0)
	for i := 0; uint64(i) < m; i++ {
		k, err := utils.ReadVarUint(r)
		if err != nil {
			return fmt.Errorf("utils.ReadVarUint, deserialize withdraw error: %v", err)
		}
		if k > math.MaxUint32 {
			return fmt.Errorf("pos larger than max of uint32")
		}
		withdrawList = append(withdrawList, uint32(k))
	}
	if m != n {
		return fmt.Errorf("length of PeerPubkeyList != length of WithdrawList, contract params error")
	}
	this.Address = address
	this.PeerPubkeyList = peerPubkeyList
	this.WithdrawList = withdrawList
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

type PreConfig struct {
	Configuration *Configuration
	SetView       uint32
}

func (this *PreConfig) Serialize(w io.Writer) error {
	if err := this.Configuration.Serialize(w); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize configuration error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.SetView)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize setView error: %v", err)
	}
	return nil
}

func (this *PreConfig) Deserialize(r io.Reader) error {
	config := new(Configuration)
	err := config.Deserialize(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize configuration error: %v", err)
	}
	setView, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize setView error: %v", err)
	}
	if setView > math.MaxUint32 {
		return fmt.Errorf("setView larger than max of uint32")
	}
	this.Configuration = config
	this.SetView = uint32(setView)
	return nil
}

type GlobalParam struct {
	CandidateFee uint64 //unit: 10^-9 ong
	MinInitStake uint32 //min init pos
	CandidateNum uint32 //num of candidate and consensus node
	PosLimit     uint32 //authorize pos limit is initPos*posLimit
	A            uint32 //fee split to all consensus node
	B            uint32 //fee split to all candidate node
	Yita         uint32 //split curve coefficient
	Penalty      uint32 //authorize pos penalty percentage
}

func (this *GlobalParam) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, this.CandidateFee); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize candidateFee error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.MinInitStake)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize minInitStake error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.CandidateNum)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize candidateNum error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.PosLimit)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize posLimit error: %v", err)
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
	if err := utils.WriteVarUint(w, uint64(this.Penalty)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize penalty error: %v", err)
	}
	return nil
}

func (this *GlobalParam) Deserialize(r io.Reader) error {
	candidateFee, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateFee error: %v", err)
	}
	minInitStake, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize minInitStake error: %v", err)
	}
	candidateNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateNum error: %v", err)
	}
	posLimit, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize posLimit error: %v", err)
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
	penalty, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize penalty error: %v", err)
	}
	if minInitStake > math.MaxUint32 {
		return fmt.Errorf("minInitStake larger than max of uint32")
	}
	if candidateNum > math.MaxUint32 {
		return fmt.Errorf("candidateNum larger than max of uint32")
	}
	if posLimit > math.MaxUint32 {
		return fmt.Errorf("posLimit larger than max of uint32")
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
	if penalty > math.MaxUint32 {
		return fmt.Errorf("penalty larger than max of uint32")
	}
	this.CandidateFee = candidateFee
	this.MinInitStake = uint32(minInitStake)
	this.CandidateNum = uint32(candidateNum)
	this.PosLimit = uint32(posLimit)
	this.A = uint32(a)
	this.B = uint32(b)
	this.Yita = uint32(yita)
	this.Penalty = uint32(penalty)
	return nil
}

type GlobalParam2 struct {
	MinAuthorizePos      uint32 //min ONT of each authorization, 500 default
	CandidateFeeSplitNum uint32 //num of peer can receive motivation(include consensus and candidate)
	Field1               []byte //reserved field
	Field2               []byte //reserved field
	Field3               []byte //reserved field
	Field4               []byte //reserved field
	Field5               []byte //reserved field
	Field6               []byte //reserved field
}

func (this *GlobalParam2) Serialize(w io.Writer) error {
	if err := utils.WriteVarUint(w, uint64(this.MinAuthorizePos)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize minAuthorizePos error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.CandidateFeeSplitNum)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize candidateFeeSplitNum error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field1); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field1 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field2); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field2 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field3); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field3 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field4); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field4 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field5); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field5 error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Field6); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize field6 error: %v", err)
	}
	return nil
}

func (this *GlobalParam2) Deserialize(r io.Reader) error {
	minAuthorizePos, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize minAuthorizePos error: %v", err)
	}
	candidateFeeSplitNum, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateFeeSplitNum error: %v", err)
	}
	field1, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize field1 error: %v", err)
	}
	field2, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize field2 error: %v", err)
	}
	field3, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize field3 error: %v", err)
	}
	field4, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize field4 error: %v", err)
	}
	field5, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize field5 error: %v", err)
	}
	field6, err := serialization.ReadVarBytes(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize field6 error: %v", err)
	}
	if minAuthorizePos > math.MaxUint32 {
		return fmt.Errorf("minAuthorizePos larger than max of uint32")
	}
	if candidateFeeSplitNum > math.MaxUint32 {
		return fmt.Errorf("candidateFeeSplitNum larger than max of uint32")
	}

	this.MinAuthorizePos = uint32(minAuthorizePos)
	this.CandidateFeeSplitNum = uint32(candidateFeeSplitNum)
	this.Field1 = field1
	this.Field2 = field2
	this.Field3 = field3
	this.Field4 = field4
	this.Field5 = field5
	this.Field6 = field6
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

type TransferPenaltyParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *TransferPenaltyParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	return nil
}

func (this *TransferPenaltyParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type WithdrawOngParam struct {
	Address common.Address
}

func (this *WithdrawOngParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	return nil
}

func (this *WithdrawOngParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.Address = address
	return nil
}

type ChangeMaxAuthorizationParam struct {
	PeerPubkey   string
	Address      common.Address
	MaxAuthorize uint32
}

func (this *ChangeMaxAuthorizationParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.MaxAuthorize)); err != nil {
		return fmt.Errorf("utils.WriteVarUint, serialize maxAuthorize error: %v", err)
	}
	return nil
}

func (this *ChangeMaxAuthorizationParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	maxAuthorize, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize maxAuthorize error: %v", err)
	}
	if maxAuthorize > math.MaxUint32 {
		return fmt.Errorf("peerCost larger than max of uint32")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.MaxAuthorize = uint32(maxAuthorize)
	return nil
}

type SetPeerCostParam struct {
	PeerPubkey string
	Address    common.Address
	PeerCost   uint32
}

func (this *SetPeerCostParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.PeerCost)); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize peerCost error: %v", err)
	}
	return nil
}

func (this *SetPeerCostParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	peerCost, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize peerCost error: %v", err)
	}
	if peerCost > math.MaxUint32 {
		return fmt.Errorf("peerCost larger than max of uint32")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.PeerCost = uint32(peerCost)
	return nil
}

type WithdrawFeeParam struct {
	Address common.Address
}

func (this *WithdrawFeeParam) Serialize(w io.Writer) error {
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	return nil
}

func (this *WithdrawFeeParam) Deserialize(r io.Reader) error {
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.Address = address
	return nil
}

type PromisePos struct {
	PeerPubkey string
	PromisePos uint64
}

func (this *PromisePos) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := utils.WriteVarUint(w, this.PromisePos); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize promisePos error: %v", err)
	}
	return nil
}

func (this *PromisePos) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	promisePos, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize promisePos error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.PromisePos = promisePos
	return nil
}

type ChangeInitPosParam struct {
	PeerPubkey string
	Address    common.Address
	Pos        uint32
}

func (this *ChangeInitPosParam) Serialize(w io.Writer) error {
	if err := serialization.WriteString(w, this.PeerPubkey); err != nil {
		return fmt.Errorf("serialization.WriteString, serialize peerPubkey error: %v", err)
	}
	if err := serialization.WriteVarBytes(w, this.Address[:]); err != nil {
		return fmt.Errorf("serialization.WriteVarBytes, serialize address error: %v", err)
	}
	if err := utils.WriteVarUint(w, uint64(this.Pos)); err != nil {
		return fmt.Errorf("serialization.WriteBool, serialize pos error: %v", err)
	}
	return nil
}

func (this *ChangeInitPosParam) Deserialize(r io.Reader) error {
	peerPubkey, err := serialization.ReadString(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.ReadAddress(r)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	pos, err := utils.ReadVarUint(r)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize pos error: %v", err)
	}
	if pos > math.MaxUint32 {
		return fmt.Errorf("pos larger than max of uint32")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.Pos = uint32(pos)
	return nil
}
