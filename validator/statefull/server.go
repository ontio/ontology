package statefull

import (
	"fmt"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
	"github.com/Ontology/validator/db"
)

func NewDBVerifier(id tc.VerifyType, db *db.Store) *DBVerifier {
	v := &DBVerifier{
		validatorID: id,
		db:          db,
	}
	return v
}

type DBVerifier struct {
	validatorID tc.VerifyType
	pid         *actor.PID
	db          *db.Store
}

func (v *DBVerifier) verifyTransaction(req *tc.VerifyReq, sender *actor.PID) {
	if sender == nil {
		log.Info("Sender is nil")
		return
	}

	var ok bool = true

	bestBlock, err := v.db.GetBestBlock()
	if err != nil {
		log.Info(fmt.Println("GetBestBlock failed", err))
		ok = false
	}

	exist := v.db.ContainTransaction(req.Txn.Hash())

	if exist {
		log.Info("The transaction already in db")
		ok = false
	}

	rsp := &tc.VerifyRsp{
		WorkerId:    req.WorkerId,
		ValidatorID: uint8(v.validatorID),
		Height:      bestBlock.Height,
		TxnHash:     req.Txn.Hash(),
		Ok:          ok,
	}
	sender.Request(rsp, v.pid)
}

func (v *DBVerifier) Stop() {
	v.pid.Stop()
}

func (v *DBVerifier) GetVerifyType() tc.VerifyType {
	return v.validatorID
}

func (v *DBVerifier) setVerifyType(id tc.VerifyType) {
	v.validatorID = id
}

func (v *DBVerifier) Receive(context actor.Context) {
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
	case *types.Block:
		go v.db.PersistBlock(msg)
	default:
		log.Info("Unknown msg type", msg)
	}
}

func (v *DBVerifier) SetPID(pid *actor.PID) {
	v.pid = pid
}

func (v *DBVerifier) GetPID() *actor.PID {
	return v.pid
}
