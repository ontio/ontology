package oep4

import (
	"fmt"
	"io"
	"math/big"
	"sort"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

const ONG_ASSET_ID AssetId = 0

const (
	KEY_INIT = "oep4_init"

	KEY_OEP4_ASSET_NUM = "oep4_asset_num"
	KEY_OEP4_ASSET_ID  = "oep4_asset_id"

	KEY_OEP4                 = "oep4"
	KEY_OEP4_BALANCE         = "oep4_balance"
	KEY_OEP4_SHARD_SUPPLY    = "oep4_shard_supply" // asset distribute shard
	KEY_OEP4_ALLOWANCE       = "oep4_allowance"
	KEY_OEP4_TRANSFER_NUM    = "oep4_transfer_num"
	KEY_OEP4_XSHARD_TRANSFER = "oep4_xshard_transfer"
	KEY_OEP4_XSHARD_RECEIVE  = "oep4_xshard_receive"
)

func genInitKey() []byte {
	return utils.ConcatKey(utils.ShardAssetAddress, []byte(KEY_INIT))
}

func genAssetNumKey() []byte {
	return utils.ConcatKey(utils.ShardAssetAddress, []byte(KEY_OEP4_ASSET_NUM))
}

func genAssetIdKey(assetAddr common.Address) []byte {
	return utils.ConcatKey(utils.ShardAssetAddress, []byte(KEY_OEP4_ASSET_ID), assetAddr[:])
}

func genAssetKey(asset AssetId) []byte {
	assetBytes := utils.GetUint64Bytes(uint64(asset))
	return utils.ConcatKey(utils.ShardAssetAddress, assetBytes, []byte(KEY_OEP4))
}

func genBalanceKey(asset AssetId, user common.Address) []byte {
	assetBytes := utils.GetUint64Bytes(uint64(asset))
	return utils.ConcatKey(utils.ShardAssetAddress, assetBytes, []byte(KEY_OEP4_BALANCE), user[:])
}

func genShardSupplyInfoKey(asset AssetId) []byte {
	assetBytes := utils.GetUint64Bytes(uint64(asset))
	return utils.ConcatKey(utils.ShardAssetAddress, assetBytes, []byte(KEY_OEP4_SHARD_SUPPLY))
}

func genAllowanceKey(asset AssetId, owner, spender common.Address) []byte {
	assetBytes := utils.GetUint64Bytes(uint64(asset))
	return utils.ConcatKey(utils.ShardAssetAddress, assetBytes, []byte(KEY_OEP4_ALLOWANCE), owner[:], spender[:])
}

func genXShardTransferNumKey(asset AssetId, user common.Address) []byte {
	assetBytes := utils.GetUint64Bytes(uint64(asset))
	return utils.ConcatKey(utils.ShardAssetAddress, assetBytes, []byte(KEY_OEP4_TRANSFER_NUM), user[:])
}

func genXShardTransferKey(asset AssetId, user common.Address, transferId *big.Int) []byte {
	assetBytes := utils.GetUint64Bytes(uint64(asset))
	return utils.ConcatKey(utils.ShardAssetAddress, assetBytes, []byte(KEY_OEP4_XSHARD_TRANSFER), user[:],
		common.BigIntToNeoBytes(transferId)[:])
}

func genXShardReceiveKey(asset AssetId, user common.Address, fromShard common.ShardID, transferId *big.Int) []byte {
	assetBytes := utils.GetUint64Bytes(uint64(asset))
	shardIdBytes := utils.GetUint64Bytes(fromShard.ToUint64())
	tranIdBytes := common.BigIntToNeoBytes(transferId)[:]
	return utils.ConcatKey(utils.ShardAssetAddress, assetBytes, []byte(KEY_OEP4_XSHARD_RECEIVE), shardIdBytes, user[:],
		tranIdBytes)
}

func initOep4ShardAsset(native *native.NativeService) {
	sink := common.NewZeroCopySink(0)
	sink.WriteBool(true)
	native.CacheDB.Put(genInitKey(), states.GenRawStorageItem(sink.Bytes()))
}

func isOep4ShardAssetInit(native *native.NativeService) (bool, error) {
	key := genInitKey()
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return false, fmt.Errorf("isOep4ShardAssetInit: read db failed, err: %s", err)
	}
	return len(raw) != 0, nil
}

func setContract(native *native.NativeService, asset AssetId, oep4 *Oep4) {
	sink := common.NewZeroCopySink(0)
	oep4.Serialization(sink)
	native.CacheDB.Put(genAssetKey(asset), states.GenRawStorageItem(sink.Bytes()))
}

func getContract(native *native.NativeService, asset AssetId) (*Oep4, error) {
	raw, err := native.CacheDB.Get(genAssetKey(asset))
	if err != nil {
		return nil, fmt.Errorf("getContract: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("getContract: store is empty")
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getContract: parse store value failed, err: %s", err)
	}
	oep4 := &Oep4{}
	if err := oep4.Deserialization(common.NewZeroCopySource(storeValue)); err != nil {
		return nil, fmt.Errorf("getContract: deserialize failed, err: %s", err)
	}
	return oep4, nil
}

func setXShardTransfer(native *native.NativeService, asset AssetId, user common.Address, transferId *big.Int,
	transfer *XShardTransferState) {
	key := genXShardTransferKey(asset, user, transferId)
	sink := common.NewZeroCopySink(0)
	transfer.Serialization(sink)
	native.CacheDB.Put(key, states.GenRawStorageItem(sink.Bytes()))
}

func getXShardTransfer(native *native.NativeService, asset AssetId, user common.Address,
	transferId *big.Int) (*XShardTransferState, error) {
	key := genXShardTransferKey(asset, user, transferId)
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getXShardTransfer: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("getXShardTransfer: transfer not exist")
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getXShardTransfer: parse store value failed, err: %s", err)
	}
	state := &XShardTransferState{}
	if err := state.Deserialization(common.NewZeroCopySource(storeValue)); err != nil {
		return nil, fmt.Errorf("getXShardTransfer: deserialize failed, err: %s", err)
	}
	return state, nil
}

func setUserBalance(native *native.NativeService, asset AssetId, user common.Address, balance *big.Int) {
	store := common.BigIntToNeoBytes(balance)
	native.CacheDB.Put(genBalanceKey(asset, user), states.GenRawStorageItem(store))
}

func getUserBalance(native *native.NativeService, asset AssetId, user common.Address) (*big.Int, error) {
	raw, err := native.CacheDB.Get(genBalanceKey(asset, user))
	if err != nil {
		return nil, fmt.Errorf("getUserBalance: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return big.NewInt(0), nil
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getUserBalance: parse store value failed, err: %s", err)
	}
	return common.BigIntFromNeoBytes(storeValue), nil
}

func setUserAllowance(native *native.NativeService, asset AssetId, owner, spender common.Address, balance *big.Int) {
	store := common.BigIntToNeoBytes(balance)
	native.CacheDB.Put(genAllowanceKey(asset, owner, spender), states.GenRawStorageItem(store))
}

func getUserAllowance(native *native.NativeService, asset AssetId, owner, spender common.Address) (*big.Int, error) {
	raw, err := native.CacheDB.Get(genAllowanceKey(asset, owner, spender))
	if err != nil {
		return nil, fmt.Errorf("getUserAllowance: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return big.NewInt(0), nil
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getUserAllowance: parse store value failed, err: %s", err)
	}
	return common.BigIntFromNeoBytes(storeValue), nil
}

func setXShardTransferNum(native *native.NativeService, asset AssetId, user common.Address, num *big.Int) {
	key := genXShardTransferNumKey(asset, user)
	native.CacheDB.Put(key, states.GenRawStorageItem(common.BigIntToNeoBytes(num)))
}

func getXShardTransferNum(native *native.NativeService, asset AssetId, user common.Address) (*big.Int, error) {
	key := genXShardTransferNumKey(asset, user)
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getXShardTransferNum: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return big.NewInt(0), nil
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getXShardTransferNum: parse store value failed, err: %s", err)
	}
	return common.BigIntFromNeoBytes(storeValue), nil
}

func receiveTransfer(native *native.NativeService, param *ShardMintParam) {
	key := genXShardReceiveKey(AssetId(param.Asset), param.Account, param.FromShard, param.TransferId)
	sink := common.NewZeroCopySink(0)
	sink.WriteBool(true)
	native.CacheDB.Put(key, states.GenRawStorageItem(sink.Bytes()))
}

func isTransferReceived(native *native.NativeService, param *ShardMintParam) (bool, error) {
	key := genXShardReceiveKey(AssetId(param.Asset), param.Account, param.FromShard, param.TransferId)
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return false, fmt.Errorf("isTransferReceived: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return false, nil
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return false, fmt.Errorf("isTransferReceived: parse store value failed, err: %s", err)
	}
	source := common.NewZeroCopySource(storeValue)
	isReceived, irr, eof := source.NextBool()
	if irr {
		return false, fmt.Errorf("isTransferReceived: deserialize store value, err: %s", common.ErrIrregularData)
	}
	if eof {
		return false, fmt.Errorf("isTransferReceived: deserialize store value, err: %s", io.ErrUnexpectedEOF)
	}
	return isReceived, nil
}

func setAssetNum(native *native.NativeService, num uint64) {
	native.CacheDB.Put(genAssetNumKey(), states.GenRawStorageItem(utils.GetUint64Bytes(num)))
}

func getAssetNum(native *native.NativeService) (uint64, error) {
	key := genAssetNumKey()
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getAssetNum: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return 0, nil
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return 0, fmt.Errorf("getAssetNum: parse store value failed, err: %s", err)
	}
	num, err := utils.GetBytesUint64(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getAssetNum: deserialize store value failed, err: %s", err)
	}
	return num, nil
}

func registerAsset(native *native.NativeService, assetAddr common.Address, assetId AssetId) {
	key := genAssetIdKey(assetAddr)
	native.CacheDB.Put(key, utils.GetUint64Bytes(uint64(assetId)))
}

func getAssetId(native *native.NativeService, assetAddr common.Address) (AssetId, error) {
	key := genAssetIdKey(assetAddr)
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return 0, fmt.Errorf("getAssetId: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return 0, fmt.Errorf("getAssetId: asset not exist")
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return 0, fmt.Errorf("getAssetId: parse store value failed, err: %s", err)
	}
	id, err := utils.GetBytesUint64(storeValue)
	if err != nil {
		return 0, fmt.Errorf("getAssetId: deserialize store value failed, err: %s", err)
	}
	return AssetId(id), nil
}

func isAssetRegister(native *native.NativeService, assetAddr common.Address) (bool, error) {
	key := genAssetIdKey(assetAddr)
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return false, fmt.Errorf("isAssetRegister: read db failed, err: %s", err)
	}
	return len(raw) != 0, nil
}

func deleteAssetId(native *native.NativeService, assetAddr common.Address) {
	key := genAssetIdKey(assetAddr)
	native.CacheDB.Delete(key)
}

func setShardSupplyInfo(native *native.NativeService, asset AssetId, supplyInfo map[common.ShardID]*big.Int) {
	key := genShardSupplyInfoKey(asset)
	sink := common.NewZeroCopySink(0)
	num := uint64(len(supplyInfo))
	sink.WriteUint64(num)
	shards := make([]common.ShardID, 0)
	for shard, _ := range supplyInfo {
		shards = append(shards, shard)
	}
	sort.SliceStable(shards, func(i, j int) bool {
		return shards[i].ToUint64() < shards[j].ToUint64()
	})
	for _, shard := range shards {
		utils.SerializationShardId(sink, shard)
		sink.WriteVarBytes(common.BigIntToNeoBytes(supplyInfo[shard]))
	}
	native.CacheDB.Put(key, states.GenRawStorageItem(sink.Bytes()))
}

func getShardSupplyInfo(native *native.NativeService, asset AssetId) (map[common.ShardID]*big.Int, error) {
	key := genShardSupplyInfoKey(asset)
	raw, err := native.CacheDB.Get(key)
	if err != nil {
		return nil, fmt.Errorf("getShardSupplyInfo: read db failed, err: %s", err)
	}
	if len(raw) == 0 {
		return map[common.ShardID]*big.Int{}, nil
	}
	storeValue, err := states.GetValueFromRawStorageItem(raw)
	if err != nil {
		return nil, fmt.Errorf("getShardSupplyInfo: parse store value failed, err: %s", err)
	}
	souce := common.NewZeroCopySource(storeValue)
	shardNum, eof := souce.NextUint64()
	if eof {
		return nil, fmt.Errorf("getShardSupplyInfo: deserialize shard num failed, err: %s", err)
	}
	shards := make(map[common.ShardID]*big.Int)
	for i := uint64(0); i < shardNum; i++ {
		if shard, err := utils.DeserializationShardId(souce); err != nil {
			return nil, fmt.Errorf("getShardSupplyInfo: deserialize shard failed, index %d, err: %s", i, err)
		} else {
			supplyBytes, _, irr, eof := souce.NextVarBytes()
			if irr {
				return nil, fmt.Errorf("getShardSupplyInfo: deserialize supply failed, index %d, err: %s", i, common.ErrIrregularData)
			}
			if eof {
				return nil, fmt.Errorf("getShardSupplyInfo: deserialize supply failed, index %d, err: %s", i, io.ErrUnexpectedEOF)
			}
			shards[shard] = common.BigIntFromNeoBytes(supplyBytes)
		}
	}
	return shards, nil
}
