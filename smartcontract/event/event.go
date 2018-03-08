package event

import (
	. "github.com/Ontology/common"
	"github.com/Ontology/events"
	"github.com/Ontology/events/message"
	"github.com/Ontology/core/types"
)

func PushSmartCodeEvent(txHash Uint256, errcode int64, action string, result interface{}) {
	smartCodeEvt := &types.SmartCodeEvent{
		TxHash: ToHexString(txHash.ToArray()),
		Action: action,
		Result: result,
		Error:  errcode,
	}
	events.DefActorPublisher.Publish(message.TopicSmartCodeEvent, smartCodeEvt)
}
