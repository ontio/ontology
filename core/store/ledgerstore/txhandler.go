package ledgerstore

import (
	//"bytes"
	//"fmt"
	"github.com/Ontology/core/store/common"
	//"sort"
	"github.com/Ontology/core/types"
	"github.com/Ontology/core/store/statestore"
	"github.com/Ontology/crypto"
	//"github.com/Ontology/core/transaction/payload"
	//"github.com/Ontology/core/states"
)

func (this *StateStore) HandleBookKeeper(stateBatch *statestore.StateBatch) error {
	//bookKeeperState, err := this.GetBookKeeperState()
	//if err != nil {
	//	return fmt.Errorf("GetBookKeeperState error %s", err)
	//}
	//currBookKeeper := bookKeeperState.CurrBookKeeper
	//nextBookKeeper := bookKeeperState.NextBookKeeper
	//isChange := false
	//if len(currBookKeeper) != len(nextBookKeeper) {
	//	isChange = true
	//}
	//for i, bookKeeper := range currBookKeeper {
	//	next := nextBookKeeper[i]
	//	if next.X.Cmp(bookKeeper.X) != 0 || next.Y.Cmp(bookKeeper.Y) != 0 {
	//		isChange = true
	//		break
	//	}
	//}
	//if isChange {
	//	bookKeeperState.CurrBookKeeper = bookKeeperState.NextBookKeeper
	//	stateBatch.Change(byte(common.ST_BookKeeper), BookerKeeper, false)
	//}
	return nil
}

func (this *StateStore) getPubkeysIndex(pubKey *crypto.PubKey, pubKeyList []*crypto.PubKey) int {
	//for index, pk := range pubKeyList {
	//	if pk.Y.Cmp(pubKey.Y) == 0 && pk.X.Cmp(pubKey.X) == 0 {
	//		return index
	//	}
	//}
	return -1
}

func (this *StateStore) HandleBookKeeperTransaction(stateBatch *statestore.StateBatch,tx *types.Transaction) error {
	//bookKeeperState, err := this.GetBookKeeperState()
	//if err != nil {
	//	return fmt.Errorf("GetBookKeeperState error %s", err)
	//}
	//bookKeeper := tx.Payload.(*payload.BookKeeper)
	//index := this.getPubkeysIndex(bookKeeper.PubKey, bookKeeperState.NextBookKeeper)
	//switch bookKeeper.Action {
	//case payload.BookKeeperAction_ADD:
	//	if index >= 0 {
	//		return nil
	//	}
	//	bookKeeperState.NextBookKeeper = append(bookKeeperState.NextBookKeeper, bookKeeper.PubKey)
	//	sort.Sort(crypto.PubKeySlice(bookKeeperState.NextBookKeeper))
	//case payload.BookKeeperAction_SUB:
	//	if index < 0 {
	//		return nil
	//	}
	//	bookSize := len(bookKeeperState.NextBookKeeper)
	//	newNextBookKeeper := make([]*crypto.PubKey, 0, bookSize-1)
	//	for i := 0; i < bookSize; i++ {
	//		if i != index {
	//			newNextBookKeeper = append(newNextBookKeeper, bookKeeperState.NextBookKeeper[i])
	//		}
	//	}
	//	bookKeeperState.NextBookKeeper = newNextBookKeeper
	//}
	//
	//stateBatch.Change(byte(common.ST_BookKeeper), BookerKeeper, false)
	return nil
}

//
//func (this *StateBatch) HandleRegisterAssertTransaction(tx *ctypes.Transaction, height uint32) error {
//	txHash := tx.Hash()
//	registerAsset := tx.Payload.(*payload.RegisterAsset)
//	err := this.TryGetOrAdd(
//		ST_Asset,
//		txHash.ToArray(),
//		&AssetState{
//			AssetId:    txHash,
//			AssetType:  registerAsset.Asset.AssetType,
//			Name:       registerAsset.Asset.Name,
//			Amount:     registerAsset.Amount,
//			Available:  registerAsset.Amount,
//			Precision:  registerAsset.Asset.Precision,
//			Owner:      registerAsset.Issuer,
//			Admin:      registerAsset.Controller,
//			Issuer:     registerAsset.Controller,
//			Expiration: height + 2*2000000,
//			IsFrozen:   false,
//		},
//		false)
//	if err != nil {
//		return fmt.Errorf("TryGetOrAdd asset %x error %s", txHash, err)
//	}
//	return nil
//}
//
//func (this *StateBatch) HandleIssueAssetTransaction(tx *ctypes.Transaction) error {
//	results := tx.GetMergedAssetIDValueFromOutputs()
//	for assetId, amount := range results {
//		state, err := this.TryGetAndChange(ST_Asset, assetId.ToArray(), false)
//		if err != nil {
//			return fmt.Errorf("TryGetAndChange asset %x error %s", assetId, err)
//		}
//		asset := state.(*AssetState)
//		asset.Available -= amount
//	}
//	return nil
//}

func (this *StateStore) HandleDeployTransaction(stateBatch *statestore.StateBatch,tx *types.Transaction) error {
	//deploy := tx.Payload.(*payload.DeployCode)
	//codeHash := deploy.Code.CodeHash()
	//if err := stateBatch.TryGetOrAdd(
	//	common.ST_Contract,
	//	codeHash.ToArray(),
	//	&states.ContractState{
	//		Code:        deploy.Code,
	//		VmType:      deploy.VmType,
	//		NeedStorage: deploy.NeedStorage,
	//		Name:        deploy.Name,
	//		Version:     deploy.CodeVersion,
	//		Author:      deploy.Author,
	//		Email:       deploy.Email,
	//		Description: deploy.Description,
	//	},
	//	false); err != nil {
	//	return fmt.Errorf("TryGetOrAdd contract error %s", err)
	//}
	return nil
}

func (this *StateStore) HandleInvokeTransaction(stateBatch *statestore.StateBatch, tx *types.Transaction, block *types.Block, eventStore common.IEventStore) error {
	//invoke := tx.Payload.(*payload.InvokeCode)
	//txHash := tx.Hash()
	//contrState, err := this.TryGet(ST_Contract, invoke.CodeHash.ToArray())
	//if err != nil {
	//	return fmt.Errorf("TryGet contract error %s", err)
	//}
	//if contrState == nil {
	//	//ledgerevent.PushSmartCodeEvent(tx.Hash(), 0, INVOKE_TRANSACTION, "Contract not found!")
	//	return nil
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
	//	//ledgerevent.PushSmartCodeEvent(txHash, SMARTCODE_ERROR, INVOKE_TRANSACTION, err)
	//	return fmt.Errorf("InvokeContract error %s", err)
	//}
	//stateMachine.CloneCache.Commit()
	//notifications := stateMachine.Notifications
	//err = eventStore.SaveEventNotifyByTx(&txHash, notifications)
	//if err != nil {
	//	return fmt.Errorf("SaveEventNotifyByTx error %s", err)
	//}
	////ledgerevent.PushSmartCodeEvent(txHash, 0, INVOKE_TRANSACTION, ret)
	return nil
}

func (this *StateStore) HandleClaimTransaction(stateBatch *statestore.StateBatch,tx *types.Transaction) error {
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

func (this *StateStore) HandleEnrollmentTransaction(stateBatch *statestore.StateBatch,tx *types.Transaction) error {
	//en := tx.Payload.(*payload.Enrollment)
	//bf := new(bytes.Buffer)
	//if err := en.PublicKey.Serialize(bf); err != nil {
	//	return err
	//}
	//stateBatch.TryAdd(common.ST_Validator, bf.Bytes(), &states.ValidatorState{PublicKey: en.PublicKey}, false)
	return nil
}

func (this *StateStore) HandleVoteTransaction(stateBatch *statestore.StateBatch,tx *types.Transaction) error {
	//vote := tx.Payload.(*payload.Vote)
	//buf := new(bytes.Buffer)
	//vote.Account.Serialize(buf)
	//this.TryAdd(ST_Vote, buf.Bytes(), &VoteState{PublicKeys: vote.PubKeys}, false)
	return nil
}


//
//func (this *StateBatch) HandleTxOutput(tx *ctypes.Transaction) error {
//	txHash := tx.Hash()
//	outputs := tx.Outputs
//	if len(outputs) == 0 {
//		return nil
//	}
//	outputSize := len(tx.Outputs)
//	unspentItems := make([]CoinState, outputSize, outputSize)
//	for i := 0; i < outputSize; i++ {
//		unspentItems[i] = Confirmed
//	}
//	this.TryAdd(ST_Coin, txHash.ToArray(), &UnspentCoinState{Item: unspentItems}, false)
//
//	for index, output := range outputs {
//		ph := output.ProgramHash
//		as := output.AssetID
//		value := output.Value
//		accountState, err := this.TryGetAndChange(ST_Account, ph.ToArray(), true)
//		if err != nil {
//			return fmt.Errorf("TryGetAndChange account state %x error %s", ph.ToArray(), err)
//		}
//		if accountState == nil {
//			balances := make([]*Balance, 0)
//			balance := &Balance{
//				AssetId: as,
//				Amount:  value,
//			}
//			balances = append(balances, balance)
//			this.TryAdd(ST_Account, ph.ToArray(),
//				&AccountState{
//					ProgramHash: ph,
//					IsFrozen:    false,
//					Balances:    balances},
//				true)
//		} else {
//			accState := accountState.(*AccountState)
//			for _, v := range accState.Balances {
//				if v.AssetId.CompareTo(as) == 0 {
//					v.Amount += value
//					break
//				}
//			}
//		}
//
//		unspent := &utxo.UTXOUnspent{
//			Txid:  txHash,
//			Index: uint32(index),
//			Value: value}
//
//		programCoinKey := append(ph.ToArray(), as.ToArray()...)
//		programState, err := this.TryGetAndChange(ST_Program_Coin, programCoinKey, false)
//		if err != nil {
//			return fmt.Errorf("TryGetAndChange programhash state %x error %s", ph.ToArray(), err)
//		}
//		if programState == nil {
//			this.TryAdd(ST_Program_Coin, programCoinKey, &ProgramUnspentCoin{Unspents: []*utxo.UTXOUnspent{unspent}}, false)
//		} else {
//			proState := programState.(*ProgramUnspentCoin)
//			proState.Unspents = append(proState.Unspents, unspent)
//		}
//	}
//	return nil
//}
//
//func (this *StateBatch) HandleTxInput(tx *ctypes.Transaction, height uint32, blockStore *BlockStore) error {
//	for _, input := range tx.UTXOInputs {
//		refTxId := &input.ReferTxID
//		refIndex := input.ReferTxOutputIndex
//		refTx, refHeight, err := blockStore.GetTransaction(refTxId)
//		if err != nil {
//			return fmt.Errorf("GetTransaction tx %x error %s", refTxId, err)
//		}
//		state, err := this.TryGetAndChange(ST_Coin, refTxId.ToArray(), false)
//		if err != nil {
//			return fmt.Errorf("TryGetAndChange coin state reftx %x error %s", refTxId, err)
//		}
//		unspentState := state.(*UnspentCoinState)
//		unspentState.Item[refIndex] = Spent
//
//		state, err = this.TryGetAndChange(ST_SpentCoin, refTxId.ToArray(), false)
//		if err != nil {
//			return fmt.Errorf("TryGetAndChange spent coin %x error %s", refTxId, err)
//		}
//		spentItem := &Item{
//			PrevIndex: refIndex,
//			EndHeight: height,
//		}
//		if state == nil {
//			items := make([]*Item, 0)
//			items = append(items, spentItem)
//			this.TryAdd(ST_SpentCoin, refTxId.ToArray(),
//				&SpentCoinState{
//					TransactionHash:   input.ReferTxID,
//					TransactionHeight: refHeight,
//					Items:             items,
//				}, false)
//		} else {
//			spentState := state.(*SpentCoinState)
//			spentState.Items = append(spentState.Items, spentItem)
//		}
//
//		refTxOutput := refTx.Outputs[refIndex]
//		ph := refTxOutput.ProgramHash
//		as := refTxOutput.AssetID
//		value := refTxOutput.Value
//		state, err = this.TryGetAndChange(ST_Account, ph.ToArray(), true)
//		if err != nil {
//			return fmt.Errorf("TryGetAndChange account state ph %x error %s", ph, err)
//		}
//		accountState := state.(*AccountState)
//		for _, v := range accountState.Balances {
//			if v.AssetId.CompareTo(as) == 0 {
//				v.Amount -= value
//			}
//		}
//
//		programKey := append(ph.ToArray(), as.ToArray()...)
//		state, err = this.TryGetAndChange(ST_Program_Coin, programKey, false)
//		if err != nil {
//			return fmt.Errorf("TryGetAndChange program coin state ph %x asset %x error %s", ph, as, err)
//		}
//		programCoin := state.(*ProgramUnspentCoin)
//		unspents := make([]*utxo.UTXOUnspent, 0, len(programCoin.Unspents)-1)
//		refTxIdByte := refTxId.ToArray()
//		for _, unspent := range programCoin.Unspents {
//			if uint32(refIndex) == unspent.Index && bytes.EqualFold(refTxIdByte, unspent.Txid.ToArray()) {
//				continue
//			}
//			unspents = append(unspents, unspent)
//		}
//		programCoin.Unspents = unspents
//	}
//	return nil
//}
