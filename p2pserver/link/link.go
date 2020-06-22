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
	"net"
	"sync"
	"sync/atomic"
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

//Link used to establish
type Link struct {
	id   common.PeerId
	addr string // The address of the node

	lock      sync.RWMutex
	conn      net.Conn               // Connect socket with the peer node
	time      int64                  // The latest time the node activity
	recvChan  chan *types.MsgPayload //msgpayload channel
	reqRecord map[string]int64       //Map RequestId to Timestamp, using for rejecting duplicate request in specific time
}

func NewLink(id common.PeerId, c net.Conn, msgChan chan *types.MsgPayload) *Link {
	link := &Link{
		id:        id,
		addr:      c.RemoteAddr().String(),
		conn:      c,
		time:      time.Now().UnixNano(),
		recvChan:  msgChan,
		reqRecord: make(map[string]int64),
	}

	return link
}

//get address
func (this *Link) GetAddr() string {
	return this.addr
}

//get connection
func (this *Link) GetConn() net.Conn {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.conn
}

//set connection
func (this *Link) SetConn(conn net.Conn) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.conn = conn
}

//record latest message time
func (this *Link) UpdateRXTime(t time.Time) {
	atomic.StoreInt64(&this.time, t.UnixNano())
}

//GetRXTime return the latest message time
func (this *Link) GetRXTime() int64 {
	return atomic.LoadInt64(&this.time)
}

func (this *Link) Rx() {
	conn := this.GetConn()
	if conn == nil {
		return
	}

	reader := bufio.NewReaderSize(conn, common.MAX_BUF_LEN)

	for {
		msg, payloadSize, err := types.ReadMessage(reader)
		if err != nil {
			log.Infof("[p2p]error read from %s :%s", this.GetAddr(), err.Error())
			break
		}

		if unknown, ok := msg.(*types.UnknownMessage); ok {
			log.Infof("skip handle unknown msg type:%s from:%d", unknown.CmdType(), this.id)
			continue
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

	this.CloseConn()
}

//close connection
func (this *Link) CloseConn() {
	this.lock.Lock()
	conn := this.conn
	this.conn = nil
	this.lock.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
}

func (this *Link) Send(msg types.Message) error {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)

	return this.SendRaw(sink.Bytes())
}

func (this *Link) SendRaw(rawPacket []byte) error {
	conn := this.GetConn()
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
		log.Infof("[p2p] error sending messge to %s :%s", this.GetAddr(), err.Error())
		this.CloseConn()
		return err
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
		for id := range this.reqRecord {
			t := this.reqRecord[id]
			if int(now-t) > common.REQ_INTERVAL {
				delete(this.reqRecord, id)
			}
		}
	}
	var dataReq = msg.(*types.DataReq)
	reqID := fmt.Sprintf("%x%s", dataReq.DataType, dataReq.Hash.ToHexString())
	this.reqRecord[reqID] = now
}
