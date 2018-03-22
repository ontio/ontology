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

package vbft

import (
	"fmt"
	"sync"

	. "github.com/Ontology/common"
)

type ConsensusRoundMsgs map[MsgType][]ConsensusMsg // indexed by MsgType (proposal, endorsement, ...)

type ConsensusRound struct {
	blockNum uint64
	msgs     map[MsgType][]ConsensusMsg
	msgHashs map[Uint256]interface{} // for msg-dup checking
}

func newConsensusRound(num uint64) *ConsensusRound {

	r := &ConsensusRound{
		blockNum: num,
		msgs:     make(map[MsgType][]ConsensusMsg),
		msgHashs: make(map[Uint256]interface{}),
	}

	r.msgs[blockProposalMessage] = make([]ConsensusMsg, 0)
	r.msgs[blockEndorseMessage] = make([]ConsensusMsg, 0)
	r.msgs[blockCommitMessage] = make([]ConsensusMsg, 0)

	return r
}

func (self *ConsensusRound) addMsg(msg ConsensusMsg) error {
	h, err := HashMsg(msg)
	if err != nil {
		return fmt.Errorf("failed to hash msg: %s", err)
	}
	if _, present := self.msgHashs[h]; present {
		return nil
	}

	msgs := self.msgs[msg.Type()]
	self.msgs[msg.Type()] = append(msgs, msg)
	self.msgHashs[h] = nil
	return nil
}

func (self *ConsensusRound) hasMsg(msg ConsensusMsg) (bool, error) {
	h, err := HashMsg(msg)
	if err != nil {
		return false, fmt.Errorf("failed to hash msg: %s", err)
	}
	if _, present := self.msgHashs[h]; present {
		return present, nil
	}
	return false, nil
}

type MsgPool struct {
	lock       sync.RWMutex
	server     *Server
	historyLen uint64
	rounds     map[uint64]*ConsensusRound // indexed by BlockNum
}

func newMsgPool(server *Server, historyLen uint64) *MsgPool {
	// TODO
	return &MsgPool{
		historyLen: historyLen,
		server:     server,
		rounds:     make(map[uint64]*ConsensusRound),
	}
}

func (pool *MsgPool) AddMsg(msg ConsensusMsg) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	if _, present := pool.rounds[blkNum]; !present {
		pool.rounds[blkNum] = newConsensusRound(blkNum)
	}

	// TODO: limit #history rounds to historyLen
	// Note: we accept msg for future rounds

	return pool.rounds[blkNum].addMsg(msg)
}

func (pool *MsgPool) HasMsg(msg ConsensusMsg) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if roundMsgs, present := pool.rounds[msg.GetBlockNum()]; !present {
		return false
	} else {
		if present, err := roundMsgs.hasMsg(msg); err != nil {
			pool.server.log.Errorf("msgpool failed to check msg avail: %s", err)
			return false
		} else {
			return present
		}
	}

	return false
}

func (pool *MsgPool) Persist() error {
	// TODO
	return nil
}

func (pool *MsgPool) GetProposalMsgs(blocknum uint64) []ConsensusMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msgs, ok := roundMsgs.msgs[blockProposalMessage]
	if !ok {
		return nil
	}
	return msgs
}

func (pool *MsgPool) GetEndorsementsMsgs(blocknum uint64) []ConsensusMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msgs, ok := roundMsgs.msgs[blockEndorseMessage]
	if !ok {
		return nil
	}
	return msgs
}

func (pool *MsgPool) GetCommitMsgs(blocknum uint64) []ConsensusMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msg, ok := roundMsgs.msgs[blockCommitMessage]
	if !ok {
		return nil
	}
	return msg
}

func (pool *MsgPool) onBlockSealed(blockNum uint64) {
	if blockNum <= pool.historyLen {
		return
	}
	pool.lock.Lock()
	defer pool.lock.Unlock()

	toFreeRound := make([]uint64, 0)
	for n := range pool.rounds {
		if n < blockNum-pool.historyLen {
			toFreeRound = append(toFreeRound, n)
		}
	}
	for _, n := range toFreeRound {
		delete(pool.rounds, n)
	}
}
