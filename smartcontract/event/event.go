package event

import (
	"github.com/Ontology/events"
	. "github.com/Ontology/common"
	"github.com/Ontology/core/ledger/ledgerevent"
)

func PushSmartCodeEvent(txHash Uint256, errcode int64, action string, result interface{}) {
	resp := map[string]interface{}{
		"TxHash": ToHexString(txHash.ToArray()),
		"Action": action,
		"Result": result,
		"Error":  errcode,
	}
	ledgerevent.DefLedgerEvt.Notify(events.EventSmartCode, resp)
}
