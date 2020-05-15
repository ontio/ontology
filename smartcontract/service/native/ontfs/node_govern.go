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
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func FsNodeRegister(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var nodeInfo FsNodeInfo
	source := common.NewZeroCopySource(native.Input)
	if err := nodeInfo.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeRegister FsNodeInfo Deserialization error!")
	}

	if !native.ContextRef.CheckWitness(nodeInfo.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeRegister CheckRegister failed!")
	}

	if nodeInfoExist(native, nodeInfo.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeRegister Node have registered!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil || globalParam == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeRegister getGlobalParam error!")
	}

	if nodeInfo.ServiceTime < uint64(native.Time) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeRegister ServiceTime error!")
	}

	if nodeInfo.Volume < globalParam.NodeMinVolume {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeRegister Volume < MinVolume!")
	}

	nodePledge := globalParam.NodePerKbPledge * nodeInfo.Volume
	err = appCallTransfer(native, utils.OngContractAddress, nodeInfo.NodeAddr, contract, nodePledge)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeRegister appCallTransfer, transfer error!")
	}

	nodeInfo.Profit = 0
	nodeInfo.Pledge = nodePledge
	nodeInfo.RestVol = nodeInfo.Volume

	addNodeInfo(native, &nodeInfo)
	return utils.BYTE_TRUE, nil
}

func FsNodeQuery(native *native.NativeService) ([]byte, error) {
	source := common.NewZeroCopySource(native.Input)
	nodeAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return EncRet(false, []byte("[Node Govern] FsNodeQuery DecodeAddress error!")), nil
	}

	nodeRawInfo := getNodeRawInfo(native, nodeAddr)
	if nodeRawInfo == nil {
		return EncRet(false, []byte("[Node Govern] FsNodeQuery getNodeRawInfo error!")), nil
	}
	return EncRet(true, nodeRawInfo), nil
}

func FsNodeUpdate(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	var newNodeInfo FsNodeInfo
	source := common.NewZeroCopySource(native.Input)
	if err := newNodeInfo.Deserialization(source); err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate FsNodeInfo Deserialization error!")
	}

	if !native.ContextRef.CheckWitness(newNodeInfo.NodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate CheckNodeAddr failed!")
	}

	globalParam, err := getGlobalParam(native)
	if err != nil || globalParam == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate getGlobalParam error!")
	}

	if newNodeInfo.ServiceTime < uint64(native.Time) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate ServiceTime error!")
	}

	if newNodeInfo.Volume < globalParam.NodeMinVolume {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate Volume < MinVolume!")
	}

	oldNodeInfo := getNodeInfo(native, newNodeInfo.NodeAddr)
	if oldNodeInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate getNodeInfo error!")
	}

	if newNodeInfo.Volume < oldNodeInfo.Volume-oldNodeInfo.RestVol {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate Volume < Have Used Volume!")
	}

	newNodePledge := globalParam.NodePerKbPledge * newNodeInfo.Volume
	if newNodeInfo.NodeAddr != oldNodeInfo.NodeAddr {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate FsNodeInfo nodeAddr changed!")
	}

	var state *ont.State
	if newNodePledge < oldNodeInfo.Pledge {
		state = &ont.State{From: contract, To: oldNodeInfo.NodeAddr, Value: oldNodeInfo.Pledge - newNodePledge}
	} else if newNodePledge > oldNodeInfo.Pledge {
		state = &ont.State{From: newNodeInfo.NodeAddr, To: contract, Value: newNodePledge - oldNodeInfo.Pledge}
	}
	if newNodePledge != oldNodeInfo.Pledge {
		err = appCallTransfer(native, utils.OngContractAddress, state.From, state.To, state.Value)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeUpdate appCallTransfer, transfer error!")
		}
	}

	newNodeInfo.Pledge = newNodePledge
	newNodeInfo.Profit = oldNodeInfo.Profit
	newNodeInfo.RestVol = oldNodeInfo.RestVol + newNodeInfo.Volume - oldNodeInfo.Volume

	addNodeInfo(native, &newNodeInfo)
	return utils.BYTE_TRUE, nil
}

func FsNodeCancel(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	source := common.NewZeroCopySource(native.Input)
	nodeAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeCancel DecodeAddress error!")
	}

	if !native.ContextRef.CheckWitness(nodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeCancel CheckNodeAddr failed!")
	}

	nodeInfo := getNodeInfo(native, nodeAddr)
	if nodeInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeCancel getFsNodeInfo error!")
	}

	if uint64(native.Time) < nodeInfo.ServiceTime {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeCancel ServiceTime not due!")
	}

	if nodeInfo.Pledge+nodeInfo.Profit > 0 {
		err = appCallTransfer(native, utils.OngContractAddress, contract, nodeInfo.NodeAddr, nodeInfo.Pledge+nodeInfo.Profit)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeCancel appCallTransfer,  transfer error!")
		}
	}

	delNodeInfo(native, nodeAddr)
	return utils.BYTE_TRUE, nil
}

func FsNodeWithDrawProfit(native *native.NativeService) ([]byte, error) {
	contract := native.ContextRef.CurrentContext().ContractAddress

	source := common.NewZeroCopySource(native.Input)
	nodeAddr, err := utils.DecodeAddress(source)
	if err != nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeWithDrawProfit DecodeAddress error!")
	}

	if !native.ContextRef.CheckWitness(nodeAddr) {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeWithDrawProfit CheckNodeAddr failed!")
	}

	nodeInfo := getNodeInfo(native, nodeAddr)
	if nodeInfo == nil {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeWithDrawProfit getFsNodeInfo error!")
	}

	if nodeInfo.Profit > 0 {
		err = appCallTransfer(native, utils.OngContractAddress, contract, nodeInfo.NodeAddr, nodeInfo.Profit)
		if err != nil {
			return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeWithDrawProfit appCallTransfer,  transfer error!")
		}
		nodeInfo.Profit = 0
	} else {
		return utils.BYTE_FALSE, errors.NewErr("[Node Govern] FsNodeWithDrawProfit profit = 0 error! ")
	}

	addNodeInfo(native, nodeInfo)
	return utils.BYTE_TRUE, nil
}
