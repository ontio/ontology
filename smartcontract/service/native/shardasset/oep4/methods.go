package oep4

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	shardsysmsg "github.com/ontio/ontology/smartcontract/service/native/shard_sysmsg"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"math/big"
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

func userBurn(native *native.NativeService, asset uint64, user common.Address, amount *big.Int) error {
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

func userMint(native *native.NativeService, asset uint64, user common.Address, amount *big.Int) error {
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

func notifyShardMint(native *native.NativeService, toShard types.ShardID, param *ShardMintParam) error {
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("notifyShardMint: failed, err: %s", err)
	}
	if err := notifyShard(native, toShard, SHARD_MINT, bf.Bytes()); err != nil {
		return fmt.Errorf("notifyShardMint: failed, err: %s", err)
	}
	return nil
}

func notifyTransferSuccess(native *native.NativeService, toShard types.ShardID, param *XShardTranSuccParam) error {
	bf := new(bytes.Buffer)
	if err := param.Serialize(bf); err != nil {
		return fmt.Errorf("notifyTransferSuccess: failed, err: %s", err)
	}
	if err := notifyShard(native, toShard, XSHARD_TRANSFER_SUCC, bf.Bytes()); err != nil {
		return fmt.Errorf("notifyTransferSuccess: failed, err: %s", err)
	}
	return nil
}

func notifyShard(native *native.NativeService, toShard types.ShardID, method string, args []byte) error {
	paramBytes := new(bytes.Buffer)
	params := shardsysmsg.NotifyReqParam{
		ToShard:    toShard,
		ToContract: utils.ShardAssetAddress,
		Method:     method,
		Args:       args,
	}
	if err := params.Serialize(paramBytes); err != nil {
		return fmt.Errorf("notifyShard: serialize param failed, err: %s", err)
	}
	if _, err := native.NativeCall(utils.ShardSysMsgContractAddress, shardsysmsg.REMOTE_NOTIFY, paramBytes.Bytes()); err != nil {
		return fmt.Errorf("notifyShard: invoke failed, err: %s", err)
	}
	return nil
}
