package oep4

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
	"math"
	"math/big"
)

// TODO: support user burn and mint

const (
	REGISTER = "oep4Register"

	NAME           = "oep4Name"
	SYMBOL         = "oep4Symbol"
	DECIMALS       = "oep4Decimals"
	TOTAL_SUPPLY   = "oep4TotalSupply" // query total supply at shard
	SHARD_SUPPLY   = "oep4ShardSupply" // query shard supply at root
	WHOLE_SUPPLY   = "oep4WholeSupply" // sum supply at all shard, only can be invoked at root
	BALANCE_OF     = "oep4BalanceOf"
	TRANSFER       = "oep4Transfer"
	TRANSFER_MULTI = "oep4TransferMulti"
	APPROVE        = "oep4Approve"
	TRANSFER_FROM  = "oep4TransferFrom"
	ALLOWANCE      = "oep4Allowance"
	MIGRATE        = "oep4Migrate"

	XSHARD_TRANSFER       = "oep4XShardTransfer"
	XSHARD_TRANFSER_RETRY = "oep4XShardTransferRetry"
	XSHARD_TRANSFER_SUCC  = "oep4XShardTransferSuccess"

	// call by shardsysmsg contract while xshard transfer
	SHARD_MINT = "oep4ShardMint"
)

func RegisterOEP4(native *native.NativeService) {
	native.Register(REGISTER, Register)
	native.Register(MIGRATE, Migrate)

	native.Register(NAME, Name)
	native.Register(SYMBOL, Symbol)
	native.Register(DECIMALS, Decimals)
	native.Register(TOTAL_SUPPLY, TotalSupply)
	native.Register(SHARD_SUPPLY, ShardSupply)
	native.Register(WHOLE_SUPPLY, WholeSupply)
	native.Register(BALANCE_OF, BalanceOf)
	native.Register(TRANSFER, Transfer)
	native.Register(TRANSFER_MULTI, TransferMulti)
	native.Register(APPROVE, Approve)
	native.Register(TRANSFER_FROM, TransferFrom)
	native.Register(ALLOWANCE, Allowance)
	native.Register(XSHARD_TRANSFER, XShardTransfer)
	native.Register(XSHARD_TRANFSER_RETRY, XShardTransferRetry)
	native.Register(XSHARD_TRANSFER_SUCC, XShardTransferSuccess)

	native.Register(SHARD_MINT, ShardMint)
}

func Register(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("Register: only can be invoked at root")
	}
	param := &RegisterParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	isReg, err := isAssetRegister(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, err: %s", err)
	}
	if isReg {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, asset has already registered")
	}
	assetNum, err := getAssetNum(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, err: %s", err)
	}
	if assetNum == math.MaxUint64 {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, asset num exceed")
	}
	assetId := assetNum + 1
	registerAsset(native, callAddr, assetId)
	setAssetNum(native, assetId)
	oep4 := &Oep4{
		Name:        param.Name,
		Symbol:      param.Symbol,
		Decimals:    param.Decimals,
		TotalSupply: param.TotalSupply,
	}
	setContract(native, assetId, oep4)
	shardSupplyInfo := map[types.ShardID]*big.Int{native.ShardID: param.TotalSupply}
	setShardSupplyInfo(native, assetId, shardSupplyInfo)
	transferEvent := &TransferEvent{
		From:   common.ADDRESS_EMPTY,
		To:     param.Account,
		Amount: param.TotalSupply,
	}
	NotifyEvent(native, transferEvent.ToNotify())
	return utils.BYTE_TRUE, nil
}

func Migrate(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("Migrate: only can be invoked at root")
	}
	param := &MigrateParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Migrate: failed, err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	isReg, err := isAssetRegister(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, err: %s", err)
	}
	if !isReg {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, asset has not registered")
	}
	assetId, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Register: failed, err: %s", err)
	}
	deleteAssetId(native, callAddr)
	registerAsset(native, param.NewAsset, assetId)
	return utils.BYTE_TRUE, nil
}

func Name(native *native.NativeService) ([]byte, error) {
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Name: failed, err: %s", err)
	}
	oep4, err := getContract(native, asset)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Name: failed, err: %s", err)
	}
	return []byte(oep4.Name), nil
}

func Symbol(native *native.NativeService) ([]byte, error) {
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Symbol: failed, err: %s", err)
	}
	oep4, err := getContract(native, asset)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Symbol: failed, err: %s", err)
	}
	return []byte(oep4.Symbol), nil
}

func Decimals(native *native.NativeService) ([]byte, error) {
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Decimals: failed, err: %s", err)
	}
	oep4, err := getContract(native, asset)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Decimals: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(new(big.Int).SetUint64(oep4.Decimals)), nil
}

func TotalSupply(native *native.NativeService) ([]byte, error) {
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TotalSupply: failed, err: %s", err)
	}
	oep4, err := getContract(native, asset)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TotalSupply: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(oep4.TotalSupply), nil
}

func ShardSupply(native *native.NativeService) ([]byte, error) {
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TotalSupply: failed, err: %s", err)
	}
	shardSupplyInfo, err := getShardSupplyInfo(native, asset)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TotalSupply: failed, err: %s", err)
	}
	shardId, err := utils.DeserializeShardId(bytes.NewReader(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TotalSupply: deserialize param failed, err: %s", err)
	}
	supply, ok := shardSupplyInfo[shardId]
	if ok {
		return ntypes.BigIntToBytes(supply), nil
	} else {
		return ntypes.BigIntToBytes(big.NewInt(0)), nil
	}
}

func WholeSupply(native *native.NativeService) ([]byte, error) {
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TotalSupply: failed, err: %s", err)
	}
	shardSupplyInfo, err := getShardSupplyInfo(native, asset)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TotalSupply: failed, err: %s", err)
	}
	whole := new(big.Int)
	for _, supply := range shardSupplyInfo {
		whole.Add(whole, supply)
	}
	return ntypes.BigIntToBytes(whole), nil
}

func BalanceOf(native *native.NativeService) ([]byte, error) {
	param := &BalanceParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BalanceOf: failed, err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BalanceOf: failed, err: %s", err)
	}
	userBalance, err := getUserBalance(native, asset, param.User)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("BalanceOf: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(userBalance), nil
}

func Transfer(native *native.NativeService) ([]byte, error) {
	param := &TransferParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Transfer: failed, err: %s", err)
	}
	if err := transfer(native, param); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Transfer: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func TransferMulti(native *native.NativeService) ([]byte, error) {
	param := &MultiTransferParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TransferMulti: failed, err: %s", err)
	}
	for index, tranParam := range param.Transfers {
		if err := transfer(native, tranParam); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("TransferMulti: failed, index %d, err: %s", index, err)
		}
	}
	return utils.BYTE_TRUE, nil
}

func Approve(native *native.NativeService) ([]byte, error) {
	param := &ApproveParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Approve: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.Owner); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Approve: check witness err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Approve: failed, err: %s", err)
	}
	balance, err := getUserBalance(native, asset, param.Owner)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Approve: failed, err: %s", err)
	}
	if balance.Cmp(param.Allowance) < 0 {
		return utils.BYTE_FALSE, fmt.Errorf("Approve: owner balance not enough")
	}
	setUserAllowance(native, asset, param.Owner, param.Spender, param.Allowance)
	event := &ApproveEvent{Asset: callAddr, Owner: param.Owner, Spender: param.Spender, Allowance: param.Allowance}
	NotifyEvent(native, event.ToNotify())
	return utils.BYTE_TRUE, nil
}

func TransferFrom(native *native.NativeService) ([]byte, error) {
	param := &TransferFromParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.Spender); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: check witness err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: failed, err: %s", err)
	}
	allowance, err := getUserAllowance(native, asset, param.From, param.Spender)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: failed, err: %s", err)
	}
	if allowance.Cmp(param.Amount) < 0 {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: allowance not enough")
	}
	fromBalance, err := getUserBalance(native, asset, param.From)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: get form balance failed, err: %s", err)
	}
	if fromBalance.Cmp(param.Amount) < 0 {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: from balance not enough")
	}
	allowance.Sub(allowance, param.Amount)
	setUserAllowance(native, asset, param.From, param.Spender, allowance)
	fromBalance.Sub(fromBalance, param.Amount)
	setUserBalance(native, asset, param.From, fromBalance)
	toBalance, err := getUserBalance(native, asset, param.To)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("TransferFrom: get to balance failed, err: %s", err)
	}
	toBalance.Add(toBalance, param.Amount)
	setUserBalance(native, asset, param.To, toBalance)
	return utils.BYTE_TRUE, nil
}

func Allowance(native *native.NativeService) ([]byte, error) {
	param := &AllowanceParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Allowance: failed, err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Allowance: failed, err: %s", err)
	}
	allowance, err := getUserAllowance(native, asset, param.Owner, param.Spender)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Allowance: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(allowance), nil
}

func XShardTransfer(native *native.NativeService) ([]byte, error) {
	param := &XShardTransferParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	// check shard id
	if native.ShardID.ToUint64() == param.ToShard.ToUint64() {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: unsupport transfer in same shard")
	}
	if !native.ShardID.IsRootShard() && !param.ToShard.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: unsupport transfer between shard")
	}

	if err := utils.ValidateOwner(native, param.From); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: check witness err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	if err := userBurn(native, asset, param.To, param.Amount); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: check witness err: %s", err)
	}
	transferNum, err := getXShardTransferNum(native, asset, param.From)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	transferNum.Add(transferNum, big.NewInt(1))
	transfer := &XShardTransferState{
		ToShard:   param.ToShard,
		ToAccount: param.To,
		Amount:    param.Amount,
		Status:    XSHARD_TRANSFER_PENDING,
	}
	setXShardTransfer(native, asset, param.From, transferNum, transfer)
	setXShardTransferNum(native, asset, param.From, transferNum)
	shardMintParam := &ShardMintParam{
		Asset:      asset,
		Account:    param.To,
		Amount:     param.Amount,
		TransferId: transferNum,
	}
	if err := notifyShardMint(native, param.ToShard, shardMintParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func XShardTransferRetry(native *native.NativeService) ([]byte, error) {
	param := &XShardTransferRetryParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferRetry: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.From); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferRetry: check witness err: %s", err)
	}
	callAddr := native.ContextRef.CallingContext().ContractAddress
	asset, err := getAssetId(native, callAddr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	transfer, err := getXShardTransfer(native, asset, param.From, param.TransferId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: failed, err: %s", err)
	}
	if transfer.Status == XSHARD_TRANSFER_COMPLETE {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: transfer has already completed")
	}
	shardMintParam := &ShardMintParam{
		Asset:      asset,
		Account:    transfer.ToAccount,
		Amount:     transfer.Amount,
		TransferId: param.TransferId,
	}
	if err := notifyShardMint(native, transfer.ToShard, shardMintParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func XShardTransferSuccess(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardSysMsgContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: only can be invoked by shard sysmsg")
	}
	param := &XShardTranSuccParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: failed, err: %s", err)
	}
	transfer, err := getXShardTransfer(native, param.Asset, param.Account, param.TransferId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: failed, err: %s", err)
	}
	transfer.Status = XSHARD_TRANSFER_COMPLETE
	setXShardTransfer(native, param.Asset, param.Account, param.TransferId, transfer)
	if native.ShardID.IsRootShard() {
		supplyInfo, err := getShardSupplyInfo(native, param.Asset)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: failed, err: %s", err)
		}
		if rootSupply, ok := supplyInfo[native.ShardID]; ok {
			if rootSupply.Cmp(transfer.Amount) < 0 {
				return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: root supply not enough")
			}
			rootSupply.Sub(rootSupply, transfer.Amount)
			supplyInfo[native.ShardID] = rootSupply
		} else {
			return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: root supply not exist")
		}
		if shardSupply, ok := supplyInfo[transfer.ToShard]; ok {
			shardSupply.Add(shardSupply, transfer.Amount)
			supplyInfo[transfer.ToShard] = shardSupply
		} else {
			supplyInfo[transfer.ToShard] = transfer.Amount
		}
		setShardSupplyInfo(native, param.Asset, supplyInfo)
	}
	return utils.BYTE_TRUE, nil
}

func ShardMint(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardSysMsgContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("ShardMint: only can be invoked by shard sysmsg")
	}
	param := &ShardMintParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardMint: failed, err: %s", err)
	}
	isReceived, err := isTransferReceived(native, param)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardMint: failed, err: %s", err)
	}
	if !isReceived {
		if err := userMint(native, param.Asset, param.Account, param.Amount); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ShardMint: failed, err: %s", err)
		}
		receiveTransfer(native, param)
	}
	if native.ShardID.IsRootShard() {
		supplyInfo, err := getShardSupplyInfo(native, param.Asset)
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: failed, err: %s", err)
		}
		if shardSupply, ok := supplyInfo[param.FromShard]; ok {
			if shardSupply.Cmp(param.Amount) < 0 {
				return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: shard supply not enough")
			}
			shardSupply.Sub(shardSupply, param.Amount)
			supplyInfo[native.ShardID] = shardSupply
		} else {
			return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: shard supply not exist")
		}
		if rootSupply, ok := supplyInfo[native.ShardID]; ok {
			rootSupply.Add(rootSupply, param.Amount)
			supplyInfo[native.ShardID] = rootSupply
		} else {
			return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSuccess: root supply not exist")
		}
		setShardSupplyInfo(native, param.Asset, supplyInfo)
	}

	tranSuccParam := &XShardTranSuccParam{
		Asset:      param.Asset,
		Account:    param.Account,
		TransferId: param.TransferId,
	}
	if err := notifyTransferSuccess(native, param.FromShard, tranSuccParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardMint: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}
