package shardasset

import (
	"github.com/ontio/ontology/smartcontract/service/native"
	"github.com/ontio/ontology/smartcontract/service/native/shardasset/oep4"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
)

func InitShardAsset() {
	native.Contracts[utils.ShardAssetAddress] = RegisterShardAsset
}

func RegisterShardAsset(native *native.NativeService) {
	oep4.RegisterOEP4(native)
	// TODO: support oep5, oep8
}
