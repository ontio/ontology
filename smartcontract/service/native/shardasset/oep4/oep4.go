package oep4

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"math"
	"math/big"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	ntypes "github.com/ontio/ontology/vm/neovm/types"
)

// TODO: support user burn and mint

const (
	INIT = "ope4ShardAssetInit"

	REGISTER = "oep4Register"
	MIGRATE  = "oep4Migrate"

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
	ASSET_ID       = "oep4AssetId"

	XSHARD_TRANSFER       = "oep4XShardTransfer"
	XSHARD_TRANFSER_RETRY = "oep4XShardTransferRetry"

	ONG_XSHARD_TRANSFER       = "ongXShardTransfer"
	ONG_XSHARD_TRANSFER_RETRY = "ongXShardTransferRetry"

	// call by shardsysmsg contract
	XSHARD_TRANSFER_SUCC = "oep4XShardTransferSuccess"
	SHARD_RECEIVE_ASSET  = "oep4ShardReceive"
	ONG_XSHARD_RECEIVE   = "ongXShardReceive"

	GET_PENDING_TRANSFER = "getOep4PendingTransfer"
	GET_TRANSFER         = "getOep4Transfer"
)

func RegisterOEP4(native *native.NativeService) {
	native.Register(INIT, Init)
	native.Register(REGISTER, Register)
	native.Register(ASSET_ID, GetAssetId)
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
	native.Register(XSHARD_TRANSFER_SUCC, XShardTransferSucc)
	native.Register(ONG_XSHARD_TRANSFER, XShardTransferOng)
	native.Register(ONG_XSHARD_TRANSFER_RETRY, XShardTransferOngRetry)
	native.Register(ONG_XSHARD_RECEIVE, XShardReceiveOng)

	native.Register(SHARD_RECEIVE_ASSET, ShardReceiveAsset)

	native.Register(GET_PENDING_TRANSFER, GetPendingXShardTransfer)
	native.Register(GET_TRANSFER, GetXShardTransferState)
}

func Init(native *native.NativeService) ([]byte, error) {
	if !native.ShardID.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("Init: only can be invoked at root")
	}
	isInit, err := isOep4ShardAssetInit(native)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("Init: failed, err: %s", err)
	}
	if isInit {
		return utils.BYTE_FALSE, fmt.Errorf("Init: has already init")
	}
	initOep4ShardAsset(native)
	supplyInfo := make(map[types.ShardID]*big.Int)
	supplyInfo[native.ShardID] = new(big.Int).SetUint64(constants.ONG_TOTAL_SUPPLY)
	setShardSupplyInfo(native, ONG_ASSET_ID, supplyInfo)
	return utils.BYTE_TRUE, nil
}

// assetId start form 1, ong id is 0
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
	setAssetNum(native, assetNum+1)
	assetId := AssetId(assetNum + 1)
	registerAsset(native, callAddr, assetId)
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

func GetAssetId(native *native.NativeService) ([]byte, error) {
	addr, err := utils.ReadAddress(bytes.NewReader(native.Input))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetAssetId: read param addr failed, err: %s", err)
	}
	assetId, err := getAssetId(native, addr)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetAssetId: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(new(big.Int).SetUint64(uint64(assetId))), nil
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
	txId, err := xShardTransfer(native, asset, param.From, param.To, param.ToShard, param.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	shardMintParam := &ShardMintParam{
		Asset:      uint64(asset),
		Account:    param.To,
		Amount:     param.Amount,
		TransferId: txId,
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
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferRetry: failed, err: %s", err)
	}
	transfer, err := getXShardTransfer(native, asset, param.From, param.TransferId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferRetry: failed, err: %s", err)
	}
	if transfer.Status == XSHARD_TRANSFER_COMPLETE {
		return utils.BYTE_TRUE, nil
	}
	shardMintParam := &ShardMintParam{
		Asset:      uint64(asset),
		Account:    transfer.ToAccount,
		Amount:     transfer.Amount,
		TransferId: param.TransferId,
	}
	if err := notifyShardMint(native, transfer.ToShard, shardMintParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferRetry: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func XShardTransferSucc(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardSysMsgContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSucc: only can be invoked by shard sysmsg")
	}
	param := &XShardTranSuccParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSucc: failed, err: %s", err)
	}
	transfer, err := getXShardTransfer(native, AssetId(param.Asset), param.Account, param.TransferId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSucc: failed, err: %s", err)
	}
	transfer.Status = XSHARD_TRANSFER_COMPLETE
	setXShardTransfer(native, AssetId(param.Asset), param.Account, param.TransferId, transfer)
	if native.ShardID.IsRootShard() {
		if err := rootTransferSucc(native, transfer.ToShard, AssetId(param.Asset), transfer.Amount); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("XShardTransferSucc: failed, err: %s", err)
		}
	}
	return utils.BYTE_TRUE, nil
}

func ShardReceiveAsset(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardSysMsgContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("ShardReceiveAsset: only can be invoked by shard sysmsg")
	}
	param := &ShardMintParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardReceiveAsset: failed, err: %s", err)
	}
	isReceived, err := isTransferReceived(native, param)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardReceiveAsset: failed, err: %s", err)
	}
	if !isReceived {
		if err := userMint(native, AssetId(param.Asset), param.Account, param.Amount); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ShardReceiveAsset: failed, err: %s", err)
		}
		receiveTransfer(native, param)
	}
	if native.ShardID.IsRootShard() {
		if err := rootReceiveAsset(native, param.FromShard, AssetId(param.Asset), param.Amount); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("ShardReceiveAsset: failed, err: %s", err)
		}
	}

	tranSuccParam := &XShardTranSuccParam{
		Asset:      param.Asset,
		Account:    param.Account,
		TransferId: param.TransferId,
	}
	if err := notifyTransferSuccess(native, param.FromShard, tranSuccParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("ShardReceiveAsset: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func XShardTransferOng(native *native.NativeService) ([]byte, error) {
	param := &XShardTransferParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOng: failed, err: %s", err)
	}
	// check shard id
	if native.ShardID.ToUint64() == param.ToShard.ToUint64() {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOng: unsupport transfer in same shard")
	}
	if !native.ShardID.IsRootShard() && !param.ToShard.IsRootShard() {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOng: unsupport transfer between shard")
	}
	if err := utils.ValidateOwner(native, param.From); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOng: check witness err: %s", err)
	}
	err := ont.AppCallTransfer(native, utils.OngContractAddress, param.From, utils.ShardAssetAddress, param.Amount.Uint64())
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOng: transfer ong failed, err: %s", err)
	}
	txId, err := xShardTransfer(native, ONG_ASSET_ID, param.From, param.To, param.ToShard, param.Amount)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOng: failed, err: %s", err)
	}
	shardMintParam := &ShardMintParam{
		Asset:      uint64(ONG_ASSET_ID),
		Account:    param.To,
		Amount:     param.Amount,
		TransferId: txId,
	}
	if err := notifyShardReceiveOng(native, param.ToShard, shardMintParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransfer: failed, err: %s", err)
	}
	return ntypes.BigIntToBytes(txId), nil
}

func XShardTransferOngRetry(native *native.NativeService) ([]byte, error) {
	param := &XShardTransferRetryParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOngRetry: failed, err: %s", err)
	}
	if err := utils.ValidateOwner(native, param.From); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOngRetry: check witness err: %s", err)
	}
	transfer, err := getXShardTransfer(native, ONG_ASSET_ID, param.From, param.TransferId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOngRetry: failed, err: %s", err)
	}
	if transfer.Status == XSHARD_TRANSFER_COMPLETE {
		return utils.BYTE_TRUE, nil
	}
	shardMintParam := &ShardMintParam{
		Asset:      uint64(ONG_ASSET_ID),
		Account:    transfer.ToAccount,
		Amount:     transfer.Amount,
		TransferId: param.TransferId,
	}
	if err := notifyShardReceiveOng(native, transfer.ToShard, shardMintParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardTransferOngRetry: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func XShardReceiveOng(native *native.NativeService) ([]byte, error) {
	if native.ContextRef.CallingContext().ContractAddress != utils.ShardSysMsgContractAddress {
		return utils.BYTE_FALSE, fmt.Errorf("XShardReceiveOng: only can be invoked by shard sysmsg")
	}
	param := &ShardMintParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardReceiveOng: failed, err: %s", err)
	}
	isReceived, err := isTransferReceived(native, param)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardReceiveOng: failed, err: %s", err)
	}
	if !isReceived {
		err := ont.AppCallTransfer(native, utils.OngContractAddress, utils.ShardAssetAddress, param.Account,
			param.Amount.Uint64())
		if err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("XShardReceiveOng: failed, err: %s", err)
		}
		receiveTransfer(native, param)
	}
	if native.ShardID.IsRootShard() {
		if err := rootReceiveAsset(native, param.FromShard, ONG_ASSET_ID, param.Amount); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("XShardReceiveOng: failed, err: %s", err)
		}
	}
	tranSuccParam := &XShardTranSuccParam{
		Asset:      param.Asset,
		Account:    param.Account,
		TransferId: param.TransferId,
	}
	if err := notifyTransferSuccess(native, param.FromShard, tranSuccParam); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("XShardReceiveOng: failed, err: %s", err)
	}
	return utils.BYTE_TRUE, nil
}

func GetPendingXShardTransfer(native *native.NativeService) ([]byte, error) {
	param := &GetPendingXShardTransferParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetPendingXShardTransfer: failed, err: %s", err)
	}
	transferNum, err := getXShardTransferNum(native, AssetId(param.Asset), param.Account)
	if err != nil {
		return nil, fmt.Errorf("GetPendingXShardTransfer: failed, err: %s", err)
	}
	increase := big.NewInt(1)
	transfers := make([]*XShardTransferState, 0)
	for i := big.NewInt(0); i.Cmp(transferNum) < 0; i.Add(i, increase) {
		transfer, err := getXShardTransfer(native, AssetId(param.Asset), param.Account, i)
		if err != nil {
			log.Debugf("GetPendingXShardTransfer: read transfer failed, tranId %s, err: %s", i.String(), err)
		}
		transfers = append(transfers, transfer)
	}
	data, err := json.Marshal(transfers)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetPendingXShardTransfer: marshal failed, err: %s", err)
	}
	return data, nil
}

func GetXShardTransferState(native *native.NativeService) ([]byte, error) {
	param := &GetXShardTransferInfoParam{}
	if err := param.Deserialize(bytes.NewReader(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetXShardTransferState: failed, err: %s", err)
	}
	transfer, err := getXShardTransfer(native, AssetId(param.Asset), param.Account, param.TransferId)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetXShardTransferState: failed, err: %s", err)
	}
	data, err := json.Marshal(transfer)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("GetXShardTransferState: marshal info failed, err: %s", err)
	}
	return data, nil
}
