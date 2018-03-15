package native

import (
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"github.com/Ontology/core/genesis"
	ctypes "github.com/Ontology/core/types"
	"math/big"
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/smartcontract/service/native/states"
	"encoding/hex"
	"github.com/Ontology/crypto"
	"github.com/Ontology/common"
)

var (
	owner common.Uint160
	totalSupplyName = []byte("totalSupply")
	decimals = big.NewInt(8)
	totalSupply = new(big.Int).Mul(big.NewInt(1000000000), (new(big.Int).Exp(big.NewInt(10), decimals, nil)))
)

func OntInit(native *NativeService) (bool, error) {
	pubKey, _ := hex.DecodeString("03d3ca3b948fc22e4e99cd348a4351c639f042de87798605d7623fcfb045836ea5")
	pk, _  := crypto.DecodePoint(pubKey)
	owner  = ctypes.AddressFromPubKey(pk)

	amount, err := getBalance(native, getOntTotalSupplyKey())
	if err != nil {
		return false, err
	}
	if amount.Value.Sign() != 0 {
		return false, errors.NewErr("Init has been completed!")
	}
	native.CloneCache.Add(scommon.ST_Storage, getOntOwnerKey(), &states.Amount{Value: totalSupply})
	native.CloneCache.Add(scommon.ST_Storage, getOntTotalSupplyKey(), &states.Amount{Value: totalSupply})
	native.Notifications = append(native.Notifications, &event.NotifyEventInfo{
		Container: native.Tx.Hash(),
		CodeHash: genesis.OntContractAddress,
		States: []interface{}{nil, owner, totalSupply},
	})
	return true, nil
}

func getOntContext() []byte {
	return genesis.OntContractAddress.ToArray()
}

func getOntTotalSupplyKey() []byte {
	return append(getOntContext(), totalSupplyName...)
}

func getOntOwnerKey() []byte {
	return append(getOntContext(), owner.ToArray()...)
}
