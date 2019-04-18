package oep4

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	shardsysmsg "github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func transfer(native *native.NativeService, param *TransferParam) error {
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return fmt.Errorf("transfer: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.From); err != nil {
		return fmt.Errorf("transfer: check witness err: %s", err)
	}
	fromBalance, err := getUserBalance(native, asset, param.From)
	if err != nil {
		return fmt.Errorf("transfer: get from balance failed, err: %s", err)
	}
	if fromBalance.Cmp(param.Amount) < 0 {
		return fmt.Errorf("transfer: from balance not enough failed, err: %s", err)
	}
	fromBalance.Sub(fromBalance, param.Amount)
	setUserBalance(native, asset, param.From, fromBalance)
	toBalance, err := getUserBalance(native, asset, param.From)
	if err != nil {
		return fmt.Errorf("transfer: get to balance failed, err: %s", err)
	}
	toBalance.Add(toBalance, param.Amount)
	setUserBalance(native, asset, param.To, toBalance)
	// push event
	event := &TransferEvent{Asset: callAddr, From: param.From, To: param.To, Amount: param.Amount}
	NotifyEvent(native, event.ToNotify())
	return nil
}

func userBurn(native *native.NativeService, asset AssetId, user common.Address, amount *big.Int) error {
	oep4, err := getContract(native, asset)
	if err != nil {
		return fmt.Errorf("userBurn: failed, err: %s", err)
	}
	if oep4.TotalSupply.Cmp(amount) < 0 {
		return fmt.Errorf("userBurn: total supply not enough")
	}
	oep4.TotalSupply.Sub(oep4.TotalSupply, amount)
	setContract(native, asset, oep4)
	balance, err := getUserBalance(native, asset, user)
	if err != nil {
		return fmt.Errorf("userBurn: failed, err: %s", err)
	}
	if balance.Cmp(amount) < 0 {
		return fmt.Errorf("userBurn: from balance not enough")
	}
	balance.Sub(balance, amount)
	setUserBalance(native, asset, user, balance)
	return nil
}

func userMint(native *native.NativeService, asset AssetId, user common.Address, amount *big.Int) error {
	oep4, err := getContract(native, asset)
	if err != nil {
		return fmt.Errorf("userMint: failed, err: %s", err)
	}
	oep4.TotalSupply.Add(oep4.TotalSupply, amount)
	setContract(native, asset, oep4)
	balance, err := getUserBalance(native, asset, user)
	if err != nil {
		return fmt.Errorf("userMint: failed, err: %s", err)
	}
	balance.Add(balance, amount)
	setUserBalance(native, asset, user, balance)
	return nil
}

func xShardTransfer(native *native.NativeService, asset AssetId, from, to common.Address, toShard types.ShardID,
	amount *big.Int) (*big.Int, error) {
	transferNum, err := getXShardTransferNum(native, asset, from)
	if err != nil {
		return nil, fmt.Errorf("xShardTransfer: failed, err: %s", err)
	}
	transferNum.Add(transferNum, big.NewInt(1))
	transfer := &XShardTransferState{
		ToShard:   toShard,
		ToAccount: to,
		Amount:    amount,
		Status:    XSHARD_TRANSFER_PENDING,
	}
	setXShardTransfer(native, asset, from, transferNum, transfer)
	setXShardTransferNum(native, asset, from, transferNum)
	return transferNum, nil
}

func rootReceiveAsset(native *native.NativeService, fromShard types.ShardID, asset AssetId, amount *big.Int) error {
	supplyInfo, err := getShardSupplyInfo(native, asset)
	if err != nil {
		return fmt.Errorf("rootReceiveAsset: failed, err: %s", err)
	}
	if shardSupply, ok := supplyInfo[fromShard]; ok {
		if shardSupply.Cmp(amount) < 0 {
			return fmt.Errorf("rootReceiveAsset: shard supply not enough")
		}
		shardSupply.Sub(shardSupply, amount)
		supplyInfo[native.ShardID] = shardSupply
	} else {
		return fmt.Errorf("rootReceiveAsset: shard supply not exist")
	}
	if rootSupply, ok := supplyInfo[native.ShardID]; ok {
		rootSupply.Add(rootSupply, amount)
		supplyInfo[native.ShardID] = rootSupply
	} else {
		return fmt.Errorf("rootReceiveAsset: root supply not exist")
	}
	setShardSupplyInfo(native, asset, supplyInfo)
	return nil
}

func rootTransferSucc(native *native.NativeService, toShard types.ShardID, asset AssetId, amount *big.Int) error {
	supplyInfo, err := getShardSupplyInfo(native, asset)
	if err != nil {
		return fmt.Errorf("rootTransferSucc: failed, err: %s", err)
	}
	if rootSupply, ok := supplyInfo[native.ShardID]; ok {
		if rootSupply.Cmp(amount) < 0 {
			return fmt.Errorf("rootTransferSucc: root supply not enough")
		}
		rootSupply.Sub(rootSupply, amount)
		supplyInfo[native.ShardID] = rootSupply
	} else {
		return fmt.Errorf("rootTransferSucc: root supply not exist")
	}
	if shardSupply, ok := supplyInfo[toShard]; ok {
		shardSupply.Add(shardSupply, amount)
		supplyInfo[toShard] = shardSupply
	} else {
		supplyInfo[toShard] = amount
	}
	setShardSupplyInfo(native, asset, supplyInfo)
	return nil
}

func notifyShardMint(native *native.NativeService, toShard types.ShardID, param *ShardMintParam) error {
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("notifyShardMint: failed, err: %s", err)
	}
	if err := shardsysmsg.NotifyShard(native, toShard, SHARD_RECEIVE_ASSET, bf.Bytes()); err != nil {
		return fmt.Errorf("notifyShardMint: failed, err: %s", err)
	}
	return nil
}

func notifyTransferSuccess(native *native.NativeService, toShard types.ShardID, param *XShardTranSuccParam) error {
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("notifyTransferSuccess: failed, err: %s", err)
	}
	if err := shardsysmsg.NotifyShard(native, toShard, XSHARD_TRANSFER_SUCC, bf.Bytes()); err != nil {
		return fmt.Errorf("notifyTransferSuccess: failed, err: %s", err)
	}
	return nil
}

func notifyShardReceiveOng(native *native.NativeService, toShard types.ShardID, param *ShardMintParam) error {
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("notifyShardReceiveOng: failed, err: %s", err)
	}
	if err := shardsysmsg.NotifyShard(native, toShard, ONG_XSHARD_RECEIVE, bf.Bytes()); err != nil {
		return fmt.Errorf("notifyShardReceiveOng: failed, err: %s", err)
	}
	return nil
}
