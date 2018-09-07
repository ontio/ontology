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
package utils

import (
	"encoding/hex"
	"fmt"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"testing"
)

func TestParseTransferInvokeCode(t *testing.T) {
	fromAccount := account.NewAccount("")
	toAccount := account.NewAccount("")
	amount := uint64(100)
	ontState1 := &ont.State{
		From:  fromAccount.Address,
		To:    toAccount.Address,
		Value: amount,
	}
	tx, err := MultiTransferTx(500, 20000, "ont", []*ont.State{ontState1})
	if err != nil {
		t.Errorf("TransferTx error:%s", err)
		return
	}
	invoke := tx.Payload.(*payload.InvokeCode)
	invokeInfo, err := ParseInvokeCode(hex.EncodeToString(invoke.Code))
	if err != nil {
		t.Errorf("ParseInvokeCode error:%s", err)
		return
	}
	states := invokeInfo.Params[0].([]*NativeTransferState)
	if len(states) != 1 {
		t.Errorf("NativeTransferState len:%d != 1", len(states))
		return
	}
	err = judgeTransferState(states[0], ontState1)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	amount = 200
	ontState2 := &ont.State{
		From:  fromAccount.Address,
		To:    toAccount.Address,
		Value: amount,
	}
	ontStates := []*ont.State{ontState1, ontState2}
	tx, err = MultiTransferTx(500, 2000, "ont", ontStates)
	if err != nil {
		t.Errorf("MultiTransferTx error:%s", err)
		return
	}
	invoke = tx.Payload.(*payload.InvokeCode)
	invokeInfo, err = ParseInvokeCode(hex.EncodeToString(invoke.Code))
	if err != nil {
		t.Errorf("ParseInvokeCode error:%s", err)
		return
	}
	states = invokeInfo.Params[0].([]*NativeTransferState)
	if len(states) != 2 {
		t.Errorf("NativeTransferState len:%d != 2", len(states))
		return
	}

	for i, state := range states {
		err = judgeTransferState(state, ontStates[i])
		if err != nil {
			t.Errorf("State:%d error:%s", i, err.Error())
			return
		}
	}
}

func TestParseApproveInvokeCode(t *testing.T) {
	fromAccount := account.NewAccount("")
	toAccount := account.NewAccount("")
	amount := uint64(100)
	ontState := &ont.State{
		From:  fromAccount.Address,
		To:    toAccount.Address,
		Value: amount,
	}
	tx, err := ApproveTx(500, 20000, "ont", ontState.From.ToBase58(), ontState.To.ToBase58(), ontState.Value)
	if err != nil {
		t.Errorf("ApproveTx error:%s", err)
		return
	}
	invoke := tx.Payload.(*payload.InvokeCode)
	invokeInfo, err := ParseInvokeCode(hex.EncodeToString(invoke.Code))
	if err != nil {
		t.Errorf("ParseInvokeCode error:%s", err)
		return
	}
	state := invokeInfo.Params[0].(*NativeTransferState)
	err = judgeTransferState(state, ontState)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}

func TestParseTransferFromInvokeCode(t *testing.T) {
	fromAccount := account.NewAccount("")
	toAccount := account.NewAccount("")
	amount := uint64(100)
	ontState := &ont.TransferFrom{
		Sender: toAccount.Address,
		From:   fromAccount.Address,
		To:     toAccount.Address,
		Value:  amount,
	}
	tx, err := TransferFromTx(500, 20000, "ont", ontState.Sender.ToBase58(), ontState.From.ToBase58(), ontState.To.ToBase58(), ontState.Value)
	if err != nil {
		t.Errorf("TransferFromTx error:%s", err)
		return
	}
	invoke := tx.Payload.(*payload.InvokeCode)
	invokeInfo, err := ParseInvokeCode(hex.EncodeToString(invoke.Code))
	if err != nil {
		t.Errorf("ParseInvokeCode error:%s", err)
		return
	}
	state := invokeInfo.Params[0].(*NativeTransferFromState)
	err = judgeTransferFromState(state, ontState)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
}

func judgeTransferState(state *NativeTransferState, ontState *ont.State) error {
	if state.From != ontState.From.ToBase58() {
		return fmt.Errorf("from account:%s != %s", state.From, ontState.From.ToBase58())
	}
	if state.To != ontState.To.ToBase58() {
		return fmt.Errorf("to account:%s != %s", state.To, ontState.To.ToBase58())
	}
	if state.Value != ontState.Value {
		return fmt.Errorf("amount:%d != %d", state.Value, ontState.Value)
	}
	return nil
}

func judgeTransferFromState(state *NativeTransferFromState, ontState *ont.TransferFrom) error {
	if state.Sender != ontState.Sender.ToBase58() {
		return fmt.Errorf("sender account:%s != %s", state.Sender, ontState.Sender.ToBase58())
	}
	if state.From != ontState.From.ToBase58() {
		return fmt.Errorf("from account:%s != %s", state.From, ontState.From.ToBase58())
	}
	if state.To != ontState.To.ToBase58() {
		return fmt.Errorf("to account:%s != %s", state.To, ontState.To.ToBase58())
	}
	if state.Value != ontState.Value {
		return fmt.Errorf("amount:%d != %d", state.Value, ontState.Value)
	}
	return nil
}
