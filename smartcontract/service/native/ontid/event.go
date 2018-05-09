package ontid

import (
	"encoding/hex"

	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func newEvent(srvc *native.NativeService, st interface{}) {
	e := event.NotifyEventInfo{}
	e.TxHash = srvc.Tx.Hash()
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}

func triggerRegisterEvent(srvc *native.NativeService, id []byte) {
	newEvent(srvc, []string{"Register", hex.EncodeToString(id)})
}

func triggerPublicEvent(srvc *native.NativeService, op string, id, pub []byte) {
	st := []string{"PublicKey", op, hex.EncodeToString(id), hex.EncodeToString(pub)}
	newEvent(srvc, st)
}

func triggerAttributeEvent(srvc *native.NativeService, op string, id, path []byte) {
	st := []string{"Attribute", op, hex.EncodeToString(id), string(path)}
	newEvent(srvc, st)
}

func triggerRecoveryEvent(srvc *native.NativeService, op string, id, addr []byte) {
	st := []string{"Recovery", op, hex.EncodeToString(id), hex.EncodeToString(addr)}
	newEvent(srvc, st)
}
