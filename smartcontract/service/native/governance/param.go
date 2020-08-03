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
	"math"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type RegisterCandidateParam struct {
	PeerPubkey string
	Address    common.Address
	InitPos    uint32
	Caller     []byte
	KeyNo      uint32
}

func (this *RegisterCandidateParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
	utils.EncodeVarUint(sink, uint64(this.InitPos))
	sink.WriteVarBytes(this.Caller)
	utils.EncodeVarUint(sink, uint64(this.KeyNo))
}

func (this *RegisterCandidateParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, _, irregular, eof := source.NextString()
	if irregular || eof {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey irregular: %v, eof: %v", irregular, eof)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	initPos, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize initPos error: %v", err)
	}
	if initPos > math.MaxUint32 {
		return fmt.Errorf("initPos larger than max of uint32")
	}
	caller, _, irregular, eof := source.NextVarBytes()
	if irregular || eof {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize caller irregular: %v, eof: %v", irregular, eof)
	}
	keyNo, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize keyNo error: %v", err)
	}
	if keyNo > math.MaxUint32 {
		return fmt.Errorf("initPos larger than max of uint32")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.InitPos = uint32(initPos)
	this.Caller = caller
	this.KeyNo = uint32(keyNo)
	return nil
}

type UnRegisterCandidateParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *UnRegisterCandidateParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
}

func (this *UnRegisterCandidateParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	return nil
}

type QuitNodeParam struct {
	PeerPubkey string
	Address    common.Address
}

func (this *QuitNodeParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
}

func (this *QuitNodeParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
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

func (this *ApproveCandidateParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
}

func (this *ApproveCandidateParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type RejectCandidateParam struct {
	PeerPubkey string
}

func (this *RejectCandidateParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
}

func (this *RejectCandidateParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	this.PeerPubkey = peerPubkey
	return nil
}

type BlackNodeParam struct {
	PeerPubkeyList []string
}

func (this *BlackNodeParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, uint64(len(this.PeerPubkeyList)))
	for _, v := range this.PeerPubkeyList {
		sink.WriteString(v)
	}
}

func (this *BlackNodeParam) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize peerPubkeyList length error: %v", err)
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, _, irregular, eof := source.NextString()
		if irregular || eof {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey irregular:%v, eof: %v", irregular, eof)
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	this.PeerPubkeyList = peerPubkeyList
	return nil
}

type WhiteNodeParam struct {
	PeerPubkey string
}

func (this *WhiteNodeParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
}

func (this *WhiteNodeParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
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

func (this *AuthorizeForPeerParam) Serialization(sink *common.ZeroCopySink) error {
	if len(this.PeerPubkeyList) > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	if len(this.PeerPubkeyList) != len(this.PosList) {
		return fmt.Errorf("length of PeerPubkeyList != length of PosList")
	}
	sink.WriteVarBytes(this.Address[:])
	utils.EncodeVarUint(sink, uint64(len(this.PeerPubkeyList)))
	for _, v := range this.PeerPubkeyList {
		sink.WriteString(v)
	}
	utils.EncodeVarUint(sink, uint64(len(this.PosList)))
	for _, v := range this.PosList {
		utils.EncodeVarUint(sink, uint64(v))
	}
	return nil
}

func (this *AuthorizeForPeerParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize peerPubkeyList length error: %v", err)
	}
	if n > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, _, irregular, eof := source.NextString()
		if irregular || eof {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey irregular: %v,eof: %v", irregular, eof)
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize posList length error: %v", err)
	}
	posList := make([]uint32, 0)
	for i := 0; uint64(i) < m; i++ {
		k, err := utils.DecodeVarUint(source)
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

func (this *WithdrawParam) Serialization(sink *common.ZeroCopySink) error {
	if len(this.PeerPubkeyList) > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	if len(this.PeerPubkeyList) != len(this.WithdrawList) {
		return fmt.Errorf("length of PeerPubkeyList != length of WithdrawList, contract params error")
	}
	sink.WriteVarBytes(this.Address[:])

	utils.EncodeVarUint(sink, uint64(len(this.PeerPubkeyList)))
	for _, v := range this.PeerPubkeyList {
		sink.WriteString(v)
	}
	utils.EncodeVarUint(sink, uint64(len(this.WithdrawList)))
	for _, v := range this.WithdrawList {
		utils.EncodeVarUint(sink, uint64(v))
	}
	return nil
}

func (this *WithdrawParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize peerPubkeyList length error: %v", err)
	}
	if n > 1024 {
		return fmt.Errorf("length of input list > 1024")
	}
	peerPubkeyList := make([]string, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := utils.DecodeString(source)
		if err != nil {
			return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
		}
		peerPubkeyList = append(peerPubkeyList, k)
	}
	m, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize withdrawList length error: %v", err)
	}
	withdrawList := make([]uint32, 0)
	for i := 0; uint64(i) < m; i++ {
		k, err := utils.DecodeVarUint(source)
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

func (this *Configuration) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, uint64(this.N))
	utils.EncodeVarUint(sink, uint64(this.C))
	utils.EncodeVarUint(sink, uint64(this.K))
	utils.EncodeVarUint(sink, uint64(this.L))
	utils.EncodeVarUint(sink, uint64(this.BlockMsgDelay))
	utils.EncodeVarUint(sink, uint64(this.HashMsgDelay))
	utils.EncodeVarUint(sink, uint64(this.PeerHandshakeTimeout))
	utils.EncodeVarUint(sink, uint64(this.MaxBlockChangeView))
}

func (this *Configuration) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize n error: %v", err)
	}
	c, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize c error: %v", err)
	}
	k, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize k error: %v", err)
	}
	l, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize l error: %v", err)
	}
	blockMsgDelay, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize blockMsgDelay error: %v", err)
	}
	hashMsgDelay, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize hashMsgDelay error: %v", err)
	}
	peerHandshakeTimeout, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize peerHandshakeTimeout error: %v", err)
	}
	maxBlockChangeView, err := utils.DecodeVarUint(source)
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

func (this *PreConfig) Serialization(sink *common.ZeroCopySink) {
	this.Configuration.Serialization(sink)
	utils.EncodeVarUint(sink, uint64(this.SetView))
}

func (this *PreConfig) Deserialization(source *common.ZeroCopySource) error {
	config := new(Configuration)
	err := config.Deserialization(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize configuration error: %v", err)
	}
	setView, err := utils.DecodeVarUint(source)
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

func (this *GlobalParam) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.CandidateFee)
	utils.EncodeVarUint(sink, uint64(this.MinInitStake))
	utils.EncodeVarUint(sink, uint64(this.CandidateNum))
	utils.EncodeVarUint(sink, uint64(this.PosLimit))
	utils.EncodeVarUint(sink, uint64(this.A))
	utils.EncodeVarUint(sink, uint64(this.B))
	utils.EncodeVarUint(sink, uint64(this.Yita))
	utils.EncodeVarUint(sink, uint64(this.Penalty))
}

func (this *GlobalParam) Deserialization(source *common.ZeroCopySource) error {
	candidateFee, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateFee error: %v", err)
	}
	minInitStake, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize minInitStake error: %v", err)
	}
	candidateNum, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateNum error: %v", err)
	}
	posLimit, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize posLimit error: %v", err)
	}
	a, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize a error: %v", err)
	}
	b, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize b error: %v", err)
	}
	yita, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize yita error: %v", err)
	}
	penalty, err := utils.DecodeVarUint(source)
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
	DappFee              uint32 //fee split to dapp bonus
	Field2               []byte //reserved field
	Field3               []byte //reserved field
	Field4               []byte //reserved field
	Field5               []byte //reserved field
	Field6               []byte //reserved field
}

func (this *GlobalParam2) Serialization(sink *common.ZeroCopySink) error {
	if this.MinAuthorizePos == 0 {
		return fmt.Errorf("globalParam2.MinAuthorizePos can not be 0")
	}
	if this.DappFee > 100 {
		return fmt.Errorf("globalParam2.DappFee must <= 100")
	}
	utils.EncodeVarUint(sink, uint64(this.MinAuthorizePos))
	utils.EncodeVarUint(sink, uint64(this.CandidateFeeSplitNum))

	utils.EncodeVarUint(sink, uint64(this.DappFee))
	sink.WriteVarBytes(this.Field2)
	sink.WriteVarBytes(this.Field3)
	sink.WriteVarBytes(this.Field4)
	sink.WriteVarBytes(this.Field5)
	sink.WriteVarBytes(this.Field6)
	return nil
}

func (this *GlobalParam2) Deserialization(source *common.ZeroCopySource) error {
	minAuthorizePos, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize minAuthorizePos error: %v", err)
	}
	candidateFeeSplitNum, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize candidateFeeSplitNum error: %v", err)
	}
	dappFee, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize dappFee error: %v", err)
	}
	field2, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize field2 error: %v", err)
	}
	field3, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize field3 error: %v", err)
	}
	field4, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize field4 error: %v", err)
	}
	field5, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarBytes, deserialize field5 error: %v", err)
	}
	field6, err := utils.DecodeVarBytes(source)
	if err != nil {
		return fmt.Errorf("utils.ReadVarUint, deserialize field6 error: %v", err)
	}
	if minAuthorizePos > math.MaxUint32 {
		return fmt.Errorf("minAuthorizePos larger than max of uint32")
	}
	if candidateFeeSplitNum > math.MaxUint32 {
		return fmt.Errorf("candidateFeeSplitNum larger than max of uint32")
	}
	if minAuthorizePos == 0 {
		return fmt.Errorf("globalParam2.MinAuthorizePos can not be 0")
	}
	if dappFee > 100 {
		return fmt.Errorf("globalParam2.DappFee must <= 100")
	}
	this.MinAuthorizePos = uint32(minAuthorizePos)
	this.CandidateFeeSplitNum = uint32(candidateFeeSplitNum)
	this.DappFee = uint32(dappFee)
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

func (this *SplitCurve) Serialization(sink *common.ZeroCopySink) error {
	if len(this.Yi) != 101 {
		return fmt.Errorf("length of split curve != 101")
	}
	utils.EncodeVarUint(sink, uint64(len(this.Yi)))
	for _, v := range this.Yi {
		utils.EncodeVarUint(sink, uint64(v))
	}
	return nil
}

func (this *SplitCurve) Deserialization(source *common.ZeroCopySource) error {
	n, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadVarUint, deserialize Yi length error: %v", err)
	}
	yi := make([]uint32, 0)
	for i := 0; uint64(i) < n; i++ {
		k, err := utils.DecodeVarUint(source)
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

func (this *TransferPenaltyParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
}

func (this *TransferPenaltyParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
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

func (this *WithdrawOngParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Address[:])
}

func (this *WithdrawOngParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
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

func (this *ChangeMaxAuthorizationParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])

	utils.EncodeVarUint(sink, uint64(this.MaxAuthorize))
}

func (this *ChangeMaxAuthorizationParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	maxAuthorize, err := utils.DecodeVarUint(source)
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

func (this *SetPeerCostParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
	utils.EncodeVarUint(sink, uint64(this.PeerCost))
	return nil
}

func (this *SetPeerCostParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	peerCost, err := utils.DecodeVarUint(source)
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

type SetFeePercentageParam struct {
	PeerPubkey string
	Address    common.Address
	PeerCost   uint32
	StakeCost  uint32
}

func (this *SetFeePercentageParam) Serialization(sink *common.ZeroCopySink) error {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
	utils.EncodeVarUint(sink, uint64(this.PeerCost))
	utils.EncodeVarUint(sink, uint64(this.StakeCost))
	return nil
}

func (this *SetFeePercentageParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	peerCost, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize peerCost error: %v", err)
	}
	stakeCost, err := utils.DecodeVarUint(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadBool, deserialize stakeCost error: %v", err)
	}
	if peerCost > math.MaxUint32 {
		return fmt.Errorf("peerCost larger than max of uint32")
	}
	if stakeCost > math.MaxUint32 {
		return fmt.Errorf("stakeCost larger than max of uint32")
	}
	this.PeerPubkey = peerPubkey
	this.Address = address
	this.PeerCost = uint32(peerCost)
	this.StakeCost = uint32(stakeCost)
	return nil
}

type WithdrawFeeParam struct {
	Address common.Address
}

func (this *WithdrawFeeParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Address[:])
}

func (this *WithdrawFeeParam) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
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

func (this *PromisePos) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	utils.EncodeVarUint(sink, this.PromisePos)
}

func (this *PromisePos) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	promisePos, err := utils.DecodeVarUint(source)
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

func (this *ChangeInitPosParam) Serialization(sink *common.ZeroCopySink) {
	sink.WriteString(this.PeerPubkey)
	sink.WriteVarBytes(this.Address[:])
	utils.EncodeVarUint(sink, uint64(this.Pos))
}

func (this *ChangeInitPosParam) Deserialization(source *common.ZeroCopySource) error {
	peerPubkey, err := utils.DecodeString(source)
	if err != nil {
		return fmt.Errorf("serialization.ReadString, deserialize peerPubkey error: %v", err)
	}
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.ReadAddress, deserialize address error: %v", err)
	}
	pos, err := utils.DecodeVarUint(source)
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

type GasAddress struct {
	Address common.Address
}

func (this *GasAddress) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.Address)
}

func (this *GasAddress) Deserialization(source *common.ZeroCopySource) error {
	address, err := utils.DecodeAddress(source)
	if err != nil {
		return fmt.Errorf("utils.DecodeAddress, deserialize address error: %v", err)
	}
	this.Address = address
	return nil
}
