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

package message

import (
	"github.com/ethereum/go-ethereum/core"
	types2 "github.com/ethereum/go-ethereum/core/types"
	"github.com/ontio/ontology/core/types"
)

const (
	TOPIC_SAVE_BLOCK_COMPLETE = "svblkcmp"
	TOPIC_SMART_CODE_EVENT    = "scevt"
	TOPIC_PENDING_TX_EVENT    = "pendingtx"
	TOPIC_CHAIN_EVENT         = "chainevt"
	TOPIC_ETH_SC_EVENT        = "ethscevt"
)

type SaveBlockCompleteMsg struct {
	Block *types.Block
}

type SmartCodeEventMsg struct {
	Event *types.SmartCodeEvent
}

type BlockConsensusComplete struct {
	Block *types.Block
}

type EthSmartCodeEventMsg struct {
	Event EthSmartCodeEvent
}
type PendingTxs []*types2.Transaction

type PendingTxMsg struct {
	Event PendingTxs
}

type EthSmartCodeEvent []*types2.Log

type ChainEventMsg struct {
	ChainEvent *core.ChainEvent
}
