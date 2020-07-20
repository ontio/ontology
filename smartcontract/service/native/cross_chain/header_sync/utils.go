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

package header_sync

import (
	"encoding/json"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/signature"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
	ccom "github.com/ontio/ontology/smartcontract/service/native/cross_chain/common"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func PutBlockHeader(native *native.NativeService, blockHeader *ccom.Header, header []byte) error {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(blockHeader.ChainID)
	if err != nil {
		return fmt.Errorf("chainIDBytes, GetUint64Bytes error: %v", err)
	}
	heightBytes, err := utils.GetUint32Bytes(blockHeader.Height)
	if err != nil {
		return fmt.Errorf("heightBytes, getUint32Bytes error: %v", err)
	}
	blockHash := blockHeader.Hash()
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(BLOCK_HEADER), chainIDBytes, blockHash.ToArray()),
		cstates.GenRawStorageItem(header))
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(HEADER_INDEX), chainIDBytes, heightBytes),
		cstates.GenRawStorageItem(blockHash.ToArray()))
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CURRENT_HEIGHT), chainIDBytes), cstates.GenRawStorageItem(heightBytes))
	notifyPutHeader(native, blockHeader.ChainID, blockHeader.Height, blockHash.ToHexString())
	return nil
}

func GetHeaderByHeight(native *native.NativeService, chainID uint64, height uint32) (*ccom.Header, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, GetUint64Bytes error: %v", err)
	}
	heightBytes, err := utils.GetUint32Bytes(height)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, getUint32Bytes error: %v", err)
	}
	blockHashStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(HEADER_INDEX), chainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, get blockHashStore error: %v", err)
	}
	if blockHashStore == nil {
		return nil, nil
	}
	blockHashBytes, err := cstates.GetValueFromRawStorageItem(blockHashStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize blockHashBytes from raw storage item err:%v", err)
	}
	header := new(ccom.Header)
	headerStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(BLOCK_HEADER), chainIDBytes, blockHashBytes))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, get headerStore error: %v", err)
	}
	if headerStore == nil {
		return nil, nil
	}
	headerBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize headerBytes from raw storage item err:%v", err)
	}
	if err := header.Deserialization(common.NewZeroCopySource(headerBytes)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHeight, deserialize header error: %v", err)
	}
	return header, nil
}

func GetHeaderByHash(native *native.NativeService, chainID uint64, hash common.Uint256) (*ccom.Header, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, getUint32Bytes error: %v", err)
	}
	header := new(ccom.Header)
	headerStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(BLOCK_HEADER), chainIDBytes, hash.ToArray()))
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, get headerStore error: %v", err)
	}
	if headerStore == nil {
		return nil, fmt.Errorf("GetHeaderByHash, can not find any records")
	}
	headerBytes, err := cstates.GetValueFromRawStorageItem(headerStore)
	if err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize from raw storage item err:%v", err)
	}
	if err := header.Deserialization(common.NewZeroCopySource(headerBytes)); err != nil {
		return nil, fmt.Errorf("GetHeaderByHash, deserialize header error: %v", err)
	}
	return header, nil
}

//verify header of any height
//find key height and get consensus peer first, then check the sign
func VerifyHeader(native *native.NativeService, header *ccom.Header) error {
	height := header.Height
	//search consensus peer
	keyHeight, err := findKeyHeight(native, height, header.ChainID)
	if err != nil {
		return fmt.Errorf("verifyHeader, findKeyHeight error:%v", err)
	}

	consensusPeer, err := getConsensusPeersByHeight(native, header.ChainID, keyHeight)
	if err != nil {
		return fmt.Errorf("verifyHeader, get ConsensusPeer error:%v", err)
	}
	if len(header.Bookkeepers)*3 < len(consensusPeer.PeerMap)*2 {
		return fmt.Errorf("verifyHeader, header Bookkeepers num %d must more than 2/3 consensus node num %d", len(header.Bookkeepers), len(consensusPeer.PeerMap))
	}
	for _, bookkeeper := range header.Bookkeepers {
		pubkey := vconfig.PubkeyID(bookkeeper)
		_, present := consensusPeer.PeerMap[pubkey]
		if !present {
			return fmt.Errorf("verifyHeader, invalid pubkey error:%v", pubkey)
		}
	}
	hash := header.Hash()
	err = signature.VerifyMultiSignature(hash[:], header.Bookkeepers, len(header.Bookkeepers), header.SigData)
	if err != nil {
		return fmt.Errorf("verifyHeader, VerifyMultiSignature error:%s, heigh:%d", err, header.Height)
	}
	return nil
}

func GetKeyHeights(native *native.NativeService, chainID uint64) (*KeyHeights, error) {
	contract := utils.HeaderSyncContractAddress
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("chainIDBytes, GetUint64Bytes error: %v", err)
	}
	value, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_HEIGHTS), chainIDBytes))
	if err != nil {
		return nil, fmt.Errorf("GetKeyHeights, get keyHeights value error: %v", err)
	}
	keyHeights := &KeyHeights{
		HeightList: make([]uint32, 0),
	}
	if value != nil {
		keyHeightsBytes, err := cstates.GetValueFromRawStorageItem(value)
		if err != nil {
			return nil, fmt.Errorf("GetKeyHeights, deserialize from raw storage item err:%v", err)
		}
		err = keyHeights.Deserialization(common.NewZeroCopySource(keyHeightsBytes))
		if err != nil {
			return nil, fmt.Errorf("GetKeyHeights, deserialize keyHeights err:%v", err)
		}
	}
	return keyHeights, nil
}

func putKeyHeights(native *native.NativeService, chainID uint64, keyHeights *KeyHeights) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	keyHeights.Serialization(sink)
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return fmt.Errorf("GetUint64Bytes error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_HEIGHTS), chainIDBytes), cstates.GenRawStorageItem(sink.Bytes()))
	return nil
}

func getConsensusPeersByHeight(native *native.NativeService, chainID uint64, height uint32) (*ConsensusPeers, error) {
	contract := utils.HeaderSyncContractAddress
	heightBytes, err := utils.GetUint32Bytes(height)
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, getUint32Bytes error: %v", err)
	}
	chainIDBytes, err := utils.GetUint64Bytes(chainID)
	if err != nil {
		return nil, fmt.Errorf("GetUint64Bytes error: %v", err)
	}
	consensusPeerStore, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(CONSENSUS_PEER), chainIDBytes, heightBytes))
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, get consensusPeerStore error: %v", err)
	}
	consensusPeers := &ConsensusPeers{
		ChainID: chainID,
		Height:  height,
		PeerMap: make(map[string]*Peer),
	}
	if consensusPeerStore == nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, can not find any record")
	}
	consensusPeerBytes, err := cstates.GetValueFromRawStorageItem(consensusPeerStore)
	if err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize from raw storage item err:%v", err)
	}
	if err := consensusPeers.Deserialization(common.NewZeroCopySource(consensusPeerBytes)); err != nil {
		return nil, fmt.Errorf("getConsensusPeerByHeight, deserialize consensusPeer error: %v", err)
	}
	return consensusPeers, nil
}

func putConsensusPeers(native *native.NativeService, consensusPeers *ConsensusPeers) error {
	contract := utils.HeaderSyncContractAddress
	sink := common.NewZeroCopySink(nil)
	consensusPeers.Serialization(sink)
	chainIDBytes, err := utils.GetUint64Bytes(consensusPeers.ChainID)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, GetUint64Bytes error: %v", err)
	}
	heightBytes, err := utils.GetUint32Bytes(consensusPeers.Height)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, getUint32Bytes 1 error: %v", err)
	}
	native.CacheDB.Put(utils.ConcatKey(contract, []byte(CONSENSUS_PEER), chainIDBytes, heightBytes), cstates.GenRawStorageItem(sink.Bytes()))

	//update key heights
	keyHeights, err := GetKeyHeights(native, consensusPeers.ChainID)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, GetKeyHeights error: %v", err)
	}
	keyHeights.HeightList = append(keyHeights.HeightList, consensusPeers.Height)
	err = putKeyHeights(native, consensusPeers.ChainID, keyHeights)
	if err != nil {
		return fmt.Errorf("putConsensusPeer, putKeyHeights error: %v", err)
	}
	return nil
}

func UpdateConsensusPeer(native *native.NativeService, header *ccom.Header) error {
	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(header.ConsensusPayload, blkInfo); err != nil {
		return fmt.Errorf("updateConsensusPeer, unmarshal blockInfo error: %s", err)
	}
	if blkInfo.NewChainConfig != nil {
		consensusPeers := &ConsensusPeers{
			ChainID: header.ChainID,
			Height:  header.Height,
			PeerMap: make(map[string]*Peer),
		}
		for _, p := range blkInfo.NewChainConfig.Peers {
			consensusPeers.PeerMap[p.ID] = &Peer{Index: p.Index, PeerPubkey: p.ID}
		}
		err := putConsensusPeers(native, consensusPeers)
		if err != nil {
			return fmt.Errorf("updateConsensusPeer, put ConsensusPeer eerror: %s", err)
		}
	}
	return nil
}

func findKeyHeight(native *native.NativeService, height uint32, chainID uint64) (uint32, error) {
	keyHeights, err := GetKeyHeights(native, chainID)
	if err != nil {
		return 0, fmt.Errorf("findKeyHeight, GetKeyHeights error: %v", err)
	}
	for _, v := range keyHeights.HeightList {
		if height > v {
			return v, nil
		}
	}
	return 0, fmt.Errorf("findKeyHeight, can not find key height with height %d", height)
}

func notifyPutHeader(native *native.NativeService, chainID uint64, height uint32, blockHash string) {
	if !config.DefConfig.Common.EnableEventLog {
		return
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			ContractAddress: utils.HeaderSyncContractAddress,
			States:          []interface{}{chainID, height, blockHash, native.Height},
		})
}

func ProcessHeader(native *native.NativeService, header *ccom.Header, h []byte) error {
	err := VerifyHeader(native, header)
	if err != nil {
		return fmt.Errorf("ProcessHeader, verifyHeader error: %v", err)
	}
	err = PutBlockHeader(native, header, h)
	if err != nil {
		return fmt.Errorf("ProcessHeader, put BlockHeader error: %v", err)
	}
	err = UpdateConsensusPeer(native, header)
	if err != nil {
		return fmt.Errorf("ProcessHeader, update ConsensusPeer error: %v", err)
	}
	return nil
}
