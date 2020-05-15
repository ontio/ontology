package ontfs

import (
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

type SpaceInfo struct {
	SpaceOwner  common.Address
	Volume      uint64
	RestVol     uint64
	CopyNumber  uint64
	PayAmount   uint64
	RestAmount  uint64
	PdpInterval uint64
	TimeStart   uint64
	TimeExpired uint64
	ValidFlag   bool
}

func (this *SpaceInfo) Serialization(sink *common.ZeroCopySink) {
	utils.EncodeAddress(sink, this.SpaceOwner)
	utils.EncodeVarUint(sink, this.Volume)
	utils.EncodeVarUint(sink, this.RestVol)
	utils.EncodeVarUint(sink, this.CopyNumber)
	utils.EncodeVarUint(sink, this.PayAmount)
	utils.EncodeVarUint(sink, this.RestAmount)
	utils.EncodeVarUint(sink, this.PdpInterval)
	utils.EncodeVarUint(sink, this.TimeStart)
	utils.EncodeVarUint(sink, this.TimeExpired)
	sink.WriteBool(this.ValidFlag)
}

func (this *SpaceInfo) Deserialization(source *common.ZeroCopySource) error {
	var err error
	this.SpaceOwner, err = utils.DecodeAddress(source)
	if err != nil {
		return err
	}
	this.Volume, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RestVol, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.CopyNumber, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PayAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.RestAmount, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.PdpInterval, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.TimeStart, err = utils.DecodeVarUint(source)
	if err != nil {
		return nil
	}
	this.TimeExpired, err = utils.DecodeVarUint(source)
	if err != nil {
		return err
	}
	this.ValidFlag, err = DecodeBool(source)
	if err != nil {
		return err
	}
	return nil
}

func addSpaceInfo(native *native.NativeService, spaceInfo *SpaceInfo) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	spaceInfoKey := GenFsSpaceKey(contract, spaceInfo.SpaceOwner)

	sink := common.NewZeroCopySink(nil)
	spaceInfo.Serialization(sink)

	utils.PutBytes(native, spaceInfoKey, sink.Bytes())
}

func delSpaceInfo(native *native.NativeService, spaceOwner common.Address) {
	contract := native.ContextRef.CurrentContext().ContractAddress
	spaceInfoKey := GenFsSpaceKey(contract, spaceOwner)
	native.CacheDB.Delete(spaceInfoKey)
}

func spaceInfoExist(native *native.NativeService, spaceOwner common.Address) bool {
	contract := native.ContextRef.CurrentContext().ContractAddress
	spaceInfoKey := GenFsSpaceKey(contract, spaceOwner)

	item, err := utils.GetStorageItem(native, spaceInfoKey)
	if err != nil || item == nil || item.Value == nil {
		return false
	}
	return true
}

func getSpaceInfoFromDb(native *native.NativeService, fileOwner common.Address) *SpaceInfo {
	contract := native.ContextRef.CurrentContext().ContractAddress
	spaceInfoKey := GenFsSpaceKey(contract, fileOwner)

	item, err := utils.GetStorageItem(native, spaceInfoKey)
	if err != nil || item == nil || item.Value == nil {
		return nil
	}

	var spaceInfo SpaceInfo
	source := common.NewZeroCopySource(item.Value)
	if err := spaceInfo.Deserialization(source); err != nil {
		return nil
	}
	return &spaceInfo
}

func getSpaceRawRealInfo(native *native.NativeService, fileOwner common.Address) []byte {
	spaceInfo := getSpaceInfoFromDb(native, fileOwner)
	if spaceInfo == nil {
		return nil
	}

	if uint64(native.Time) > spaceInfo.TimeExpired {
		spaceInfo.ValidFlag = false
	}

	sink := common.NewZeroCopySink(nil)
	spaceInfo.Serialization(sink)
	return sink.Bytes()
}

func getAndUpdateSpaceInfo(native *native.NativeService, fileOwner common.Address) *SpaceInfo {
	spaceInfo := getSpaceInfoFromDb(native, fileOwner)
	if spaceInfo == nil {
		return nil
	}

	if uint64(native.Time) > spaceInfo.TimeExpired {
		spaceInfo.ValidFlag = false
		addSpaceInfo(native, spaceInfo)
	}

	return spaceInfo
}
