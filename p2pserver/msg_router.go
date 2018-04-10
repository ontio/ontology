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

package p2pserver

import (
	"github.com/ontio/ontology/common/log"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	msg "github.com/ontio/ontology/p2pserver/message"
	_ "github.com/ontio/ontology/p2pserver/peer"
)

type MessageHandler func(data msgCommon.MsgPayload, p2p *P2PServer) error

func DefaultMsgHandler(data msgCommon.MsgPayload, p2p *P2PServer) error {
	return nil
}

type MessageRouter struct {
	msgHandlers  map[string]MessageHandler
	RecvSyncChan chan msgCommon.MsgPayload
	RecvConsChan chan msgCommon.MsgPayload
	stopSyncCh   chan bool
	stopConsCh   chan bool
	p2p          *P2PServer
}

func NewMsgRouter(p2p *P2PServer) *MessageRouter {
	msgRouter := &MessageRouter{}
	msgRouter.init(p2p)
	return msgRouter
}

func (self *MessageRouter) init(p2p *P2PServer) {
	self.msgHandlers = make(map[string]MessageHandler)
	self.RecvSyncChan = p2p.network.GetMsgChan(false)
	self.RecvConsChan = p2p.network.GetMsgChan(true)
	self.stopSyncCh = make(chan bool)
	self.stopConsCh = make(chan bool)
	self.p2p = p2p
}

func (self *MessageRouter) RegisterMsgHandler(key string, handler MessageHandler) {
	self.msgHandlers[key] = handler
}

func (self *MessageRouter) UnRegisterMsgHandler(key string) {
	delete(self.msgHandlers, key)
}

func (self *MessageRouter) Start() {
	go self.hookChan(self.RecvSyncChan, self.stopSyncCh)
	go self.hookChan(self.RecvConsChan, self.stopConsCh)
	log.Info("Start to read p2p message...")
}

func (self *MessageRouter) hookChan(channel chan msgCommon.MsgPayload, stopCh chan bool) {
	for {
		select {
		case data, ok := <-channel:
			if ok {
				msgType, err := msg.MsgType(data.Payload)
				if err != nil {
					log.Info("failed to get msg type")
					continue
				}

				handler, ok := self.msgHandlers[msgType]
				if ok {
					go handler(data, self.p2p)
				} else {
					log.Info("Unkown message handler for the msg: ", msgType)
				}
			}
		case <-stopCh:
			return
		}
	}
}

func (self *MessageRouter) Stop() {

	if self.stopSyncCh != nil {
		self.stopSyncCh <- true
	}
	if self.stopConsCh != nil {
		self.stopConsCh <- true
	}
}
