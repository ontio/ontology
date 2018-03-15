package native

import (
	"github.com/Ontology/smartcontract/storage"
	scommon "github.com/Ontology/core/store/common"
	"math/big"
	"bytes"
	"github.com/Ontology/smartcontract/service/native/states"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/common"
)

type (
	Handler func(native *NativeService) (bool, error)
)

type NativeService struct {
	CloneCache *storage.CloneCache
	ServiceMap  map[string]Handler
	Notifications []*event.NotifyEventInfo
	Input []byte
	Tx *types.Transaction
}

func NewNativeService(dbCache scommon.IStateStore, input []byte, tx *types.Transaction) *NativeService {
	var nativeService NativeService
	nativeService.CloneCache = storage.NewCloneCache(dbCache)
	nativeService.Input = input
	nativeService.Tx = tx
	nativeService.ServiceMap = make(map[string]Handler)
	nativeService.Register("Common.Token.Transfer", Transfer)
	nativeService.Register("Ont.Token.Init", OntInit)
	return &nativeService
}

func(native *NativeService) Register(methodName string, handler Handler) {
	native.ServiceMap[methodName] = handler
}

func(native *NativeService) IsValid() (bool, error){
	bf := bytes.NewBuffer(native.Input)
	serviceName, err := serialization.ReadVarBytes(bf); if err != nil {
		return false, err
	}
	if _, ok := native.ServiceMap[string(serviceName)]; !ok {
		return false, errors.NewErr("Native does not support this service!")
	}
	native.Input = bf.Bytes()
	return true, nil
}

func Transfer(native *NativeService) (bool, error) {
	transfers := new(states.Transfers)
	if err := transfers.Deserialize(bytes.NewBuffer(native.Input)); err != nil {
		return false, errors.NewDetailErr(err, errors.ErrNoCode, "[Transfer] Transfers deserialize error!")
	}
	for _, p := range transfers.Params {
		for _, s := range p.States {
			if s.Value.Cmp(big.NewInt(0)) < 0 {
				return false, errors.NewErr("[Transfer] transfer amount invalid!")
			}
			if s.From.CompareTo(s.To) == 0 {
				return true, nil
			}
			if !checkWitness(native, s.From) {
				return false, errors.NewErr("[Transfer] Authentication failed!")
			}
			fromKey := append(p.Contract.ToArray(), s.From.ToArray()...)
			fromBalance, err := getBalance(native, fromKey); if err != nil {
				return false, err
			}
			balance := fromBalance.Value.Cmp(s.Value)
			if balance < 0 {
				return false, errors.NewErr("[Transfer] balance insufficient!")
			}
			if balance == 0 {
				native.CloneCache.Delete(scommon.ST_Storage, fromKey)
			} else {
				native.CloneCache.Add(scommon.ST_Storage, fromKey, &states.Amount{Value: new(big.Int).Sub(fromBalance.Value, s.Value)})
			}
			toKey := append(p.Contract.ToArray(), s.To.ToArray()...)
			toBalance, err := getBalance(native, toKey); if err != nil {
				return false, err
			}
			native.CloneCache.Add(scommon.ST_Storage, toKey, &states.Amount{Value: new(big.Int).Add(toBalance.Value, s.Value)})
			native.Notifications = append(native.Notifications, &event.NotifyEventInfo{
				Container: native.Tx.Hash(),
				CodeHash: p.Contract,
				States: []interface{}{s.From, s.To, s.Value},
			})
		}
	}
	return true, nil
}



func getBalance(native *NativeService, key []byte) (*states.Amount, error) {
	balance, err := native.CloneCache.Get(scommon.ST_Storage, key)
	if err != nil {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] storage error!")
	}
	amountState, ok := balance.(*states.Amount); if !ok {
		return nil, errors.NewDetailErr(err, errors.ErrNoCode, "[getBalance] get amount error!")
	}

	return amountState, nil
}

func checkWitness(native *NativeService, u160 common.Uint160) bool {
	addresses := native.Tx.GetSignatureAddresses()
	for _, v := range addresses {
		if v.CompareTo(u160) == 0 {
			return true
		}
	}
	return false
}







