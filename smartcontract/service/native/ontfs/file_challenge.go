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

package ontfs

import (
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type Challenge struct {
	FileHash        []byte
	FileOwner       common.Address
	NodeAddr        common.Address
	ChallengeHeight uint64
	Reward          uint64
	ExpiredTime     uint64
	State           uint64
}

type ChallengeList struct {
	Challenges []Challenge
}

func (this *Challenge) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.FileHash)
	utils.EncodeAddress(sink, this.FileOwner)
	utils.EncodeAddress(sink, this.NodeAddr)
	utils.EncodeVarUint(sink, this.ChallengeHeight)
	utils.EncodeVarUint(sink, this.Reward)
	utils.EncodeVarUint(sink, this.ExpiredTime)
	utils.EncodeVarUint(sink, this.State)
}

func (this *Challenge) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.FileHash, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	this.FileOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.ChallengeHeight, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Reward, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ExpiredTime, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.State, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *ChallengeList) Serialization(sink *common.ZeroCopySink) {
	challengeCount := uint64(len(this.Challenges))
	utils.EncodeVarUint(sink, challengeCount)

	for _, challenge := range this.Challenges {
		sinkTmp := common.NewZeroCopySink(nil)
		challenge.Serialization(sinkTmp)
		sink.WriteVarBytes(sinkTmp.Bytes())
	}
}

func (this *ChallengeList) Deserialization(source *common.ZeroCopySource) error {
	challengeCount, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	if 0 == challengeCount {
		return nil
	}

	for i := uint64(0); i < challengeCount; i++ {
		var challenge Challenge
		challengeTmp, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}
		src := common.NewZeroCopySource(challengeTmp)
		if err = challenge.Deserialization(src); err != nil {
			return err
		}
		this.Challenges = append(this.Challenges, challenge)
	}
	return nil
}

func addChallenge(native *native.NativeService, fileChallenge *Challenge) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileChallengeKey := GenChallengeKey(contract, fileChallenge.NodeAddr, fileChallenge.FileHash)

	sink := common.NewZeroCopySink(nil)
	fileChallenge.Serialization(sink)

	utils.PutBytes(native, fileChallengeKey, sink.Bytes())
}

func getChallenge(native *native.NativeService, nodeAddr common.Address, fileHash []byte) *Challenge {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileChallengeKey := GenChallengeKey(contract, nodeAddr, fileHash)

	item, err := utils.GetStorageItem(native, fileChallengeKey)
	if err != nil || item == nil || item.Value == nil {
		return nil
	}

	var challenge Challenge
	source := common.NewZeroCopySource(item.Value)
	if err := challenge.Deserialization(source); err != nil {
		return nil
	}
	if challenge.ExpiredTime < uint64(native.Time) && challenge.State == NoReplyAndValid {
		challenge.State = NoReplyAndExpire
	}
	return &challenge
}

func delChallenge(native *native.NativeService, nodeAddr common.Address, fileHash []byte) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	fileChallengeKey := GenChallengeKey(contract, nodeAddr, fileHash)
	native.CacheDB.Delete(fileChallengeKey)
}

func getChallengeFileHashList(native *native.NativeService, nodeAddr common.Address) []FileHash {
	contract := native.ContextRef.CurrentContext().ContractAddress

	challengePrefix := GenChallengePrefix(contract, nodeAddr)
	challengePrefixLen := len(challengePrefix)

	var fileHashList []FileHash

	iter := native.CacheDB.NewIterator(challengePrefix[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		fileHash := FileHash{
			FHash: make([]byte, len(key[challengePrefixLen:])),
		}
		copy(fileHash.FHash, key[challengePrefixLen:])
		fileHashList = append(fileHashList, fileHash)
	}
	iter.Release()

	return fileHashList
}

func getNodeChallengeList(native *native.NativeService, nodeAddr common.Address) *ChallengeList {
	fileHashList := getChallengeFileHashList(native, nodeAddr)
	if fileHashList == nil {
		return nil
	}

	var challengeList ChallengeList

	for _, fileHash := range fileHashList {
		challenge := getChallenge(native, nodeAddr, fileHash.FHash)
		if challenge == nil {
			fmt.Printf("[APP SDK] getNodeChallengeList getChallenge(%v)(%v) error", nodeAddr, fileHash.FHash)
			continue
		}
		challengeList.Challenges = append(challengeList.Challenges, *challenge)
	}

	return &challengeList
}

func getFileChallengeList(native *native.NativeService, pdpRecordList *PdpRecordList) *ChallengeList {
	var challengeList ChallengeList

	for _, pdpRecord := range pdpRecordList.PdpRecords {
		challenge := getChallenge(native, pdpRecord.NodeAddr, pdpRecord.FileHash)
		if challenge == nil {
			continue
		}
		challengeList.Challenges = append(challengeList.Challenges, *challenge)
	}

	return &challengeList
}
