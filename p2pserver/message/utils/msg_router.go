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
	"github.com/ontio/ontology/p2pserver/message/types"
	p2p "github.com/ontio/ontology/p2pserver/net/protocol"
	"github.com/ontio/ontology/p2pserver/protocols"
)

// MessageHandler defines the unified api for each net message
type MessageHandler func(data *types.MsgPayload, p2p p2p.P2P, pid *actor.PID)

// MessageRouter mostly route different message type-based to the
// related message handler
type MessageRouter struct {
	msgHander  protocols.Protocol
	RecvChan   chan *types.MsgPayload // The channel to handle sync msg
	stopRecvCh chan bool              // To stop sync channel
	p2p        p2p.P2P                // Refer to the p2p network
	pid        *actor.PID             // P2P actor
}

// NewMsgRouter returns a message router object
func NewMsgRouter(p2p p2p.P2P) *MessageRouter {
	router := &MessageRouter{}
	router.RecvChan = p2p.GetMsgChan()
	router.stopRecvCh = make(chan bool)
	router.p2p = p2p
	router.msgHander = &MsgHandler{}
	return router
}

// SetPID sets p2p actor
func (this *MessageRouter) SetPID(pid *actor.PID) {
	this.pid = pid
}

// Start starts the loop to handle the message from the network
func (this *MessageRouter) Start() {
	go this.hookChan(this.RecvChan, this.stopRecvCh)
	log.Debug("[p2p]MessageRouter start to parse p2p message...")
}

// hookChan loops to handle the message from the network
func (this *MessageRouter) hookChan(channel chan *types.MsgPayload,
	stopCh chan bool) {
	for {
		select {
		case data, ok := <-channel:
			if ok {
				sender := this.p2p.GetPeer(data.Id)
				if sender == nil {
					log.Warnf("[router] remote peer %d invalid.", data.Id)
					continue
				}

				ctx := protocols.NewContext(sender, this.p2p, this.pid, data.PayloadSize)
				go this.msgHander.HandleMessage(ctx, data.Payload)
			}
		case <-stopCh:
			return
		}
	}
}

// Stop stops the message router's loop
func (this *MessageRouter) Stop() {
	if this.stopRecvCh != nil {
		this.stopRecvCh <- true
	}
}
