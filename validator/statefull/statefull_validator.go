/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package statefull

import (
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/errors"
	"github.com/ontio/ontology/validator/db"
	vatypes "github.com/ontio/ontology/validator/types"
)
type StatefulValidator struct {
	vatypes.ValidatorActor
	bestBlock db.BestBlock
}
// NewValidator returns Validator for stateful check of tx
func NewStatefulValidator(id string) (vatypes.Validator, error) {

	validator := &StatefulValidator{}
	validator.Id = id
	props := actor.FromProducer(func() actor.Actor {
		return validator
	})

	pid, err := actor.SpawnNamed(props, id)
	validator.Pid = pid
	return validator, err
}

func (self *StatefulValidator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("statefull-validator: started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("statefull-validator: stopping")
	case *actor.Restarting:
		log.Info("statefull-validator: restarting")
	case *vatypes.VerifyTxReq:
		log.Debugf("statefull-validator: receive tx %x", msg.Tx.Hash())
		sender := context.Sender()
		height := ledger.DefLedger.GetCurrentBlockHeight()

		errCode := errors.ErrNoError
		hash := msg.Tx.Hash()

		exist, err := ledger.DefLedger.IsContainTransaction(hash)
		if err != nil {
			log.Warn("query db error:", err)
			errCode = errors.ErrUnknown
		} else if exist {
			errCode = errors.ErrDuplicatedTx
		}

		response := &vatypes.VerifyTxRsp{
			WorkerId: msg.WorkerId,
			Type:     self.VerifyType(),
			Hash:     msg.Tx.Hash(),
			Height:   height,
			ErrCode:  errCode,
		}

		sender.Tell(response)
	case *vatypes.UnRegisterValidatorRsp:
		context.Self().Stop()
	case *types.Block:

		//bestBlock, _ := self.db.GetBestBlock()
		//if bestBlock.Height+1 < msg.Header.Height {
		//	// add sync block request
		//} else if bestBlock.Height+1 == msg.Header.Height {
		//	self.db.PersistBlock(msg)
		//}

	default:
		log.Info("statefull-validator: unknown msg ", msg, "type", reflect.TypeOf(msg))
	}

}

func (self *StatefulValidator) VerifyType() vatypes.VerifyType {
	return vatypes.Statefull
}

func (self *StatefulValidator) Register(poolId *actor.PID) {
	poolId.Tell(&vatypes.RegisterValidatorReq{
		Validator: self.Pid,
		Type:   self.VerifyType(),
		Id:     self.Id,
	})
}

func (self *StatefulValidator) UnRegister(poolId *actor.PID) {
	poolId.Tell(&vatypes.UnRegisterValidatorReq{
		Id: self.Id,
	})
}
