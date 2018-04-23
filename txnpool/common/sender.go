package common

import (
	"github.com/ontio/ontology-eventbus/actor"
	ttypes "github.com/ontio/ontology/txnpool/types"
	ctypes "github.com/ontio/ontology/core/types"
	vtypes "github.com/ontio/ontology/validator/types"
	"github.com/ontio/ontology/common/log"
	"sync"
)

type validator struct {
	Validator *actor.PID
	Type      vtypes.VerifyType
	Id        string
}
type validators struct {
	sync.RWMutex
	entries    map[vtypes.VerifyType][]*validator // Registered validator container
	robinState map[vtypes.VerifyType]int          // Keep the round robin index for each verify type
}

type Sender struct {
	// The registered validators
	validatorActor *validators
	txPoolActor    *actor.PID
	verifyRspActor *actor.PID
	netActor       *actor.PID
}

func NewSender() *Sender {
	send := &Sender{
		validatorActor: &validators{
			entries:    make(map[vtypes.VerifyType][]*validator),
			robinState: make(map[vtypes.VerifyType]int),
		},
	}
	return send
}

// RegisterActor registers an actor with the actor type and pid.
func (self *Sender) RegisterActor(tpe ttypes.ActorType, pid *actor.PID) {
	if tpe == ttypes.TxPoolActor {
		self.txPoolActor = pid
	} else if tpe == ttypes.VerifyRspActor {
		self.verifyRspActor = pid
	} else if tpe == ttypes.NetActor {
		self.netActor = pid
	}
}

// UnRegisterActor cancels the actor with the actor type.
func (self *Sender) UnRegisterActor(tpe ttypes.ActorType) {
	if tpe == ttypes.TxPoolActor {
		self.txPoolActor = nil
	} else if tpe == ttypes.VerifyRspActor {
		self.verifyRspActor = nil
	} else if tpe == ttypes.NetActor {
		self.netActor = nil
	}
}

func (self *Sender) SendTxToNetActor(tx *ctypes.Transaction) {
	if self.netActor != nil {
		self.netActor.Tell(tx)
	}
}

// GetPID returns an actor pid with the actor type, If the type
// doesn't exist, return nil.
func (self *Sender) GetPid(tpe ttypes.ActorType) *actor.PID {
	if tpe == ttypes.TxPoolActor {
		return self.txPoolActor
	} else if tpe == ttypes.VerifyRspActor {
		return self.verifyRspActor
	} else if tpe == ttypes.NetActor {
		return self.netActor
	}
	return nil
}
func (self *Sender) SendVerifyTxReq(verifyType vtypes.VerifyType, req *vtypes.VerifyTxReq) bool {
	if verifyType == vtypes.Statefull {
		return self.sendVerifyStatefulTxReq(req)
	} else if verifyType == vtypes.Stateless {
		return self.sendVerifyStatelessTxReq(req)
	}
	return false
}

// sendReq2Validator sends a check request to the validators
func (self *Sender) sendVerifyStatelessTxReq(req *vtypes.VerifyTxReq) bool {
	rspPid := self.verifyRspActor
	if rspPid == nil {
		log.Info("VerifyRspActor not exist")
		return false
	}

	pids := self.getNextValidator()
	if pids == nil {
		return false
	}
	for _, pid := range pids {
		pid.Request(req, rspPid)
	}

	return true
}

// sendReq2StatefulV sends a check request to the stateful validator
func (self *Sender) sendVerifyStatefulTxReq(req *vtypes.VerifyTxReq) bool {
	rspPid := self.verifyRspActor
	if rspPid == nil {
		log.Info("VerifyRspActor not exist")
		return false
	}

	pid := self.getNextValidatorByType(vtypes.Statefull)
	log.Info("worker send tx to the stateful")
	if pid == nil {
		return false
	}

	pid.Request(req, rspPid)
	return true
}

// getNextValidatorPIDs returns the next pids to verify the transaction using
// roundRobin LB.
//return two stateful and stateless validoter
func (self *Sender) getNextValidator() []*actor.PID {
	self.validatorActor.Lock()
	defer self.validatorActor.Unlock()

	if len(self.validatorActor.entries) == 0 {
		return nil
	}

	pids := make([]*actor.PID, 0, len(self.validatorActor.entries))
	for k, v := range self.validatorActor.entries {
		preIndex := self.validatorActor.robinState[k]
		nextIndex := (preIndex + 1) % len(v)
		self.validatorActor.robinState[k] = nextIndex
		pids = append(pids, v[nextIndex].Validator)
	}
	return pids
}

// getNextValidatorPID returns the next pid with the verify type using roundRobin LB
func (self *Sender) getNextValidatorByType(key vtypes.VerifyType) *actor.PID {
	self.validatorActor.Lock()
	defer self.validatorActor.Unlock()

	length := len(self.validatorActor.entries[key])
	if length == 0 {
		return nil
	}

	entries := self.validatorActor.entries[key]
	preIndex := self.validatorActor.robinState[key]
	nextIndex := (preIndex + 1) % length
	self.validatorActor.robinState[key] = nextIndex
	return entries[nextIndex].Validator
}

// registerValidator registers a validator to verify a transaction.
func (self *Sender) RegisterValidator(pid *actor.PID, tpe vtypes.VerifyType, id string) {
	self.validatorActor.Lock()
	defer self.validatorActor.Unlock()

	_, ok := self.validatorActor.entries[tpe]

	if !ok {
		self.validatorActor.entries[tpe] = make([]*validator, 0, 1)
	}
	self.validatorActor.entries[tpe] = append(self.validatorActor.entries[tpe], &validator{pid, tpe, id})
}

// unRegisterValidator cancels a validator with the verify type and id.
func (self *Sender) UnRegisterValidator(verifyType vtypes.VerifyType,
	id string) {

	self.validatorActor.Lock()
	defer self.validatorActor.Unlock()

	tmpSlice, ok := self.validatorActor.entries[verifyType]
	if !ok {
		log.Error("No validator on check type:%d\n", verifyType)
		return
	}

	for i, v := range tmpSlice {
		if v.Id == id {
			self.validatorActor.entries[verifyType] =
				append(tmpSlice[0:i], tmpSlice[i+1:]...)
			if v.Validator != nil {
				v.Validator.Tell(&vtypes.UnRegisterValidatorRsp{Id: id, VerifyType: verifyType})
			}
			if len(self.validatorActor.entries[verifyType]) == 0 {
				delete(self.validatorActor.entries, verifyType)
			}
		}
	}
}

func (self *Sender) Stop() {
	self.netActor.Stop()
	self.txPoolActor.Stop()
	self.verifyRspActor.Stop()
	self.netActor.Stop()
}
