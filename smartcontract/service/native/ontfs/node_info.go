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
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type FsNodeInfo struct {
	Pledge         uint64
	Profit         uint64
	Volume         uint64
	RestVol        uint64
	ServiceTime    uint64
	MinPdpInterval uint64
	NodeAddr       common.Address
	NodeNetAddr    []byte
}

type FsNodeInfoList struct {
	NodesInfo []FsNodeInfo
}

func (this *FsNodeInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeVarUint(sink, this.Pledge)
	utils.EncodeVarUint(sink, this.Profit)
	utils.EncodeVarUint(sink, this.Volume)
	utils.EncodeVarUint(sink, this.RestVol)
	utils.EncodeVarUint(sink, this.ServiceTime)
	utils.EncodeVarUint(sink, this.MinPdpInterval)
	utils.EncodeAddress(sink, this.NodeAddr)
	sink.WriteVarBytes(this.NodeNetAddr)
}

func (this *FsNodeInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.Pledge, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Profit, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.Volume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RestVol, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ServiceTime, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.MinPdpInterval, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.NodeAddr, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.NodeNetAddr, err = DecodeVarBytes(source)
	if err != nil {
		return err
	}
	return nil
}

func (this *FsNodeInfoList) Serialization(sink *common.ZeroCopySink) {
	fileCount := uint64(len(this.NodesInfo))
	utils.EncodeVarUint(sink, fileCount)
	if fileCount == 0 {
		return
	}

	for _, nodeInfo := range this.NodesInfo {
		sinkTmp := common.NewZeroCopySink(nil)
		nodeInfo.Serialization(sinkTmp)
		sink.WriteVarBytes(sinkTmp.Bytes())
	}
}

func (this *FsNodeInfoList) Deserialization(source *common.ZeroCopySource) error {
	fileCount, err := utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	if fileCount == 0 {
		return nil
	}

	for i := uint64(0); i < fileCount; i++ {
		nodeInfoTmp, err := DecodeVarBytes(source)
		if err != nil {
			return err
		}
		var nodeInfo FsNodeInfo
		src := common.NewZeroCopySource(nodeInfoTmp)
		if err = nodeInfo.Deserialization(src); err != nil {
			return err
		}
		this.NodesInfo = append(this.NodesInfo, nodeInfo)
	}
	return nil
}

func addNodeInfo(native *native.NativeService, nodeInfo *FsNodeInfo) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	nodeInfoKey := GenFsNodeInfoKey(contract, nodeInfo.NodeAddr)

	sink := common.NewZeroCopySink(nil)
	nodeInfo.Serialization(sink)

	utils.PutBytes(native, nodeInfoKey, sink.Bytes())
}

func delNodeInfo(native *native.NativeService, nodeAddr common.Address) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	nodeInfoKey := GenFsNodeInfoKey(contract, nodeAddr)
	native.CacheDB.Delete(nodeInfoKey)
}

func nodeInfoExist(native *native.NativeService, nodeAddr common.Address) bool {
	contract := native.ContextRef.CurrentContext().ContractAddress
	nodeInfoKey := GenFsNodeInfoKey(contract, nodeAddr)

	item, err := utils.GetStorageItem(native, nodeInfoKey)
	if err != nil || item == nil || item.Value == nil {
		return false
	}
	return true
}

func getNodeInfo(native *native.NativeService, nodeAddr common.Address) *FsNodeInfo {
	nodeRawInfo := getNodeRawInfo(native, nodeAddr)
	if nodeRawInfo == nil {
		return nil
	}
	var fsNodeInfo FsNodeInfo
	source := common.NewZeroCopySource(nodeRawInfo)
	if err := fsNodeInfo.Deserialization(source); err != nil {
		return nil
	}
	return &fsNodeInfo
}

func getNodeRawInfo(native *native.NativeService, nodeAddr common.Address) []byte {
	contract := native.ContextRef.CurrentContext().ContractAddress
	nodeInfoKey := GenFsNodeInfoKey(contract, nodeAddr)

	item, err := utils.GetStorageItem(native, nodeInfoKey)
	if err != nil || item == nil || item.Value == nil {
		return nil
	}
	return item.Value
}

func getNodeAddrList(native *native.NativeService) []common.Address {
	contract := native.ContextRef.CurrentContext().ContractAddress

	nodeInfoPrefix := GenFsNodeInfoPrefix(contract)
	nodeInfoPrefixLen := len(nodeInfoPrefix)

	var fsNodeAddrList []common.Address

	iter := native.CacheDB.NewIterator(nodeInfoPrefix[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()

		nodeAddr, err := common.AddressParseFromBytes(key[nodeInfoPrefixLen:])
		if err != nil {
			log.Errorf("getNodeAddrList AddressParseFromBytes error: ", err.Error())
			continue
		}
		fsNodeAddrList = append(fsNodeAddrList, nodeAddr)
	}
	iter.Release()

	return fsNodeAddrList
}

func getNodeInfoList(native *native.NativeService) map[common.Address]*FsNodeInfo {
	contract := native.ContextRef.CurrentContext().ContractAddress

	nodeInfoPrefix := GenFsNodeInfoPrefix(contract)
	nodeInfoPrefixLen := len(nodeInfoPrefix)

	fsNodeInfoList := make(map[common.Address]*FsNodeInfo)
	iter := native.CacheDB.NewIterator(nodeInfoPrefix[:])
	for has := iter.First(); has; has = iter.Next() {
		key := iter.Key()
		item, err := utils.GetStorageItem(native, iter.Key())
		if err != nil || item == nil || item.Value == nil {
			log.Error("getNodeInfoList GetStorageItem ", err)
			continue
		}

		nodeAddr, err := common.AddressParseFromBytes(key[nodeInfoPrefixLen:])
		if err != nil {
			log.Errorf("getNodeInfoList AddressParseFromBytes error: ", err.Error())
			continue
		}

		var fsNodeInfo FsNodeInfo
		source := common.NewZeroCopySource(item.Value)
		if err = fsNodeInfo.Deserialization(source); err != nil {
			log.Errorf("getNodeInfoList Deserialization error: ", err.Error())
			continue
		}

		fsNodeInfoList[nodeAddr] = &fsNodeInfo
	}
	iter.Release()

	return fsNodeInfoList
}
