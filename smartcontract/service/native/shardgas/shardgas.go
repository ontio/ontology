package shardgas

import (
	"bytes"
	"fmt"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

const (
	// function names
	INIT_NAME         = "init"
	DEPOSIT_GAS_NAME  = "depositGas"
	WITHDRAS_GAS_NAME = "withdrawGas"
	ACQUIRE_GAS_NAME  = "acquireWithdrawGas"

	// Key prefix
	KEY_VERSION = "version"
	KEY_BALANCE = "balance"
)

var ShardGasMgmtVersion = shardmgmt.VERSION_CONTRACT_SHARD_MGMT

func InitShardGasManagement() {
	native.Contracts[utils.ShardGasMgmtContractAddress] = RegisterShardGasMgmtContract
}

func RegisterShardGasMgmtContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardGasMgmtInit)
	native.Register(DEPOSIT_GAS_NAME, DespositGasToShard)
	native.Register(WITHDRAS_GAS_NAME, WithdrawGasFromShard)
	native.Register(ACQUIRE_GAS_NAME, AcquireWithdrawGasFromShard)
}

func ShardGasMgmtInit(native *native.NativeService) ([]byte, error) {
	// check if admin
	// get admin from database
	adminAddress, err := global_params.GetStorageRole(native,
		global_params.GenerateOperatorKey(utils.ParamContractAddress))
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("getAdmin, get admin error: %v", err)
	}

	//check witness
	if err := utils.ValidateOwner(native, adminAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard gas, checkWitness error: %v", err)
	}

	contract := native.ContextRef.CurrentContext().ContractAddress

	// check if shard-mgmt initialized
	ver, err := getVersion(native, contract)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("init shard gas, get version: %s", err)
	}
	if ver == 0 {
		// initialize shardmgmt version
		if err := setVersion(native, contract); err != nil {
			return utils.BYTE_FALSE, fmt.Errorf("init shard gas version: %s", err)
		}

		return utils.BYTE_TRUE, nil
	}

	if ver < ShardGasMgmtVersion {
		// make upgrade
		return utils.BYTE_FALSE, fmt.Errorf("upgrade TBD")
	}
	return utils.BYTE_FALSE, fmt.Errorf("version downgrade from %d to %d", ver, ShardGasMgmtVersion)
}

func DespositGasToShard(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid cmd param: %s", err)
	}

	params := new(DepositGasParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, params.UserAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid creator: %s", err)
	}
	if params.Amount > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, invalid amount")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, version check: %s", err)
	}
	if ok, err := checkShardID(native, params.ShardID); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, shardID check: %s", err)
	}

	gasInfo, err := getUserBalance(native, contract, params.ShardID, params.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, get user balance: %s", err)
	}

	// TODO: TRANSFER ONG

	gasInfo.Balance += params.Amount
	if err := setUserDeposit(native, contract, params.ShardID, params.UserAddress, gasInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, update user balance: %s", err)
	}

	evt := &shardstates.DepositGasEvent{
		SourceShardID: native.ShardID.ToUint64(),
		Height:        uint64(native.Height),
		ShardID:       params.ShardID,
		User:          params.UserAddress,
		Amount:        params.Amount,
	}
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func WithdrawGasFromShard(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid cmd param: %s", err)
	}

	params := new(WithdrawGasRequestParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, params.UserAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid creator: %s", err)
	}
	if params.Amount > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, invalid amount")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, version check: %s", err)
	}
	if ok, err := checkShardID(native, params.ShardID); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, shardID check: %s", err)
	}

	gasInfo, err := getUserBalance(native, contract, params.ShardID, params.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, get user balance: %s", err)
	}

	if gasInfo.Balance <= params.Amount {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, not enough balance for withdraw")
	}
	if len(gasInfo.PendingWithdraw) >= shardstates.CAP_PENDING_WITHDRAW {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, overlimited withdraw request")
	}

	gasInfo.Balance -= params.Amount
	gasInfo.WithdrawBalance += params.Amount
	gasInfo.PendingWithdraw = append(gasInfo.PendingWithdraw, &shardstates.GasWithdrawInfo{
		Height: uint64(native.Height),
		Amount: uint64(params.Amount),
	})

	if err := setUserDeposit(native, contract, params.ShardID, params.UserAddress, gasInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, update user balance: %s", err)
	}

	evt := &shardstates.WithdrawGasReqEvent{
		SourceShardID: native.ShardID.ToUint64(),
		Height:        uint64(native.Height),
		ShardID:       params.ShardID,
		User:          params.UserAddress,
		Amount:        params.Amount,
	}
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("withdraw gas, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func AcquireWithdrawGasFromShard(native *native.NativeService) ([]byte, error) {
	cp := new(CommonParam)
	if err := cp.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid cmd param: %s", err)
	}

	params := new(AcquireWithdrawGasParam)
	if err := params.Deserialize(bytes.NewBuffer(cp.Input)); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid param: %s", err)
	}
	if err := utils.ValidateOwner(native, params.UserAddress); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid creator: %s", err)
	}
	if params.Amount > constants.ONG_TOTAL_SUPPLY {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, invalid amount")
	}
	if native.Height <= shardstates.WITHDRAW_GAS_DELAY_DURATION {
		return utils.BYTE_FALSE, fmt.Errorf("acqure gas, pending to withdraw")
	}

	contract := native.ContextRef.CurrentContext().ContractAddress
	if ok, err := checkVersion(native, contract); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, version check: %s", err)
	}
	if ok, err := checkShardID(native, params.ShardID); !ok || err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, shardID check: %s", err)
	}

	gasInfo, err := getUserBalance(native, contract, params.ShardID, params.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, get user balance: %s", err)
	}

	if params.Amount > gasInfo.WithdrawBalance {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, not enough withdraw balance")
	}

	withdrawAmount := uint64(0)
	pendingWithdraw := make([]*shardstates.GasWithdrawInfo, 0)
	for _, w := range gasInfo.PendingWithdraw {
		if w.Height < uint64(native.Height)-shardstates.WITHDRAW_GAS_DELAY_DURATION {
			if withdrawAmount+w.Amount < params.Amount {
				withdrawAmount += w.Amount
			} else {
				w.Amount -= params.Amount - withdrawAmount
				withdrawAmount = params.Amount
				pendingWithdraw = append(pendingWithdraw, w)
			}
		} else {
			pendingWithdraw = append(pendingWithdraw, w)
		}
	}

	gasInfo.WithdrawBalance -= withdrawAmount
	gasInfo.PendingWithdraw = pendingWithdraw

	// TODO: transfer $withdrawAmount ong from contract to params.UserAddress

	if err := setUserDeposit(native, contract, params.ShardID, params.UserAddress, gasInfo); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, update user balance: %s", err)
	}

	evt := &shardstates.WithdrawGasDoneEvent{
		SourceShardID: native.ShardID.ToUint64(),
		Height:        uint64(native.Height),
		ShardID:       params.ShardID,
		User:          params.UserAddress,
		Amount:        withdrawAmount,
	}
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("acquire gas, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}
