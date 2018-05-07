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

package native

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"math/big"

	"github.com/ontio/ontology/common"
	vbftconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/genesis"
	cstates "github.com/ontio/ontology/core/states"
	scommon "github.com/ontio/ontology/core/store/common"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native/states"
)

var (
	ADDRESS_HEIGHT    = []byte("addressHeight")
	TRANSFER_NAME     = "transfer"
	TOTAL_SUPPLY_NAME = []byte("totalSupply")
)

func getAddressHeightKey(contract, address common.Address) []byte {
	temp := append(ADDRESS_HEIGHT, address[:]...)
	return append(contract[:], temp...)
}

func getHeightStorageItem(height uint32) *cstates.StorageItem {
	return &cstates.StorageItem{Value: big.NewInt(int64(height)).Bytes()}
}

func getAmountStorageItem(value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: value.Bytes()}
}

func getToAmountStorageItem(toBalance, value *big.Int) *cstates.StorageItem {
	return &cstates.StorageItem{Value: new(big.Int).Add(toBalance, value).Bytes()}
}

func getTotalSupplyKey(contract common.Address) []byte {
	return append(contract[:], TOTAL_SUPPLY_NAME...)
}

func getTransferKey(contract, from common.Address) []byte {
	return append(contract[:], from[:]...)
}

func getApproveKey(contract common.Address, state *states.State) []byte {
	temp := append(contract[:], state.From[:]...)
	return append(temp, state.To[:]...)
}

func getTransferFromKey(contract common.Address, state *states.TransferFrom) []byte {
	temp := append(contract[:], state.From[:]...)
	return append(temp, state.Sender[:]...)
}

func isTransferValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Transfer amount invalid!")
	}

	if native.ContextRef.CheckWitness(state.From) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func transfer(native *NativeService, contract common.Address, state *states.State) (*big.Int, *big.Int, error) {
	if err := isTransferValid(native, state); err != nil {
		return nil, nil, err
	}

	fromBalance, err := fromTransfer(native, getTransferKey(contract, state.From), state.Value)
	if err != nil {
		return nil, nil, err
	}

	toBalance, err := toTransfer(native, getTransferKey(contract, state.To), state.Value)
	if err != nil {
		return nil, nil, err
	}
	return fromBalance, toBalance, nil
}

func transferFrom(native *NativeService, currentContract common.Address, state *states.TransferFrom) error {
	if err := isTransferFromValid(native, state); err != nil {
		return err
	}

	if err := fromApprove(native, getTransferFromKey(currentContract, state), state.Value); err != nil {
		return err
	}

	if _, err := fromTransfer(native, getTransferKey(currentContract, state.From), state.Value); err != nil {
		return err
	}

	if _, err := toTransfer(native, getTransferKey(currentContract, state.To), state.Value); err != nil {
		return err
	}
	return nil
}

func isTransferFromValid(native *NativeService, state *states.TransferFrom) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("TransferFrom amount invalid!")
	}

	if native.ContextRef.CheckWitness(state.Sender) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func isApproveValid(native *NativeService, state *states.State) error {
	if state.Value.Sign() < 0 {
		return errors.NewErr("Approve amount invalid!")
	}
	if native.ContextRef.CheckWitness(state.From) == false {
		return errors.NewErr("[Sender] Authentication failed!")
	}
	return nil
}

func fromApprove(native *NativeService, fromApproveKey []byte, value *big.Int) error {
	approveValue, err := getStorageBigInt(native, fromApproveKey)
	if err != nil {
		return err
	}
	approveBalance := new(big.Int).Sub(approveValue, value)
	sign := approveBalance.Sign()
	if sign < 0 {
		return fmt.Errorf("[TransferFrom] approve balance insufficient! have %d, got %d", approveValue.Int64(), value.Int64())
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromApproveKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromApproveKey, getAmountStorageItem(approveBalance))
	}
	return nil
}

func fromTransfer(native *NativeService, fromKey []byte, value *big.Int) (*big.Int, error) {
	fromBalance, err := getStorageBigInt(native, fromKey)
	if err != nil {
		return nil, err
	}
	balance := new(big.Int).Sub(fromBalance, value)
	sign := balance.Sign()
	if sign < 0 {
		return nil, errors.NewErr("[Transfer] balance insufficient!")
	} else if sign == 0 {
		native.CloneCache.Delete(scommon.ST_STORAGE, fromKey)
	} else {
		native.CloneCache.Add(scommon.ST_STORAGE, fromKey, getAmountStorageItem(balance))
	}
	//native.CloneCache.Add(scommon.ST_STORAGE, fromKey, getAmountStorageItem(balance))
	return fromBalance, nil
}

func toTransfer(native *NativeService, toKey []byte, value *big.Int) (*big.Int, error) {
	toBalance, err := getStorageBigInt(native, toKey)
	if err != nil {
		return nil, err
	}
	native.CloneCache.Add(scommon.ST_STORAGE, toKey, getToAmountStorageItem(toBalance, value))
	return toBalance, nil
}

func getStartHeight(native *NativeService, contract, from common.Address) (uint32, error) {
	startHeight, err := getStorageBigInt(native, getAddressHeightKey(contract, from))
	if err != nil {
		return 0, err
	}
	return uint32(startHeight.Int64()), nil
}

func getStorageBigInt(native *NativeService, key []byte) (*big.Int, error) {
	balance, err := native.CloneCache.Get(scommon.ST_STORAGE, key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] storage error!")
	}
	if balance == nil {
		return big.NewInt(0), nil
	}
	item, ok := balance.(*cstates.StorageItem)
	if !ok {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] get amount error!")
	}
	return new(big.Int).SetBytes(item.Value), nil
}

func addNotifications(native *NativeService, contract common.Address, state *states.State) {
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			TxHash:          native.Tx.Hash(),
			ContractAddress: contract,
			States:          []interface{}{TRANSFER_NAME, state.From.ToBase58(), state.To.ToBase58(), state.Value},
		})
}

func concatKey(contract common.Address, args ...[]byte) []byte {
	temp := contract[:]
	for _, arg := range args {
		temp = append(temp, arg...)
	}
	return temp
}

func validateOwner(native *NativeService, address string) error {
	addrBytes, err := hex.DecodeString(address)
	if err != nil {
		return errors.NewErr("[validateOwner] Decode address hex string to bytes failed!")
	}
	addr, err := common.AddressParseFromBytes(addrBytes)
	if err != nil {
		return errors.NewErr("[validateOwner] Decode bytes to address failed!")
	}
	if native.ContextRef.CheckWitness(addr) == false {
		return errors.NewErr("[validateOwner] Authentication failed!")
	}
	return nil
}

func getGovernanceView(native *NativeService, contract common.Address) (*big.Int, error) {
	governanceViewBytes, err := native.CloneCache.Get(scommon.ST_STORAGE, concatKey(contract, []byte(GOVERNANCE_VIEW)))
	if err != nil {
		return new(big.Int), errors.NewDetailErr(err, errors.ErrNoCode, "[getGovernanceView] Get governanceViewBytes error!")
	}
	governanceView := new(states.GovernanceView)
	if governanceViewBytes == nil {
		return new(big.Int), errors.NewDetailErr(err, errors.ErrNoCode, "[getGovernanceView] Get nil governanceViewBytes!")
	} else {
		governanceViewStore, _ := governanceViewBytes.(*cstates.StorageItem)
		err = json.Unmarshal(governanceViewStore.Value, governanceView)
		if err != nil {
			return new(big.Int), errors.NewDetailErr(err, errors.ErrNoCode, "[getGovernanceView] Unmarshal governanceView error!")
		}
	}
	return governanceView.View, nil
}

func addCommonEvent(native *NativeService, contract common.Address, name string, params interface{}) {
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			TxHash:          native.Tx.Hash(),
			ContractAddress: contract,
			States:          []interface{}{name, params},
		})
}

func appCallTransferOng(native *NativeService, from common.Address, to common.Address, amount *big.Int) error {
	buf := bytes.NewBuffer(nil)
	var sts []*states.State
	sts = append(sts, &states.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := &states.Transfers{
		States: sts,
	}
	err := transfers.Serialize(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOng] transfers.Serialize error!")
	}

	if _, err := native.ContextRef.AppCall(genesis.OngContractAddress, "transfer", []byte{}, buf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOng] appCall error!")
	}
	return nil
}

func appCallTransferOnt(native *NativeService, from common.Address, to common.Address, amount *big.Int) error {
	buf := bytes.NewBuffer(nil)
	var sts []*states.State
	sts = append(sts, &states.State{
		From:  from,
		To:    to,
		Value: amount,
	})
	transfers := &states.Transfers{
		States: sts,
	}
	err := transfers.Serialize(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] transfers.Serialize error!")
	}

	if _, err := native.ContextRef.AppCall(genesis.OntContractAddress, "transfer", []byte{}, buf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallTransferOnt] appCall error!")
	}
	return nil
}

func appCallApproveOng(native *NativeService, from common.Address, to common.Address, amount *big.Int) error {
	buf := bytes.NewBuffer(nil)
	sts := &states.State{
		From:  from,
		To:    to,
		Value: amount,
	}
	err := sts.Serialize(buf)
	if err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallApproveOng] transfers.Serialize error!")
	}

	if _, err := native.ContextRef.AppCall(genesis.OngContractAddress, "approve", []byte{}, buf.Bytes()); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[appCallApproveOng] appCall error!")
	}
	return nil
}

func Shufflehash(txid common.Uint256, ts uint32, id []byte, idx int) (uint64, error) {
	data, err := json.Marshal(struct {
		Txid           common.Uint256 `json:"txid"`
		BlockTimestamp uint32         `json:"block_timestamp"`
		NodeID         []byte         `json:"node_id"`
		Index          int            `json:"index"`
	}{txid, ts, id, idx})
	if err != nil {
		return 0, err
	}

	hash := fnv.New64a()
	hash.Write(data)
	return hash.Sum64(), nil
}

func calDposTable(native *NativeService, config *states.Configuration,
	peers []*states.PeerStakeInfo) ([]uint32, map[uint32]*vbftconfig.PeerConfig, error) {
	// get stake sum of top-k peers
	var sum uint64
	for i := 0; i < int(config.K); i++ {
		sum += peers[i].Stake
	}

	// calculate peer ranks
	scale := config.L/config.K - 1
	if scale <= 0 {
		return nil, nil, errors.NewErr("[calDposTable] L is equal or less than K!")
	}

	peerRanks := make([]uint64, 0)
	for i := 0; i < int(config.K); i++ {
		if peers[i].Stake == 0 {
			return nil, nil, errors.NewErr(fmt.Sprintf("[calDposTable] peers rank %d, has zero stake!", i))
		}
		s := uint64(math.Ceil(float64(peers[i].Stake) * float64(scale) * float64(config.K) / float64(sum)))
		peerRanks = append(peerRanks, s)
	}

	// calculate pos table
	chainPeers := make(map[uint32]*vbftconfig.PeerConfig, 0)
	posTable := make([]uint32, 0)
	for i := 0; i < int(config.K); i++ {
		nodeId, err := vbftconfig.StringID(peers[i].PeerPubkey)
		if err != nil {
			return nil, nil, errors.NewDetailErr(err, errors.ErrNoCode,
				fmt.Sprintf("[calDposTable] Failed to format NodeID, index: %d: %s", peers[i].Index, err))
		}
		chainPeers[peers[i].Index] = &vbftconfig.PeerConfig{
			Index: peers[i].Index,
			ID:    nodeId,
		}
		for j := uint64(0); j < peerRanks[i]; j++ {
			posTable = append(posTable, peers[i].Index)
		}
	}

	// shuffle
	for i := len(posTable) - 1; i > 0; i-- {
		h, err := Shufflehash(native.Tx.Hash(), native.Height, chainPeers[posTable[i]].ID.Bytes(), i)
		if err != nil {
			return nil, nil, errors.NewDetailErr(err, errors.ErrNoCode, "[calDposTable] Failed to calculate hash value")
		}
		j := h % uint64(i)
		posTable[i], posTable[j] = posTable[j], posTable[i]
	}

	return posTable, chainPeers, nil
}
