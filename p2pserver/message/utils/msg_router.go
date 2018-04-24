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
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/log"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	msgTypes "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

// MessageHandler defines the unified api for each net message
type MessageHandler func(data *msgCommon.MsgPayload, args ...interface{}) error

// DefaultMsgHandler defines the default message handler
func DefaultMsgHandler(data *msgCommon.MsgPayload, args ...interface{}) error {
	return nil
}

// MessageRouter mostly route different message type-based to the
// related message handler
type MessageRouter struct {
	msgHandlers  map[string]MessageHandler  // Msg handler mapped to msg type
	RecvSyncChan chan *msgCommon.MsgPayload // The channel to handle sync msg
	RecvConsChan chan *msgCommon.MsgPayload // The channel to handle consensus msg
	stopSyncCh   chan bool                  // To stop sync channel
	stopConsCh   chan bool                  // To stop consensus channel
	p2p          p2p.P2P                    // Refer to the p2p network
	pid          *actor.PID                 // P2P actor
}

// NewMsgRouter returns a message router object
func NewMsgRouter(p2p p2p.P2P) *MessageRouter {
	msgRouter := &MessageRouter{}
	msgRouter.init(p2p)
	return msgRouter
}

// init initializes the message router's attributes
func (self *MessageRouter) init(p2p p2p.P2P) {
	self.msgHandlers = make(map[string]MessageHandler)
	self.RecvSyncChan = p2p.GetMsgChan(false)
	self.RecvConsChan = p2p.GetMsgChan(true)
	self.stopSyncCh = make(chan bool)
	self.stopConsCh = make(chan bool)
	self.p2p = p2p

	// Register message handler
	self.RegisterMsgHandler(msgCommon.VERSION_TYPE, VersionHandle)
	self.RegisterMsgHandler(msgCommon.VERACK_TYPE, VerAckHandle)
	self.RegisterMsgHandler(msgCommon.GetADDR_TYPE, AddrReqHandle)
	self.RegisterMsgHandler(msgCommon.ADDR_TYPE, AddrHandle)
	self.RegisterMsgHandler(msgCommon.PING_TYPE, PingHandle)
	self.RegisterMsgHandler(msgCommon.PONG_TYPE, PongHandle)
	self.RegisterMsgHandler(msgCommon.GET_HEADERS_TYPE, HeadersReqHandle)
	self.RegisterMsgHandler(msgCommon.HEADERS_TYPE, BlkHeaderHandle)
	self.RegisterMsgHandler(msgCommon.INV_TYPE, InvHandle)
	self.RegisterMsgHandler(msgCommon.GET_DATA_TYPE, DataReqHandle)
	self.RegisterMsgHandler(msgCommon.BLOCK_TYPE, BlockHandle)
	self.RegisterMsgHandler(msgCommon.CONSENSUS_TYPE, ConsensusHandle)
	self.RegisterMsgHandler(msgCommon.NOT_FOUND_TYPE, NotFoundHandle)
	self.RegisterMsgHandler(msgCommon.TX_TYPE, TransactionHandle)
	self.RegisterMsgHandler(msgCommon.DISCONNECT_TYPE, DisconnectHandle)
}

// RegisterMsgHandler registers msg handler with the msg type
func (self *MessageRouter) RegisterMsgHandler(key string,
	handler MessageHandler) {
	self.msgHandlers[key] = handler
}

// UnRegisterMsgHandler un-registers the msg handler with
// the msg type
func (self *MessageRouter) UnRegisterMsgHandler(key string) {
	delete(self.msgHandlers, key)
}

// SetPID sets p2p actor
func (self *MessageRouter) SetPID(pid *actor.PID) {
	self.pid = pid
}

// Start starts the loop to handle the message from the network
func (self *MessageRouter) Start() {
	go self.hookChan(self.RecvSyncChan, self.stopSyncCh)
	go self.hookChan(self.RecvConsChan, self.stopConsCh)
	log.Info("MessageRouter start to parse p2p message...")
}

// hookChan loops to handle the message from the network
func (self *MessageRouter) hookChan(channel chan *msgCommon.MsgPayload,
	stopCh chan bool) {
	for {
		select {
		case data, ok := <-channel:
			if ok {
				msgType, err := msgTypes.MsgType(data.Payload)
				if err != nil {
					log.Info("failed to get msg type")
					continue
				}

				handler, ok := self.msgHandlers[msgType]
				if ok {
					go handler(data, self.p2p, self.pid)
				} else {
					log.Info("Unkown message handler for the msg: ",
						msgType)
				}
			}
		case <-stopCh:
			return
		}
	}
}

// Stop stops the message router's loop
func (self *MessageRouter) Stop() {

	if self.stopSyncCh != nil {
		self.stopSyncCh <- true
	}
	if self.stopConsCh != nil {
		self.stopConsCh <- true
	}
}
