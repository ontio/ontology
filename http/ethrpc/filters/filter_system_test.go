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

package filters

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	common2 "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/ontology/events"
	"github.com/ontio/ontology/events/message"
)

func TestNewEventSystem(t *testing.T) {
	events.Init()

	es := NewEventSystem(nil)
	fmt.Println("es:", es)

	header := &types.Header{Number: big.NewInt(1)}
	chainEvtMsg := &message.ChainEventMsg{
		ChainEvent: &core.ChainEvent{
			Block: types.NewBlockWithHeader(header),
			Hash:  common2.Hash{},
			Logs:  make([]*types.Log, 0),
		},
	}
	events.DefActorPublisher.Publish(message.TOPIC_CHAIN_EVENT, chainEvtMsg)
	time.Sleep(time.Second)

	scEvt := make(message.EthSmartCodeEvent, 0)
	scEvt = append(scEvt, &types.Log{})
	ethLog := &message.EthSmartCodeEventMsg{Event: scEvt}
	events.DefActorPublisher.Publish(message.TOPIC_ETH_SC_EVENT, ethLog)
	time.Sleep(time.Second)

	pendingTxEvt := &message.PendingTxMsg{Event: []*types.Transaction{&types.Transaction{}}}
	events.DefActorPublisher.Publish(message.TOPIC_PENDING_TX_EVENT, pendingTxEvt)
	time.Sleep(time.Second)
}
