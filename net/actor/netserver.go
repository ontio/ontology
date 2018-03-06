package actor

import (
	"github.com/ONTID/eventbus/actor"
	"github.com/Ontology/net/protocol"
	"github.com/Ontology/common/log"
)

var NetServerPid *actor.PID
var node protocol.Noder
type MsgActor struct{}

type GetConnectionCntReq struct {
}
type GetConnectionCntRsp struct {
	Cnt uint
}

func (state *MsgActor) Receive(context actor.Context) {
	switch context.Message().(type) {
	case *GetConnectionCntReq:
		connectionCnt := node.GetConnectionCnt()
		context.Sender().Request(&GetConnectionCntRsp{Cnt: connectionCnt}, context.Self())
	default:
		err := node.Xmit(context.Message())
		if nil != err {
			log.Error("Error Xmit message ", err.Error())
		}
	}
}

func init() {
	props := actor.FromProducer(func() actor.Actor { return &MsgActor{} })
	NetServerPid = actor.Spawn(props)
}

func SetNode(netNode protocol.Noder){
	node = netNode
}
