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
	"sync"

	"github.com/ontio/ontology/common"
)

type ConsensusRound struct {
	msgs     map[MsgType][]BftConsensusMsg
	msgHashs map[common.Uint256]BftConsensusMsg // for msg-dup checking
}

func newConsensusRound() *ConsensusRound {
	return &ConsensusRound{
		msgs:     make(map[MsgType][]BftConsensusMsg),
		msgHashs: make(map[common.Uint256]BftConsensusMsg),
	}
}

func (self *ConsensusRound) addMsg(msg BftConsensusMsg) {
	msgHash := HashMsg(msg)
	if _, present := self.msgHashs[msgHash]; present {
		return
	}

	msgs := self.msgs[msg.Type()]
	self.msgs[msg.Type()] = append(msgs, msg)
	self.msgHashs[msgHash] = msg
}

type MsgPool struct {
	lock       sync.RWMutex
	server     *Server
	historyLen uint32
	rounds     map[uint32]*ConsensusRound // indexed by BlockNum
}

func newMsgPool(server *Server, historyLen uint32) *MsgPool {
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

func (pool *MsgPool) AddMsg(msg BftConsensusMsg) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	blkNum := msg.GetBlockNum()
	currNum := pool.server.GetCurrentBlockNo()
	if blkNum > currNum+pool.historyLen || blkNum < currNum {
		return
	}

	if _, present := pool.rounds[blkNum]; !present {
		pool.rounds[blkNum] = newConsensusRound()
	}

	pool.rounds[blkNum].addMsg(msg)
}

func (pool *MsgPool) HasMsg(msg BftConsensusMsg) bool {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, present := pool.rounds[msg.GetBlockNum()]
	return present && roundMsgs.msgHashs[HashMsg(msg)] != nil
}

func (pool *MsgPool) GetBlockSubmitMsgs(blocknum uint32) []BftConsensusMsg {
	return pool.getRoundMsg(blocknum, BlockSubmitMessage)
}

func (pool *MsgPool) GetProposalMsg(blocknum, proposer uint32) *blockProposalMsg {
	for _, msg := range pool.getRoundMsg(blocknum, BlockProposalMessage) {
		pMsg := msg.(*blockProposalMsg)
		if pMsg.Proposer == proposer {
			return pMsg
		}
	}
	return nil
}

func (pool *MsgPool) getRoundMsg(blocknum uint32, msgType MsgType) (result []BftConsensusMsg) {
	pool.lock.RLock()
	defer pool.lock.RUnlock()

	roundMsgs, ok := pool.rounds[blocknum]
	if !ok {
		return nil
	}
	msg, ok := roundMsgs.msgs[msgType]
	if !ok {
		return nil
	}
	result = append(result, msg...)
	return
}

func (pool *MsgPool) OnBlockSealed(blockNum uint32) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	for n := range pool.rounds {
		if n+pool.historyLen < blockNum {
			delete(pool.rounds, n)
		}
	}
}

func (pool *MsgPool) DropBftMsgs(block uint32) (result []BftConsensusMsg) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

	roundMsgs, ok := pool.rounds[block]
	if !ok {
		return nil
	}
	for _, msgType := range []MsgType{BlockProposalMessage, BlockEndorseMessage, BlockCommitMessage} {
		result = append(result, roundMsgs.msgs[msgType]...)
		roundMsgs.msgs[msgType] = nil
	}
	roundMsgs.msgHashs = make(map[common.Uint256]BftConsensusMsg)
	for _, msg := range roundMsgs.msgs[BlockSubmitMessage] {
		roundMsgs.msgHashs[HashMsg(msg)] = msg
	}
	return
}
