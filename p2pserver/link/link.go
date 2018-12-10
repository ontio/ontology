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

package link

import (
	"errors"
	"fmt"
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"

	tsp "github.com/ontio/ontology/p2pserver/net/transport"
	tspCreator "github.com/ontio/ontology/p2pserver/net/transport/creator"
)



//Link used to establish
type Link struct {
	id        uint64
	addr      string                  // The address of the node
	conn      tsp.Connection          // Connect socket with the peer node
	port      uint16                   // The server port of the node
	time      time.Time                // The latest time the node activity
	recvChan  chan *types.RecvMessage //receive message  channel
	reqRecord map[string]int64        //Map RequestId to Timestamp, using for rejecting duplicate request in specific time
}

func NewLink() *Link {
	link := &Link{
		reqRecord: make(map[string]int64, 0),
	}
	return link
}

//SetID set peer id to link
func (this *Link) SetID(id uint64) {
	this.id = id
}

//GetID return if from peer
func (this *Link) GetID() uint64 {
	return this.id
}

//If there is connection return true
func (this *Link) Valid() bool {
	return this.conn != nil
}

//set message channel for link layer
func (this *Link) SetChan(msgchan chan *types.RecvMessage) {
	this.recvChan = msgchan
}

//get address
func (this *Link) GetAddr() string {
	return this.addr
}

//set address
func (this *Link) SetAddr(addr string) {
	this.addr = addr
}

//set port number
func (this *Link) SetPort(p uint16) {
	this.port = p
}

//get port number
func (this *Link) GetPort() uint16 {
	return this.port
}

//get connection
func (this *Link) GetConn() tsp.Connection{
	return this.conn
}

//set connection
func (this *Link) SetConn(conn tsp.Connection) {
	this.conn = conn
}

//record latest message time
func (this *Link) UpdateRXTime(t time.Time) {
	this.time = t
}

//GetRXTime return the latest message time
func (this *Link) GetRXTime() time.Time {
	return this.time
}

func (this *Link) Rx(tspType byte) {
	conn := this.conn
	if conn == nil {
		return
	}

	for {
		reader, err :=  conn.GetReader()
		if err != nil {
			log.Errorf("[p2p]error GetReader, err:%s", err.Error())
			break
		}

		msg, payloadSize, err := types.ReadMessage(reader)
		if err != nil {
			log.Infof("[p2p]error read from %s :%s", this.GetAddr(), err.Error())
			break
		} else {
			log.Infof("[p2p]success read msg %s from %s ", msg.CmdType(), this.GetAddr())
		}

		t := time.Now()
		this.UpdateRXTime(t)

		if !this.needSendMsg(msg) {
			log.Debugf("skip handle msgType:%s from:%d", msg.CmdType(), this.id)
			log.Infof("skip handle msgType:%s from:%d", msg.CmdType(), this.id)
			continue
		}
		this.addReqRecord(msg)
		log.Infof("Start send to recvChan msgType:%s from:%d", msg.CmdType(), this.id)
		msgPayload := &types.MsgPayload{
			Id:          this.id,
			Addr:        this.addr,
			PayloadSize: payloadSize,
			Payload:     msg,
		}
		this.recvChan <- &types.RecvMessage{
			tspType,
			msgPayload,
		}
	}

	this.disconnectNotify()
}

//disconnectNotify push disconnect msg to channel
func (this *Link) disconnectNotify(tspType byte) {
	log.Debugf("[p2p]call disconnectNotify for %s", this.GetAddr())
	this.CloseConn()

	msg, _ := types.MakeEmptyMessage(common.DISCONNECT_TYPE)
	discMsg := &types.MsgPayload{
		Id:      this.id,
		Addr:    this.addr,
		Payload: msg,
	}
	this.recvChan <- &types.RecvMessage{
		tspType,
		discMsg,
	}
}

//close connection
func (this *Link) CloseConn() {
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
}

func (this *Link) Tx(msg types.Message, tspType byte) error {
	conn := this.conn
	if conn == nil {
		return errors.New("[p2p]tx link invalid")
	}

	sink := comm.NewZeroCopySink(nil)
	err := types.WriteMessage(sink, msg)
	if err != nil {
		log.Debugf("[p2p]error serialize messge ", err.Error())
		return err
	}

	payload := sink.Bytes()
	nByteCnt := len(payload)
	log.Tracef("[p2p]TX buf length: %d\n", nByteCnt)

	nCount := nByteCnt / common.PER_SEND_LEN
	if nCount == 0 {
		nCount = 1
	}
	conn.SetWriteDeadline(time.Now().Add(time.Duration(nCount*common.WRITE_DEADLINE) * time.Second))
	_, err = conn.Write(payload)
	if err != nil {
		log.Infof("[p2p]error sending messge %s to %s :%s", msg.CmdType(), this.GetAddr(), err.Error())
		this.disconnectNotify(tspType)
		return err
	} else {
		log.Infof("[p2p]success sending messge %s to %s", msg.CmdType(), this.GetAddr())
	}

	return nil
}

//needSendMsg check whether the msg is needed to push to channel
func (this *Link) needSendMsg(msg types.Message) bool {
	if msg.CmdType() != common.GET_DATA_TYPE {
		return true
	}
	var dataReq = msg.(*types.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.ToHexString())
	now := time.Now().Unix()

	tspType := config.DefConfig.P2PNode.TransportType
	transport, err := tspCreator.GetTransportFactory().GetTransport(tspType)
	if err != nil {
		log.Errorf("[p2p]Get the transport of %s, err:%s", tspType, err.Error())
		return false
	}

	if t, ok := this.reqRecord[reqID]; ok {
		if int(now-t) < transport.GetReqInterval() {
			return false
		}
	}
	return true
}

//addReqRecord add request record by removing outdated request records
func (this *Link) addReqRecord(msg types.Message) {
	if msg.CmdType() != common.GET_DATA_TYPE {
		return
	}

	tspType := config.DefConfig.P2PNode.TransportType
	transport, err := tspCreator.GetTransportFactory().GetTransport(tspType)
	if err != nil {
		log.Errorf("[p2p]Get the transport of %s, err:%s", tspType, err.Error())
		return
	}

	now := time.Now().Unix()
	if len(this.reqRecord) >= common.MAX_REQ_RECORD_SIZE-1 {
		for id := range this.reqRecord {
			t := this.reqRecord[id]
			if int(now-t) > transport.GetReqInterval() {
				delete(this.reqRecord, id)
			}
		}
	}
	var dataReq = msg.(*types.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.ToHexString())
	this.reqRecord[reqID] = now
}
