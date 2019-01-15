package shardgas

import (
	"fmt"
	"bytes"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/smartcontract/service/native/shardgas/states"
)

const (
	// function names
	INIT_NAME         = "init"
	DEPOSIT_GAS_NAME  = "depositGas"
	WITHDRAS_GAS_NAME = "withdrawGas"

	// Key prefix
	KEY_VERSION  = "version"
	KEY_BALANCE  = "balance"
	KEY_WITHDRAW = "withdraw"
)

var ShardGasMgmtVersion = shardmgmt.VERSION_CONTRACT_SHARD_MGMT

func InitShardGasManagement() {
	native.Contracts[utils.ShardGasMgmtContractAddress] = RegisterShardGasMgmtContract
}

func RegisterShardGasMgmtContract(native *native.NativeService) {
	native.Register(INIT_NAME, ShardGasMgmtInit)
	native.Register(DEPOSIT_GAS_NAME, DespositGasToShard)
	native.Register(WITHDRAS_GAS_NAME, WithdrawGasFromShard)
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

	amount, err := getUserDespoit(native, contract, params.ShardID, params.UserAddress)
	if err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, get user balance: %s", err)
	}

	// TODO: TRANSFER ONG

	if err := setUserDeposit(native, contract, params.ShardID, params.UserAddress, amount + params.Amount); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, update user balance: %s", err)
	}

	evt := &shardgas_states.DepositGasEvent{
		ShardID: params.ShardID,
		User: params.UserAddress,
		Amount: params.Amount,
	}
	if err := shardmgmt.AddNotification(native, contract, evt); err != nil {
		return utils.BYTE_FALSE, fmt.Errorf("deposit gas, add notification: %s", err)
	}

	return utils.BYTE_TRUE, nil
}

func WithdrawGasFromShard(native *native.NativeService) ([]byte, error) {
	return utils.BYTE_FALSE, nil
}
