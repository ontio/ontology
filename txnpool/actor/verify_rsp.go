package actor

import (
	"reflect"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/validator/types"
	"github.com/ontio/ontology/txnpool/proc"
)

 //VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	server *proc.TXPoolServer
}

//NewVerifyRspActor creates an actor to handle the verified result from validators
func NewVerifyRspActor(svr *proc.TXPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{}
	a.setServer(svr)
	return a
}

// Receive implements the actor interface
func (self *VerifyRspActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("txpool-verify actor: started and be ready to receive validator's msg")

	case *actor.Stopping:
		log.Warn("txpool-verify actor: stopping")

	case *actor.Restarting:
		log.Warn("txpool-verify actor: Restarting")

	case *types.RegisterValidator:
		log.Debugf("txpool-verify actor:: validator %v connected", msg.Sender)
		self.server.RegisterValidator(msg)

	case *types.UnRegisterValidator:
		log.Debugf("txpool-verify actor:: validator %d:%v disconnected", msg.Type, msg.Id)

		self.server.UnRegisterValidator(msg.Type, msg.Id)

	case *types.CheckTxRsp:
		log.Debug("txpool-verify actor:: Receives verify rsp message")

		self.server.AssignValidateRspToWorker(msg)

	default:
		log.Warn("txpool-verify actor:Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *VerifyRspActor) setServer(s *proc.TXPoolServer) {
	self.server = s
}
