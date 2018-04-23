package actor

import (
	"reflect"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology-eventbus/actor"
	vtypes "github.com/ontio/ontology/validator/types"
	"github.com/ontio/ontology/txnpool/proc"
	tcomn "github.com/ontio/ontology/txnpool/common"
)

//VerifyRspActor: Handle the response from the validators
type VerifyRspActor struct {
	txPoolServer *proc.TxPoolServer
	sender       *tcomn.Sender
}

//NewVerifyRspActor creates an actor to handle the verified result from validators
func NewVerifyRspActor(sender *tcomn.Sender, svr *proc.TxPoolServer) *VerifyRspActor {
	a := &VerifyRspActor{sender: sender}
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

	case *vtypes.RegisterValidatorReq:
		log.Debugf("txpool-verify actor:: validator %v connected", msg.Validator)
		self.sender.RegisterValidator(msg.Validator, msg.Type, msg.Id)

	case *vtypes.UnRegisterValidatorReq:
		log.Debugf("txpool-verify actor:: validator %d:%v disconnected", msg.VerifyType, msg.Id)

		self.sender.UnRegisterValidator(msg.VerifyType, msg.Id)

	case *vtypes.VerifyTxRsp:
		log.Debug("txpool-verify actor:: Receives verify rsp message")

		self.txPoolServer.PutVerifyTxRsp(msg)

	default:
		log.Warn("txpool-verify actor:Unknown msg ", msg, "type", reflect.TypeOf(msg))
	}
}

func (self *VerifyRspActor) setServer(s *proc.TxPoolServer) {
	self.txPoolServer = s
}
