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
	txPoolServer *proc.TxPoolServer
}

//NewVerifyRspActor creates an actor to handle the verified result from validators
func NewVerifyRspActor(svr *proc.TxPoolServer) *VerifyRspActor {
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

	case *types.RegisterValidatorReq:
		log.Debugf("txpool-verify actor:: validator %v connected", msg.Validator)
		self.txPoolServer.RegisterValidator(msg.Validator,msg.Type,msg.Id)

	case *types.UnRegisterValidatorReq:
		log.Debugf("txpool-verify actor:: validator %d:%v disconnected", msg.Type, msg.Id)

		self.txPoolServer.UnRegisterValidator(msg.Type, msg.Id)

	case *types.VerifyTxRsp:
		log.Debug("txpool-verify actor:: Receives verify rsp message")

		self.txPoolServer.AssignRspToWorker(msg)

	default:
		log.Warn("txpool-verify actor:Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *VerifyRspActor) setServer(s *proc.TxPoolServer) {
	self.txPoolServer = s
}
