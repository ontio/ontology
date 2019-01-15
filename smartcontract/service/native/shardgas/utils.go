package shardgas

import (
	"bytes"
	"fmt"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/serialization"
	cstates "github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/utils"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt/states"
)

func getVersion(native *native.NativeService, contract common.Address) (uint32, error) {
	versionBytes, err := native.CacheDB.Get(utils.ConcatKey(contract, []byte(KEY_VERSION)))
	if err != nil {
		return 0, fmt.Errorf("get version: %s", err)
	}

	if versionBytes == nil {
		return 0, nil
	}

	value, err := cstates.GetValueFromRawStorageItem(versionBytes)
	if err != nil {
		return 0, fmt.Errorf("get versoin, deserialized from raw storage item: %s", err)
	}

	ver, err := serialization.ReadUint32(bytes.NewBuffer(value))
	if err != nil {
		return 0, fmt.Errorf("serialization.ReadUint32, deserialize version: %s", err)
	}
	return ver, nil
}

func setVersion(native *native.NativeService, contract common.Address) error {
	buf := new(bytes.Buffer)
	if err := serialization.WriteUint32(buf, ShardGasMgmtVersion); err != nil {
		return fmt.Errorf("failed to serialize version: %s", err)
	}

	native.CacheDB.Put(utils.ConcatKey(contract, []byte(KEY_VERSION)), cstates.GenRawStorageItem(buf.Bytes()))
	return nil
}

func checkVersion(native *native.NativeService, contract common.Address) (bool, error) {
	ver, err := getVersion(native, contract)
	if err != nil {
		return false, err
	}
	return ver == ShardGasMgmtVersion, nil
}

func checkShardID(native *native.NativeService, shardID uint64) (bool, error) {
	shardState, err := shardmgmt.GetShardState(native, utils.ShardMgmtContractAddress, shardID)
	if err != nil {
		return false, err
	}

	if shardState == nil {
		return false, fmt.Errorf("invalid shard %d", shardID)
	}

	return shardState.State == shardstates.SHARD_STATE_ACTIVE, nil
}

func getUserDespoit(native *native.NativeService, contract common.Address, shardID uint64, user common.Address) (uint64, error) {
	shardIDByte, err := shardutil.GetUint64Bytes(shardID)
	if err != nil {
		return 0, fmt.Errorf("ser ShardID %s", err)
	}
	keyBytes := utils.ConcatKey(contract, []byte(KEY_BALANCE), shardIDByte, user[:])
	amountBytes, err := native.CacheDB.Get(keyBytes)
	if err != nil {
		return 0, fmt.Errorf("get balance from db: %s", err)
	}
	if len(amountBytes) == 0 {
		return 0, nil
	}

	return shardutil.GetBytesUint64(amountBytes)
}

func setUserDeposit(native *native.NativeService, contract common.Address, shardID uint64, user common.Address, amount uint64) error {
	shardIDByte, err := shardutil.GetUint64Bytes(shardID)
	if err != nil {
		return fmt.Errorf("ser ShardID %s", err)
	}
	keyBytes := utils.ConcatKey(contract, []byte(KEY_BALANCE), shardIDByte, user[:])

	amountBytes, err := shardutil.GetUint64Bytes(amount)
	if err != nil {
		return fmt.Errorf("ser Amount %s", err)
	}
	native.CacheDB.Put(keyBytes, cstates.GenRawStorageItem(amountBytes))
	return nil
}
