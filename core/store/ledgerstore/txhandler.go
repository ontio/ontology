package ledgerstore

import (
	"bytes"
	"fmt"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/types"
	"github.com/Ontology/smartcontract/event"
	vmtypes "github.com/Ontology/vm/types"
	"github.com/Ontology/smartcontract/service/native"
	"github.com/Ontology/common/log"
)

const (
	DEPLOY_TRANSACTION = "DeployTransaction"
	INVOKE_TRANSACTION = "InvokeTransaction"
)

func (this *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	deploy := tx.Payload.(*payload.DeployCode)
	code := &vmtypes.VmCode{
			Code: deploy.Code,
			VmType: deploy.VmType,
		}
	codeHash := code.AddressFromVmCode()
	if err := stateBatch.TryGetOrAdd(
		common.ST_Contract,
		codeHash.ToArray(),
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
	//var notifies []*event.NotifyEventInfo
	switch invoke.Code.VmType {
	case vmtypes.NativeVM:
		na := native.NewNativeService(stateBatch, invoke.Code.Code, tx)
		if ok, err := na.Invoke(); !ok {
			log.Error("Native contract execute error:", err)
			event.PushSmartCodeEvent(txHash, 0, INVOKE_TRANSACTION, err)
		}
		na.CloneCache.Commit()
		//notifies = na.Notifications
	case vmtypes.NEOVM:
		//stateMachine := service.NewStateMachine(this, stateBatch, vmtypes.Application, block.Header.Timestamp)
		//e := neovm.NewExecutionEngine(
		//	tx,
		//	new(neovm.ECDsaCrypto),
		//	&CacheCodeTable{this},
		//	stateMachine,
		//)
	case vmtypes.WASMVM:

	}
	//if err := eventStore.SaveEventNotifyByTx(&txHash, notifies); err != nil {
	//	return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	//}
	return nil
}

func (this *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	//TODO
	//p := tx.Payload.(*payload.Claim)
	//for _, c := range p.Claims {
	//	state, err := this.TryGetAndChange(ST_SpentCoin, c.ReferTxID.ToArray(), false)
	//	if err != nil {
	//		return fmt.Errorf("TryGetAndChange error %s", err)
	//	}
	//	spentcoins := state.(*SpentCoinState)
	//
	//	newItems := make([]*Item, 0, len(spentcoins.Items)-1)
	//	for index, item := range spentcoins.Items {
	//		if uint16(index) != c.ReferTxOutputIndex {
	//			newItems = append(newItems, item)
	//		}
	//	}
	//	spentcoins.Items = newItems
	//}
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




