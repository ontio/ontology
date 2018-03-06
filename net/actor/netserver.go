package actor

import (
	"github.com/ONTID/eventbus/actor"
	"github.com/Ontology/net/protocol"
	"github.com/Ontology/common/log"
)

var NetServerPid *actor.PID
var node protocol.Noder
type MsgActor struct{}

func (state *MsgActor) Receive(context actor.Context) {
	err := node.Xmit(context.Message())
	if nil != err {
		log.Error("Error Xmit message ", err.Error())
	}
}

func init() {
	props := actor.FromProducer(func() actor.Actor { return &MsgActor{} })
	NetServerPid = actor.Spawn(props)
}

func SetNode(netNode protocol.Noder){
	node = netNode
}