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
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync/atomic"
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

const SEND_THROTTLE_SIZE = 512 * 1024

var ErrBufferFull = errors.New("send buffers full")

//Link used to establish
type Link struct {
	id        common.PeerId
	addr      string                 // The address of the node
	conn      net.Conn               // Connect socket with the peer node
	time      time.Time              // The latest time the node activity
	recvChan  chan *types.MsgPayload //msgpayload channel
	reqRecord map[string]int64       //Map RequestId to Timestamp, using for rejecting duplicate request in specific time

	sendBuffer *SendBuffer
}

type buffData struct {
	data   []byte
	result chan error
}

type SendBuffer struct {
	ThrottleSize uint64 // read only once set

	bufferSize int64 // atomic read/write
	buffers    *LockFreeList
}

func (self *SendBuffer) Close() {
	self.buffers.TakeAndSeal()
	atomic.StoreInt64(&self.bufferSize, 0)
}

// return true if exceed throttle size
func (self *SendBuffer) IncrBuffSize(size int) bool {
	newVal := atomic.AddInt64(&self.bufferSize, int64(size))
	return newVal > int64(size)+SEND_THROTTLE_SIZE
}

func (self *SendBuffer) TryPush(packet []byte) error {
	if self.IncrBuffSize(len(packet)) {
		self.IncrBuffSize(-len(packet))
		return ErrBufferFull
	}
	if !self.buffers.Push(buffData{data: packet, result: nil}) {
		self.IncrBuffSize(-len(packet))
		return io.ErrClosedPipe
	}

	return nil
}

// blocking until data writen to io
func (self *SendBuffer) Push(packet []byte) error {
	result := make(chan error)
	self.IncrBuffSize(len(packet))
	if !self.buffers.Push(buffData{data: packet, result: result}) {
		self.IncrBuffSize(-len(packet))
		return io.ErrClosedPipe
	}

	return <-result
}

func NewLink(id common.PeerId, conn net.Conn) *Link {
	link := &Link{
		id:         id,
		sendBuffer: &SendBuffer{ThrottleSize: SEND_THROTTLE_SIZE, buffers: &LockFreeList{}},
		reqRecord:  make(map[string]int64),
		time:       time.Now(),
		conn:       conn,
		addr:       conn.RemoteAddr().String(),
	}

	return link
}

//GetID return if from peer
func (self *Link) GetID() common.PeerId {
	return self.id
}

//If there is connection return true
func (self *Link) Valid() bool {
	return self.conn != nil
}

//set message channel for link layer
func (self *Link) SetChan(msgchan chan *types.MsgPayload) {
	self.recvChan = msgchan
}

//get address
func (self *Link) GetAddr() string {
	return self.addr
}

//get connection
func (self *Link) GetConn() net.Conn {
	return self.conn
}

//set connection
func (self *Link) SetConn(conn net.Conn) {
	self.conn = conn
}

//GetRXTime return the latest message time
func (self *Link) GetRXTime() time.Time {
	return self.time
}

func (self *Link) StartReadWriteLoop() {
	go self.readLoop()
	go self.sendLoop()
}

func (self *Link) readLoop() {
	conn := self.conn
	if conn == nil {
		return
	}

	reader := bufio.NewReaderSize(conn, common.MAX_BUF_LEN)

	for {
		msg, payloadSize, err := types.ReadMessage(reader)
		if err != nil {
			log.Infof("[p2p]error read from %s :%s", self.GetAddr(), err.Error())
			break
		}

		self.time = time.Now()

		if !self.needSendMsg(msg) {
			log.Debugf("skip handle msgType:%s from:%d", msg.CmdType(), self.id)
			continue
		}

		self.addReqRecord(msg)
		self.recvChan <- &types.MsgPayload{
			Id:          self.id,
			Addr:        self.addr,
			PayloadSize: payloadSize,
			Payload:     msg,
		}

	}

	self.CloseConn()
}

//close connection
func (self *Link) CloseConn() {
	self.sendBuffer.Close()
	if self.conn != nil {
		_ = self.conn.Close()
		self.conn = nil
	}
}

func (self *Link) Send(msg types.Message) error {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)

	return self.SendRaw(sink.Bytes())
}

func (self *Link) TrySendRaw(packet []byte) error {
	return self.sendBuffer.TryPush(packet)
}

func (self *Link) TrySend(msg types.Message) error {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)
	return self.TrySendRaw(sink.Bytes())
}

func (self *Link) SendRaw(rawPacket []byte) error {
	return self.sendBuffer.Push(rawPacket)
}

// only called by sendLoop
func (self *Link) writeToConn(rawPacket []byte) error {
	conn := self.conn
	if conn == nil {
		return errors.New("[p2p]tx link invalid")
	}
	nByteCnt := len(rawPacket)
	log.Tracef("[p2p]TX buf length: %d\n", nByteCnt)

	nCount := nByteCnt / common.PER_SEND_LEN
	if nCount == 0 {
		nCount = 1
	}
	_ = conn.SetWriteDeadline(time.Now().Add(time.Duration(nCount*common.WRITE_DEADLINE) * time.Second))
	_, err := conn.Write(rawPacket)
	if err != nil {
		log.Infof("[p2p] error sending messge to %s :%s", self.GetAddr(), err.Error())
		self.CloseConn()
		return err
	}

	return nil
}

//needSendMsg check whether the msg is needed to push to channel
func (self *Link) needSendMsg(msg types.Message) bool {
	if msg.CmdType() != common.GET_DATA_TYPE {
		return true
	}
	var dataReq = msg.(*types.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.ToHexString())
	now := time.Now().Unix()

	if t, ok := self.reqRecord[reqID]; ok {
		if int(now-t) < common.REQ_INTERVAL {
			return false
		}
	}
	return true
}

//addReqRecord add request record by removing outdated request records
func (self *Link) addReqRecord(msg types.Message) {
	if msg.CmdType() != common.GET_DATA_TYPE {
		return
	}
	now := time.Now().Unix()
	if len(self.reqRecord) >= common.MAX_REQ_RECORD_SIZE-1 {
		for id := range self.reqRecord {
			t := self.reqRecord[id]
			if int(now-t) > common.REQ_INTERVAL {
				delete(self.reqRecord, id)
			}
		}
	}
	var dataReq = msg.(*types.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.ToHexString())
	self.reqRecord[reqID] = now
}

const sendBufSize = 64 * 1024

func (self *Link) sendLoop() {
	buffers := make([]byte, 0, sendBufSize)
	var results []chan error
	buffList := make([]buffData, 0, 64)
	for {
		owned, sealed := self.sendBuffer.buffers.Take()
		if sealed {
			return
		}
		buffList = getBufferData(buffList[:0], owned)
		if len(buffList) > 0 {
			for i := len(buffList) - 1; i >= 0; i -= 1 {
				buffers = append(buffers, buffList[i].data...)
				if buffList[i].result != nil {
					results = append(results, buffList[i].result)
				}
				if len(buffers) >= sendBufSize/2 || i == 0 {
					err := self.writeToConn(buffers)
					self.sendBuffer.IncrBuffSize(-len(buffers))
					for _, c := range results {
						c <- err
					}
					if err != nil {
						return
					}
					buffers = buffers[:0]
					results = results[:0]
				}
			}
		} else {
			// no buffer has been taken, yield this goroutine to avoid busy loop
			runtime.Gosched()
		}
	}
}

func getBufferData(buffList []buffData, owned *OwnedList) []buffData {
	for buf := owned.Pop(); buf != nil; buf = owned.Pop() {
		buffList = append(buffList, buf.(buffData))
	}

	return buffList
}
