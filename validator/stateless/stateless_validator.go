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

package stateless

import (
	"reflect"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/core/validation"
	vatypes "github.com/ontio/ontology/validator/types"
)


type StatelessValidator struct {
	vatypes.ValidatorActor
}

// NewValidator spawns a validator actor and return its pid wraped in Validator
func NewStatelessValidator(id string) (vatypes.Validator, error) {
	validator := &StatelessValidator{}
	validator.Id = id
	props := actor.FromProducer(func() actor.Actor {
		return validator
	})

	pid, err := actor.SpawnNamed(props, id)
	validator.Pid = pid
	return validator, err
}

func (self *StatelessValidator) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		log.Info("stateless-validator: started and be ready to receive txn")
	case *actor.Stopping:
		log.Info("stateless-validator: stopping")
	case *actor.Restarting:
		log.Info("stateless-validator: restarting")
	case *actor.Stopped:
		log.Info("stateless-validator: stopped")
	case *vatypes.VerifyTxReq:
		log.Debugf("stateless-validator receive tx %x", msg.Tx.Hash())
		sender := context.Sender()
		errCode := validation.VerifyTransaction(&msg.Tx)

		response := &vatypes.VerifyTxRsp{
			WorkerId: msg.WorkerId,
			ErrCode:  errCode,
			Hash:     msg.Tx.Hash(),
			VerifyType:     self.VerifyType(),
			Height:   0,
		}

		sender.Tell(response)
	case *vatypes.UnRegisterValidatorRsp:
		context.Self().Stop()
	default:
		log.Info("stateless-validator: unknown msg ", msg, "type", reflect.TypeOf(msg))
	}

}

func (self *StatelessValidator) VerifyType() vatypes.VerifyType {
	return vatypes.Stateless
}

// Register send RegisterValidator message to txpool
func (self *StatelessValidator) Register(poolId *actor.PID) {
	poolId.Tell(&vatypes.RegisterValidatorReq{
		Validator: self.Pid,
		Type:   self.VerifyType(),
		Id:     self.Id,
	})
}

// UnRegister send UnRegisterValidator message to txpool
func (self *StatelessValidator) UnRegister(poolId *actor.PID) {
	poolId.Tell(&vatypes.UnRegisterValidatorReq{
		Id: self.Id,
	})

}
