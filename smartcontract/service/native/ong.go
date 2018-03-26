package native

import (
	scommon "github.com/Ontology/core/store/common"
	"github.com/Ontology/errors"
	"math/big"
	"github.com/Ontology/smartcontract/service/native/states"
	cstates "github.com/Ontology/core/states"
	"bytes"
	"github.com/Ontology/core/genesis"
	"github.com/Ontology/smartcontract/event"
)

var (
	totalSupplyName = []byte("totalSupply")
	decimals = big.NewInt(8)
	totalSupply = new(big.Int).Mul(big.NewInt(1000000000), (new(big.Int).Exp(big.NewInt(10), decimals, nil)))
)

func OngInit(native *NativeService) error {
	contract := native.ContextRef.CurrentContext().ContractAddress
	amount, err := getStorageBigInt(native, getTotalSupplyKey(contract))
	if err != nil {
		return err
	}

	if amount != nil && amount.Sign() != 0 {
		return errors.NewErr("Init ong has been completed!")
	}

	native.CloneCache.Add(scommon.ST_Storage, append(contract[:], getOntContext()...), &cstates.StorageItem{Value: totalSupply.Bytes()})
	return nil
}

func OngTransfer(native *NativeService) error {
	transfers := new(states.Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OngTransfer] Transfers deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	for _, v := range transfers.States {
		if err := transfer(native, contract, v); err != nil {
			return err
		}
		native.Notifications = append(native.Notifications,
			&event.NotifyEventInfo{
				Container: native.Tx.Hash(),
				CodeHash: contract,
				States: []interface{}{v.From, v.To, v.Value},
			})
	}
	return nil
}

func OngApprove(native *NativeService) error {
	state := new(states.State)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OngApprove] state deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	native.CloneCache.Add(scommon.ST_Storage, getApproveKey(contract, state), &cstates.StorageItem{Value: state.Value.Bytes()})
	return nil
}

func OngTransferFrom(native *NativeService) error {
	state := new(states.TransferFrom)
	if err := state.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return errors.NewDetailErr(err, errors.ErrNoCode, "[OntTransferFrom] State deserialize error!")
	}
	contract := native.ContextRef.CurrentContext().ContractAddress
	if err := transferFrom(native, contract, state); err != nil {
		return err
	}
	native.Notifications = append(native.Notifications,
		&event.NotifyEventInfo{
			Container: native.Tx.Hash(),
			CodeHash: contract,
			States: []interface{}{state.From, state.To, state.Value},
		})
	return nil
}

func getOntContext() []byte {
	return genesis.OntContractAddress[:]
}


