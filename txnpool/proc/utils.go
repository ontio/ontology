package proc

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	httpcom "github.com/ontio/ontology/http/base/common"
	params "github.com/ontio/ontology/smartcontract/service/native/global_params"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
)

// getGlobalGasPrice returns a global gas price
func getGlobalGasPrice() (uint64, error) {
	mutable, err := httpcom.NewNativeInvokeTransaction(0, 0, nutils.ParamContractAddress, 0, "getGlobalParam", []interface{}{[]interface{}{"gasPrice"}})
	if err != nil {
		return 0, fmt.Errorf("NewNativeInvokeTransaction error:%s", err)
	}
	tx, err := mutable.IntoImmutable()
	if err != nil {
		return 0, err
	}
	result, err := ledger.DefLedger.PreExecuteContract(tx)
	if err != nil {
		return 0, fmt.Errorf("PreExecuteContract failed %v", err)
	}

	queriedParams := new(params.Params)
	data, err := hex.DecodeString(result.Result.(string))
	if err != nil {
		return 0, fmt.Errorf("decode result error %v", err)
	}

	err = queriedParams.Deserialization(common.NewZeroCopySource([]byte(data)))
	if err != nil {
		return 0, fmt.Errorf("deserialize result error %v", err)
	}
	_, param := queriedParams.GetParam("gasPrice")
	if param.Value == "" {
		return 0, fmt.Errorf("failed to get param for gasPrice")
	}

	gasPrice, err := strconv.ParseUint(param.Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse uint %v", err)
	}

	return gasPrice, nil
}

// getGasPriceConfig returns the bigger one between global and cmd configured
func getGasPriceConfig() uint64 {
	globalGasPrice, err := getGlobalGasPrice()
	if err != nil {
		log.Info(err)
		return 0
	}

	if globalGasPrice < config.DefConfig.Common.GasPrice {
		return config.DefConfig.Common.GasPrice
	}
	return globalGasPrice
}
