package p2pserver

import (
	"github.com/Ontology/common/log"
	msgCommon "github.com/Ontology/p2pserver/common"
	msg "github.com/Ontology/p2pserver/message"
)

type MessageHandler func(data msgCommon.MsgPayload, p2p *P2PServer) error

func DefaultMsgHandler(data msgCommon.MsgPayload, p2p *P2PServer) error {
	return nil
}

type MessageRouter struct {
	msgHandlers map[string]MessageHandler
	ReceiveChan chan msgCommon.MsgPayload

	//Actors       map[string]actor
	stopCh chan bool
	p2p    *P2PServer
}

func NewMsgRouter(p2p *P2PServer) *MessageRouter {
	msgRouter := &MessageRouter{}
	msgRouter.init(p2p)
	return msgRouter
}

func (self *MessageRouter) init(p2p *P2PServer) {
	self.msgHandlers = make(map[string]MessageHandler)
	self.ReceiveChan = make(chan msgCommon.MsgPayload, 10000)
	self.stopCh = make(chan bool)
	self.p2p = p2p
}

func (self *MessageRouter) RegisterMsgHandler(key string, handler MessageHandler) {
	self.msgHandlers[key] = handler
}

func (self *MessageRouter) UnRegisterMsgHandler(key string) {
	delete(self.msgHandlers, key)
}

func (self *MessageRouter) Start() {
	for {
		select {
		case data, ok := <-self.ReceiveChan:
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
		case <-self.stopCh:
			return
		}
	}
}

func (self *MessageRouter) Stop() {
	if self.ReceiveChan != nil {
		close(self.ReceiveChan)
	}

	if self.stopCh != nil {
		self.stopCh <- true
	}
}
