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
	"errors"
	"sync"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
)

var errDropFarFutureMsg = errors.New("msg pool dropped msg for far future")

type ConsensusRoundMsgs map[MsgType][]ConsensusMsg // indexed by MsgType (proposal, endorsement, ...)

type ConsensusRound struct {
	blockNum uint32
	msgs     map[MsgType][]ConsensusMsg
	msgHashs map[common.Uint256]interface{} // for msg-dup checking
}

func newConsensusRound(num uint32) *ConsensusRound {

	r := &ConsensusRound{
		blockNum: num,
		msgs:     make(map[MsgType][]ConsensusMsg),
		msgHashs: make(map[common.Uint256]interface{}),
	}

	r.msgs[BlockProposalMessage] = make([]ConsensusMsg, 0)
	r.msgs[BlockEndorseMessage] = make([]ConsensusMsg, 0)
	r.msgs[BlockCommitMessage] = make([]ConsensusMsg, 0)

	return r
}

func (self *ConsensusRound) addMsg(msg ConsensusMsg, msgHash common.Uint256) error {
	if _, present := self.msgHashs[msgHash]; present {
		return nil
	}

	msgs := self.msgs[msg.Type()]
	self.msgs[msg.Type()] = append(msgs, msg)
	self.msgHashs[msgHash] = nil
	return nil
}

func (self *ConsensusRound) hasMsg(msg ConsensusMsg, msgHash common.Uint256) (bool, error) {
	if _, present := self.msgHashs[msgHash]; present {
		return present, nil
	}
	return false, nil
}

type MsgPool struct {
	lock       sync.RWMutex
	server     *Server
	historyLen uint32
	rounds     map[uint32]*ConsensusRound // indexed by BlockNum
}

func newMsgPool(server *Server, historyLen uint32) *MsgPool {
	// TODO
	return &MsgPool{
		historyLen: historyLen,
		server:     server,
		rounds:     make(map[uint32]*ConsensusRound),
	}
}

func (pool *MsgPool) clean() {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	pool.rounds = make(map[uint32]*ConsensusRound)
}

func (pool *MsgPool) AddMsg(msg ConsensusMsg, msgHash common.Uint256) error {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	if blkNum > pool.server.GetCurrentBlockNo()+pool.historyLen {
		return errDropFarFutureMsg
	}

	if _, present := pool.rounds[blkNum]; !present {
		pool.rounds[blkNum] = newConsensusRound(blkNum)
	}

	// TODO: limit #history rounds to historyLen
	// Note: we accept msg for future rounds

	return pool.rounds[blkNum].addMsg(msg, msgHash)
}

func (pool *MsgPool) HasMsg(msg ConsensusMsg, msgHash common.Uint256) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	if roundMsgs, present := pool.rounds[msg.GetBlockNum()]; !present {
		return false
	} else {
		if present, err := roundMsgs.hasMsg(msg, msgHash); err != nil {
			log.Errorf("msgpool failed to check msg avail: %s", err)
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

func (pool *MsgPool) GetProposalMsgs(blocknum uint32) []ConsensusMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msgs, ok := roundMsgs.msgs[BlockProposalMessage]
	if !ok {
		return nil
	}
	return msgs
}

func (pool *MsgPool) GetEndorsementsMsgs(blocknum uint32) []ConsensusMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msgs, ok := roundMsgs.msgs[BlockEndorseMessage]
	if !ok {
		return nil
	}
	return msgs
}

func (pool *MsgPool) GetCommitMsgs(blocknum uint32) []ConsensusMsg {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msg, ok := roundMsgs.msgs[BlockCommitMessage]
	if !ok {
		return nil
	}
	return msg
}

func (pool *MsgPool) onBlockSealed(blockNum uint32) {
	if blockNum <= pool.historyLen {
		return
	}
	pool.lock.Lock()
	defer pool.lock.Unlock()

	toFreeRound := make([]uint32, 0)
	for n := range pool.rounds {
		if n < blockNum-pool.historyLen {
			toFreeRound = append(toFreeRound, n)
		}
	}
	for _, n := range toFreeRound {
		delete(pool.rounds, n)
	}
}
