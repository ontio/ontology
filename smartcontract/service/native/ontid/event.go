package ontid

import (
	"encoding/hex"

	"github.com/ontio/ontology/smartcontract/event"
	"github.com/ontio/ontology/smartcontract/service/native"
)

func newEvent(srvc *native.NativeService, st interface{}) {
	e := event.NotifyEventInfo{}
	e.ContractAddress = srvc.ContextRef.CurrentContext().ContractAddress
	e.States = st
	srvc.Notifications = append(srvc.Notifications, &e)
	return
}

func triggerRegisterEvent(srvc *native.NativeService, id []byte) {
	newEvent(srvc, []string{"Register", string(id)})
}

func triggerPublicEvent(srvc *native.NativeService, op string, id, pub []byte, keyID uint32) {
	st := []interface{}{"PublicKey", op, string(id), keyID, hex.EncodeToString(pub)}
	newEvent(srvc, st)
}

func triggerAttributeEvent(srvc *native.NativeService, op string, id, path []byte) {
	st := []string{"Attribute", op, string(id), string(path)}
	newEvent(srvc, st)
}

func triggerRecoveryEvent(srvc *native.NativeService, op string, id, addr []byte) {
	st := []string{"Recovery", op, string(id), hex.EncodeToString(addr)}
	newEvent(srvc, st)
}
