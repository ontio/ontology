package native

import (
	"github.com/Ontology/smartcontract/storage"
	scommon "github.com/Ontology/core/store/common"
	"bytes"
	"github.com/Ontology/errors"
	"github.com/Ontology/common/serialization"
	"github.com/Ontology/core/types"
	"github.com/Ontology/smartcontract/event"
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
	nativeService.Register("Token.Common.Transfer", Transfer)
	nativeService.Register("Token.Ont.Init", OntInit)
	return &nativeService
}

func(native *NativeService) Register(methodName string, handler Handler) {
	native.ServiceMap[methodName] = handler
}

func(native *NativeService) Invoke() (bool, error){
	bf := bytes.NewBuffer(native.Input)
	serviceName, err := serialization.ReadVarBytes(bf); if err != nil {
		return false, err
	}
	service, ok := native.ServiceMap[string(serviceName)]; if !ok {
		return false, errors.NewErr("Native does not support this service!")
	}
	native.Input = bf.Bytes()
	return service(native)
}








