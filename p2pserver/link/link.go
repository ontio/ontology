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
	"bufio"
	"net"
	"sync"
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

const MaxWriteBuff = 16 * 1024 // add backpressure when pending msg in write buffer execced this

//Link used to establish
type Link struct {
	id        uint64
	addr      string                  // The address of the node
	conn      net.Conn                // Connect socket with the peer node
	port      uint16                  // The server port of the node
	time      time.Time               // The latest time the node activity
	recvChan  chan *types.MsgPayload  //msgpayload channel
	reqRecord map[types.DataReq]int64 //Map RequestId to Timestamp, using for rejecting duplicate request in specific time

	// owned by write loop
	writeBuff *comm.ZeroCopySink

	maxWriteBuff int
	lock         sync.Mutex
	cond         *sync.Cond
	err          error
	pendingBuff  *comm.ZeroCopySink
}

func NewLink() *Link {
	link := &Link{
		reqRecord:    make(map[types.DataReq]int64, 0),
		maxWriteBuff: MaxWriteBuff,
		writeBuff:    comm.NewZeroCopySink(make([]byte, 0, MaxWriteBuff)),
		pendingBuff:  comm.NewZeroCopySink(make([]byte, 0, MaxWriteBuff)),
	}
	link.cond = sync.NewCond(&link.lock)

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
func (this *Link) SetChan(msgchan chan *types.MsgPayload) {
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
func (this *Link) GetConn() net.Conn {
	return this.conn
}

//set connection
func (this *Link) SetConn(conn net.Conn) {
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

func (this *Link) RunReadWriteLoop() {
	go this.writeLoop()
	go this.readLoop()
}

func (this *Link) readLoop() {
	conn := this.conn
	if conn == nil {
		return
	}

	reader := bufio.NewReaderSize(conn, common.MAX_BUF_LEN)

	for {
		this.lock.Lock()
		for this.err == nil && this.writeBuff.Size() > uint64(this.maxWriteBuff) {
			this.cond.Wait()
		}
		if this.err != nil {
			this.lock.Unlock()
			break
		}
		this.lock.Unlock()

		msg, payloadSize, err := types.ReadMessage(reader)
		if err != nil {
			this.lock.Lock()
			this.err = err
			this.disconnectNotify()
			this.cond.Signal() // notify write loop
			this.lock.Unlock()
			log.Infof("[p2p]error read from %s :%s", this.GetAddr(), err.Error())
			return
		}

		t := time.Now()
		this.UpdateRXTime(t)

		if !this.needSendMsg(msg) {
			log.Debugf("skip handle msgType:%s from:%d", msg.CmdType(), this.id)
			continue
		}
		this.addReqRecord(msg)
		this.recvChan <- &types.MsgPayload{
			Id:          this.id,
			Addr:        this.addr,
			PayloadSize: payloadSize,
			Payload:     msg,
		}
	}
}

//disconnectNotify push disconnect msg to channel
func (this *Link) disconnectNotify() {
	log.Debugf("[p2p]call disconnectNotify for %s", this.GetAddr())
	this.CloseConn()

	msg, _ := types.MakeEmptyMessage(common.DISCONNECT_TYPE)
	discMsg := &types.MsgPayload{
		Id:      this.id,
		Addr:    this.addr,
		Payload: msg,
	}
	this.recvChan <- discMsg
}

//close connection
func (this *Link) CloseConn() {
	if this.conn != nil {
		this.conn.Close()
		this.conn = nil
	}
}

func (this *Link) writeLoop() {
	conn := this.conn
	if conn == nil {
		return
	}

	for {
		this.lock.Lock()
		for this.err == nil && this.writeBuff.Size() == 0 {
			this.cond.Wait()
		}
		if this.err != nil {
			this.lock.Unlock()
			return
		}
		this.writeBuff.Reset()
		temp := this.writeBuff
		this.writeBuff = this.pendingBuff
		this.pendingBuff = temp
		this.lock.Unlock()

		payload := this.writeBuff.Bytes()
		nByteCnt := len(payload)
		log.Tracef("[p2p]TX buf length: %d\n", nByteCnt)

		nCount := nByteCnt / common.PER_SEND_LEN
		if nCount == 0 {
			nCount = 1
		}

		conn.SetWriteDeadline(time.Now().Add(time.Duration(nCount*common.WRITE_DEADLINE) * time.Second))
		_, err := conn.Write(payload)
		if err != nil {
			this.lock.Lock()
			this.err = err
			this.disconnectNotify()
			this.cond.Signal() // notify read loop
			this.lock.Unlock()

			log.Infof("[p2p]error sending messge to %s :%s", this.GetAddr(), err.Error())
			return
		}

		this.time = time.Now()

		this.cond.Signal() // notify read loop
	}
}

func (this *Link) Send(msg types.Message) {
	this.lock.Lock()
	types.WriteMessage(this.writeBuff, msg)
	this.lock.Unlock()
	this.cond.Broadcast()
}

func (this *Link) SendRaw(rawPacket []byte) {
	this.lock.Lock()
	this.writeBuff.WriteBytes(rawPacket)
	this.lock.Unlock()
	this.cond.Broadcast()
}

//needSendMsg check whether the msg is needed to push to channel
func (this *Link) needSendMsg(msg types.Message) bool {
	if msg.CmdType() != common.GET_DATA_TYPE {
		return true
	}
	var reqID = *msg.(*types.DataReq)
	now := time.Now().Unix()

	if t, ok := this.reqRecord[reqID]; ok {
		if int(now-t) < common.REQ_INTERVAL {
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
	now := time.Now().Unix()
	if len(this.reqRecord) >= common.MAX_REQ_RECORD_SIZE-1 {
		for id, t := range this.reqRecord {
			if int(now-t) > common.REQ_INTERVAL {
				delete(this.reqRecord, id)
			}
		}
	}
	var dataReq = *msg.(*types.DataReq)
	this.reqRecord[dataReq] = now
}
