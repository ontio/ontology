package ledgerstore

import (
	"bytes"
	"fmt"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/states"
	"github.com/Ontology/core/store/common"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	//"github.com/Ontology/smartcontract"
	"github.com/Ontology/smartcontract/event"
	//"github.com/Ontology/smartcontract/service"
	"sort"
)

const (
	DEPLOY_TRANSACTION = "DeployTransaction"
	INVOKE_TRANSACTION = "InvokeTransaction"
)

func (this *StateStore) HandleBookKeeper(stateBatch *statestore.StateBatch) error {
	bookKeeperState, err := this.GetBookKeeperState()
	if err != nil {
		return fmt.Errorf("GetBookKeeperState error %s", err)
	}
	currBookKeeper := bookKeeperState.CurrBookKeeper
	nextBookKeeper := bookKeeperState.NextBookKeeper
	isChange := false
	if len(currBookKeeper) != len(nextBookKeeper) {
		isChange = true
	}
	for i, bookKeeper := range currBookKeeper {
		next := nextBookKeeper[i]
		if next.X.Cmp(bookKeeper.X) != 0 || next.Y.Cmp(bookKeeper.Y) != 0 {
			isChange = true
			break
		}
	}
	if isChange {
		bookKeeperState.CurrBookKeeper = bookKeeperState.NextBookKeeper
		stateBatch.Change(byte(common.ST_BookKeeper), BookerKeeper, false)
	}
	return nil
}

func (this *StateStore) getPubkeysIndex(pubKey *crypto.PubKey, pubKeyList []*crypto.PubKey) int {
	for index, pk := range pubKeyList {
		if pk.Y.Cmp(pubKey.Y) == 0 && pk.X.Cmp(pubKey.X) == 0 {
			return index
		}
	}
	return -1
}

func (this *StateStore) HandleBookKeeperTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	bookKeeperState, err := this.GetBookKeeperState()
	if err != nil {
		return fmt.Errorf("GetBookKeeperState error %s", err)
	}

	bookKeeper := tx.Payload.(*payload.BookKeeper)
	index := this.getPubkeysIndex(bookKeeper.PubKey, bookKeeperState.NextBookKeeper)
	switch bookKeeper.Action {
	case payload.BookKeeperAction_ADD:
		if index >= 0 {
			return nil
		}
		bookKeeperState.NextBookKeeper = append(bookKeeperState.NextBookKeeper, bookKeeper.PubKey)
		sort.Sort(crypto.PubKeySlice(bookKeeperState.NextBookKeeper))
	case payload.BookKeeperAction_SUB:
		if index < 0 {
			return nil
		}
		bookSize := len(bookKeeperState.NextBookKeeper)
		newNextBookKeeper := make([]*crypto.PubKey, 0, bookSize-1)
		for i := 0; i < bookSize; i++ {
			if i != index {
				newNextBookKeeper = append(newNextBookKeeper, bookKeeperState.NextBookKeeper[i])
			}
		}
		bookKeeperState.NextBookKeeper = newNextBookKeeper
	}

	stateBatch.Change(byte(common.ST_BookKeeper), BookerKeeper, false)
	return nil
}

func (this *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction) error {
	deploy := tx.Payload.(*payload.DeployCode)
	codeHash := deploy.Code.CodeHash()
	if err := stateBatch.TryGetOrAdd(
		common.ST_Contract,
		codeHash.ToArray(),
		&states.ContractState{
			Code:        deploy.Code,
			VmType:      deploy.VmType,
			NeedStorage: deploy.NeedStorage,
			Name:        deploy.Name,
			Version:     deploy.CodeVersion,
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
	//invoke := tx.Payload.(*payload.InvokeCode)
	//txHash := tx.Hash()

	//contrState, err := stateBatch.TryGet(common.ST_Contract, invoke.Code.ToArray())
	//if err != nil {
	//	return fmt.Errorf("TryGet contract error %s", err)
	//}
	//if contrState == nil {
		event.PushSmartCodeEvent(tx.Hash(), 0, INVOKE_TRANSACTION, "Contract not found!")
		//return nil
	//}
	//
	//contract := contrState.Value.(*ContractState)
	//stateMachine := service.NewStateMachine(this, types.Application, block)
	//smc, err := smartcontract.NewSmartContract(
	//	&smartcontract.Context{
	//		VmType:         contract.VmType,
	//		StateMachine:   stateMachine,
	//		SignableData:   tx,
	//		CacheCodeTable: &CacheCodeTable{this},
	//		Input:          invoke.Code,
	//		Code:           contract.Code.Code,
	//		ReturnType:     contract.Code.ReturnType,
	//	})
	//if err != nil {
	//	return fmt.Errorf("NewSmartContract error %s", err)
	//}
	//_, err = smc.InvokeContract()
	//if err != nil {
	//	event.PushSmartCodeEvent(txHash, SMARTCODE_ERROR, INVOKE_TRANSACTION, err)
	//}
	//stateMachine.CloneCache.Commit()
	//notifications := stateMachine.Notifications
	//err = eventStore.SaveEventNotifyByTx(&txHash, notifications)
	//if err != nil {
	//	return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	//}
	//event.PushSmartCodeEvent(txHash, 0, INVOKE_TRANSACTION, err)
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
