package event

import (
	"github.com/Ontology/events"
	"github.com/Ontology/core/ledger"
	. "github.com/Ontology/common"
)

func PushSmartCodeEvent(txHash Uint256, errcode int64, action string, result interface{}) {
	resp := map[string]interface{}{
		"TxHash": ToHexString(txHash.ToArrayReverse()),
		"Action": action,
		"Result": result,
		"Error":  errcode,
	}
	ledger.DefaultLedger.Blockchain.BCEvents.Notify(events.EventSmartCode, resp)
}
