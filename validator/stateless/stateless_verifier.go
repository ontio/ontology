package stateless

import (
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/validation"
	"github.com/Ontology/errors"
	"github.com/Ontology/eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
)

func NewVerifier(id tc.VerifyType) *StatelessVerifier {
	v := &StatelessVerifier{}
	v.init(id)
	return v
}

type StatelessVerifier struct {
	validatorID tc.VerifyType
	pid         *actor.PID
}

func (v *StatelessVerifier) init(id tc.VerifyType) {
	v.setVerifyType(id)
}

func (v *StatelessVerifier) checkAttributeProgram(tx *types.Transaction) error {
	//TODO: implement CheckAttributeProgram
	return nil
}

func (v *StatelessVerifier) verifyTransaction(req *tc.VerifyReq, sender *actor.PID) {
	if sender == nil {
		log.Info("Sender is nil")
		return
	}

	var ok bool = true
	if err := v.checkAttributeProgram(req.Txn); err != nil {
		log.Warn("[VerifyTransaction],", err)
		ok = false
	}

	if err := validation.VerifyTransaction(req.Txn); err != errors.ErrNoError {
		log.Warn("[VerifyTransaction],", err)
		ok = false
	}
	// Todo:
	rsp := &tc.VerifyRsp{
		WorkerId:    req.WorkerId,
		ValidatorID: uint8(v.validatorID),
		Height:      8,
		TxnHash:     req.Txn.Hash(),
		Ok:          ok,
	}
	sender.Request(rsp, v.pid)
}

func (v *StatelessVerifier) Stop() {
	v.pid.Stop()
}

func (v *StatelessVerifier) GetVerifyType() tc.VerifyType {
	return v.validatorID
}

func (v *StatelessVerifier) setVerifyType(id tc.VerifyType) {
	v.validatorID = id
}

func (v *StatelessVerifier) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("Server started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("Server stopping")
	case *actor.Restarting:
		log.Info("Server Restarting")
	case *tc.VerifyReq:
		sender := context.Sender()
		go v.verifyTransaction(msg, sender)
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (v *StatelessVerifier) SetPID(pid *actor.PID) {
	v.pid = pid
}

func (v *StatelessVerifier) GetPID() *actor.PID {
	return v.pid
}
