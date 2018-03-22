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

package ledgerstore

import (
	"bytes"
	"fmt"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/types"
	"github.com/Ontology/smartcontract/event"
	"github.com/Ontology/smartcontract/service/native"
	vmtypes "github.com/Ontology/vm/types"
)

const (
	DEPLOY_TRANSACTION = "DeployTransaction"
	INVOKE_TRANSACTION = "InvokeTransaction"
)

func (this *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	deploy := tx.Payload.(*payload.DeployCode)
	code := &vmtypes.VmCode{
		Code:   deploy.Code,
		VmType: deploy.VmType,
	}
	codeHash := code.AddressFromVmCode()
	if err := stateBatch.TryGetOrAdd(
		common.ST_Contract,
		codeHash[:],
		&payload.DeployCode{
			Code:        deploy.Code,
			VmType:      deploy.VmType,
			NeedStorage: deploy.NeedStorage,
			Name:        deploy.Name,
			Version:     deploy.Version,
			Author:      deploy.Author,
			Email:       deploy.Email,
			Description: deploy.Description,
		},
		false); err != nil {
		return fmt.Errorf("TryGetOrAdd contract error %s", err)
	}
	return nil
}

func (this *StateStore) HandleInvokeTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction, block *types.Block, eventStore common.IEventStore) error {
	invoke := tx.Payload.(*payload.InvokeCode)
	txHash := tx.Hash()
	switch invoke.Code.VmType {
	case vmtypes.NativeVM:
		na := native.NewNativeService(stateBatch, invoke.Code.Code, tx)
		if ok, err := na.Invoke(); !ok {
			log.Error("Native contract execute error:", err)
			event.PushSmartCodeEvent(txHash, 0, INVOKE_TRANSACTION, err)
		}
		na.CloneCache.Commit()
	case vmtypes.NEOVM:
	//TODO
	case vmtypes.WASMVM:
		//TODO
	}
	return nil
}

func (this *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	//TODO
	return nil
}

func (this *StateStore) HandleEnrollmentTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	en := tx.Payload.(*payload.Enrollment)
	bf := new(bytes.Buffer)
	if err := en.PublicKey.Serialize(bf); err != nil {
		return err
	}
	stateBatch.TryAdd(common.ST_Validator, bf.Bytes(), &states.ValidatorState{PublicKey: en.PublicKey}, false)
	return nil
}

func (this *StateStore) HandleVoteTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	vote := tx.Payload.(*payload.Vote)
	buf := new(bytes.Buffer)
	vote.Account.Serialize(buf)
	stateBatch.TryAdd(common.ST_Vote, buf.Bytes(), &states.VoteState{PublicKeys: vote.PubKeys}, false)
	return nil
}
