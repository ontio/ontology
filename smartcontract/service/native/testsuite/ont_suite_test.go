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
package testsuite

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/laizy/bigint"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	_ "github.com/ontio/ontology/smartcontract/service/native/init"
	"github.com/ontio/ontology/smartcontract/service/native/ong"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/stretchr/testify/assert"
)

func setOntBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := ont.GenBalanceKey(utils.OntContractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func ontBalanceOf(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := ont.OntBalanceOf(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}
func ontBalanceOfV2(native *native.NativeService, addr common.Address) uint64 {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := ont.OntBalanceOfV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func setOngBalance(db *storage.CacheDB, addr common.Address, value uint64) {
	balanceKey := ont.GenBalanceKey(utils.OngContractAddress, addr)
	item := utils.GenUInt64StorageItem(value)
	db.Put(balanceKey, item.ToArray())
}

func ongBalanceOf(native *native.NativeService, addr common.Address) uint64 {
	native.ContextRef.CurrentContext().ContractAddress = utils.OngContractAddress
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := ong.OngBalanceOf(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func ongBalanceOfV2(native *native.NativeService, addr common.Address) bigint.Int {
	native.ContextRef.CurrentContext().ContractAddress = utils.OngContractAddress
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := ong.OngBalanceOfV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return bigint.New(val)
}

func ongAllowance(native *native.NativeService, from, to common.Address) uint64 {
	native.ContextRef.CurrentContext().ContractAddress = utils.OngContractAddress
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, from)
	utils.EncodeAddress(sink, to)
	native.Input = sink.Bytes()
	buf, _ := ong.OngAllowance(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func ongAllowanceV2(native *native.NativeService, from, to common.Address) bigint.Int {
	native.ContextRef.CurrentContext().ContractAddress = utils.OngContractAddress
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, from)
	utils.EncodeAddress(sink, to)
	native.Input = sink.Bytes()
	buf, _ := ong.OngAllowanceV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return bigint.New(val)
}

func ongTransferFromV2(native *native.NativeService, spender, from, to common.Address, amt states.NativeTokenBalance) error {
	native.ContextRef.CurrentContext().ContractAddress = utils.OngContractAddress
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &ont.TransferFromStateV2{Sender: spender, TransferStateV2: ont.TransferStateV2{From: from, To: to, Value: amt}}
	native.Input = common.SerializeToBytes(state)

	_, err := ong.OngTransferFromV2(native)
	return err
}

func ontTotalAllowance(native *native.NativeService, addr common.Address) int {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := ont.TotalAllowance(native)
	val := common.BigIntFromNeoBytes(buf)
	return int(val.Uint64())
}

func ontTotalAllowanceV2(native *native.NativeService, addr common.Address) uint64 {
	sink := common.NewZeroCopySink(nil)
	utils.EncodeAddress(sink, addr)
	native.Input = sink.Bytes()
	buf, _ := ont.TotalAllowanceV2(native)
	val := common.BigIntFromNeoBytes(buf)
	return val.Uint64()
}

func ontTransfer(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := ont.TransferState{from, to, value}
	native.Input = common.SerializeToBytes(&ont.TransferStates{States: []ont.TransferState{state}})
	_, err := ont.OntTransfer(native)
	return err
}

func ontTransferV2(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &ont.TransferStateV2{from, to, states.NativeTokenBalance{Balance: bigint.New(value)}}
	native.Input = common.SerializeToBytes(&ont.TransferStatesV2{States: []*ont.TransferStateV2{state}})
	_, err := ont.OntTransferV2(native)
	return err
}

func ontTransferFrom(native *native.NativeService, sender, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &ont.TransferFrom{sender, ont.TransferState{from, to, value}}
	native.Input = common.SerializeToBytes(state)
	_, err := ont.OntTransferFrom(native)
	return err
}

func ontTransferFromV2(native *native.NativeService, sender, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)
	state := &ont.TransferFromStateV2{sender, ont.TransferStateV2{from, to, states.NativeTokenBalance{Balance: bigint.New(value)}}}
	native.Input = common.SerializeToBytes(state)
	_, err := ont.OntTransferFromV2(native)
	return err
}

func ontApprove(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	native.Input = common.SerializeToBytes(&ont.TransferState{from, to, value})
	_, err := ont.OntApprove(native)
	return err
}

func ontApproveV2(native *native.NativeService, from, to common.Address, value uint64) error {
	native.Tx.SignedAddr = append(native.Tx.SignedAddr, from)

	native.Input = common.SerializeToBytes((&ont.TransferStateV2{from, to, states.NativeTokenBalance{Balance: bigint.New(value)}}))
	_, err := ont.OntApproveV2(native)
	return err
}

func unboundGovernanceOng(native *native.NativeService) error {
	_, err := ont.UnboundOngToGovernance(native)
	return err
}

func TestTransfer(t *testing.T) {
	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()

		setOntBalance(native.CacheDB, a, 10000)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		assert.Equal(t, ontBalanceOf(native, a), 10000)
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 0)

		assert.Nil(t, ontTransfer(native, a, b, 10))
		assert.Equal(t, ontBalanceOf(native, a), 9990)
		assert.Equal(t, ontBalanceOf(native, b), 10)

		assert.Nil(t, ontTransfer(native, b, c, 10))
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 10)

		return nil, nil
	})
}

func TestTotalAllowance(t *testing.T) {
	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		assert.Equal(t, ontBalanceOf(native, a), 10000)
		assert.Equal(t, ontBalanceOf(native, b), 0)
		assert.Equal(t, ontBalanceOf(native, c), 0)

		assert.Nil(t, ontApprove(native, a, b, 10))
		assert.Equal(t, ontTotalAllowance(native, a), 10)
		assert.Equal(t, ontTotalAllowance(native, b), 0)

		assert.Nil(t, ontApprove(native, a, c, 100))
		assert.Equal(t, ontTotalAllowance(native, a), 110)
		assert.Equal(t, ontTotalAllowance(native, c), 0)

		assert.Nil(t, ontTransferFrom(native, c, a, c, 100))
		return nil, nil
	})
}

func TestGovernanceUnbound(t *testing.T) {
	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1

		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		assert.Equal(t, ongAllowance(native, utils.OntContractAddress, testAddr), uint64(5000000000))

		return nil, nil
	})

	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceContractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		assert.Nil(t, unboundGovernanceOng(native))
		assert.EqualValues(t, ongBalanceOf(native, gov)+ongBalanceOf(native, testAddr), constants.ONG_TOTAL_SUPPLY)

		return nil, nil
	})

	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceContractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		assert.EqualValues(t, ongBalanceOf(native, gov)+ongBalanceOf(native, testAddr), constants.ONG_TOTAL_SUPPLY)

		return nil, nil
	})

	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceContractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 10000
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		native.Time = config.GetOntHolderUnboundDeadline() - 100
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, ontTransfer(native, testAddr, testAddr, 1))
		assert.EqualValues(t, ongBalanceOf(native, gov)+ongBalanceOf(native, testAddr), constants.ONG_TOTAL_SUPPLY)

		return nil, nil
	})
}

//************************ version 2 ***************************

func TestTransferV2(t *testing.T) {
	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)
		// default networkid is mainnet, need set ong balance for ont contract
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		assert.Equal(t, ontBalanceOfV2(native, a), uint64(10000*states.ScaleFactor))
		assert.Equal(t, ontBalanceOfV2(native, b), uint64(0))
		assert.Equal(t, ontBalanceOfV2(native, c), uint64(0))
		assert.Equal(t, ontBalanceOf(native, a), 10000)

		assert.Nil(t, ontTransferV2(native, a, b, 10*states.ScaleFactor))
		assert.Equal(t, ontBalanceOfV2(native, a), uint64(9990*states.ScaleFactor))
		assert.Equal(t, ontBalanceOfV2(native, b), uint64(10*states.ScaleFactor))
		assert.Equal(t, ontBalanceOf(native, a), 9990)
		assert.Equal(t, ontBalanceOf(native, b), 10)

		assert.Nil(t, ontTransferV2(native, b, c, 10*states.ScaleFactor))
		assert.Equal(t, ontBalanceOfV2(native, b), uint64(0))
		assert.Equal(t, ontBalanceOfV2(native, c), uint64(10*states.ScaleFactor))

		return nil, nil
	})
}

func TestTotalAllowanceV2(t *testing.T) {
	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		a, b, c := RandomAddress(), RandomAddress(), RandomAddress()
		setOntBalance(native.CacheDB, a, 10000)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		assert.Equal(t, ontBalanceOfV2(native, a), uint64(10000*states.ScaleFactor))
		assert.Equal(t, ontBalanceOfV2(native, b), uint64(0))
		assert.Equal(t, ontBalanceOfV2(native, c), uint64(0))

		assert.Nil(t, ontApproveV2(native, a, b, 10*states.ScaleFactor))
		assert.Equal(t, ontTotalAllowanceV2(native, a), uint64(10*states.ScaleFactor))
		assert.Equal(t, ontTotalAllowanceV2(native, b), uint64(0))

		assert.Nil(t, ontApproveV2(native, a, c, uint64(100*states.ScaleFactor)))
		assert.Equal(t, ontTotalAllowanceV2(native, a), uint64(110*states.ScaleFactor))
		assert.Equal(t, ontTotalAllowanceV2(native, c), uint64(0))
		fmt.Println(ontBalanceOfV2(native, a))

		assert.Nil(t, ontTransferFromV2(native, c, a, c, uint64(100*states.ScaleFactor)))
		return nil, nil
	})
}

func TestGovernanceUnboundV2(t *testing.T) {
	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))
		assert.Equal(t, ongAllowanceV2(native, utils.OntContractAddress, testAddr).String(), big.NewInt(5000000000*states.ScaleFactor).String())
		native.ContextRef.CurrentContext().ContractAddress = utils.OntContractAddress
		native.Time = native.Time + 100000
		native.Height = config.GetAddDecimalsHeight()
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, constants.ONT_TOTAL_SUPPLY_V2/2))

		alll := ongAllowanceV2(native, utils.OntContractAddress, testAddr)
		fmt.Println("alll:", alll.String())
		assert.Nil(t, ongTransferFromV2(native, testAddr, utils.OntContractAddress, testAddr, states.NativeTokenBalance{Balance: bigint.New(1)}))
		native.ContextRef.CurrentContext().ContractAddress = utils.OntContractAddress
		native.Time = native.Time + 100000
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))
		return nil, nil
	})

	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceContractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))
		assert.Nil(t, unboundGovernanceOng(native))
		assert.EqualValues(t, ongBalanceOfV2(native, gov).Add(ongBalanceOfV2(native, testAddr)).String(), constants.ONG_TOTAL_SUPPLY_V2.String())

		return nil, nil
	})

	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceContractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))
		assert.EqualValues(t, ongBalanceOfV2(native, gov).Add(ongBalanceOfV2(native, testAddr)).String(), constants.ONG_TOTAL_SUPPLY_V2.String())

		return nil, nil
	})

	InvokeNativeContract(t, utils.OntContractAddress, func(native *native.NativeService) ([]byte, error) {
		gov := utils.GovernanceContractAddress
		testAddr, _ := common.AddressParseFromBytes([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF})
		setOntBalance(native.CacheDB, testAddr, constants.ONT_TOTAL_SUPPLY)
		setOngBalance(native.CacheDB, utils.OntContractAddress, constants.ONG_TOTAL_SUPPLY)

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 1
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))
		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 10000
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))
		native.Time = config.GetOntHolderUnboundDeadline() - 100
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))

		native.Time = constants.GENESIS_BLOCK_TIMESTAMP + 18*constants.UNBOUND_TIME_INTERVAL

		assert.Nil(t, unboundGovernanceOng(native))
		assert.Nil(t, ontTransferV2(native, testAddr, testAddr, 1))
		assert.EqualValues(t, ongBalanceOfV2(native, gov).Add(ongBalanceOfV2(native, testAddr)).String(), constants.ONG_TOTAL_SUPPLY_V2.String())

		return nil, nil
	})
}
