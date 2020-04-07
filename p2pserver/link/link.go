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
	"time"

	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

//Link used to establish
type Link struct {
	id        uint64
	addr      string                 // The address of the node
	conn      net.Conn               // Connect socket with the peer node
	port      uint16                 // The server port of the node
	time      time.Time              // The latest time the node activity
	recvChan  chan *types.MsgPayload //msgpayload channel
	reqRecord map[string]int64       //Map RequestId to Timestamp, using for rejecting duplicate request in specific time
	stopCh    chan struct{}          // close connection indicator
}

func NewLink() *Link {
	link := &Link{
		reqRecord: make(map[string]int64, 0),
		stopCh:    make(chan struct{}),
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
	select {
	case <-this.stopCh:
		return false
	default:
	}
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
	// fresh connection, fresh stopCh
	this.stopCh = make(chan struct{})
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

func (this *Link) Rx() {
	conn := this.conn
	if conn == nil {
		return
	}

	reader := bufio.NewReaderSize(conn, common.MAX_BUF_LEN)

loop:
	for {
		select {
		case <-this.stopCh:
			break loop
		default:
		}

		msg, payloadSize, err := types.ReadMessage(reader)
		if err != nil {
			log.Infof("[p2p]error read from %s :%s", this.GetAddr(), err.Error())
			break
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

	this.disconnectNotify()
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

func (l *Link) CloseLink() {
	select {
	case <-l.stopCh:
		return
	default:
	}
	// if we can reach here, the channel is not closed
	close(l.stopCh)

	// we don't close socket here, we wait Rx to get out from Read
	// then we will break Rx loop, if Rx sleep on Read for a long time
	// the p2pserver.retryInactivePeer will try reconnect(because the link state is not ESTABLISH)
	// when:
	// 1. reconnect success ==> VersionHandle will delete old peer ==> put old peer to Np.trashCh
	//    ==> close after 10 minutes
	// 2. reconnect fail ==> after retry ntimes ==> delete from Np(NbrPeers) ==> put to Np.trashCh
	//    ==> close after 10 minutes
	// so the Peer *in* netserver.Np is promise to close when error happend.
	// but we have to close peer's underline socket immediately if error happend *before*
	// inserting peer to netserver.Np
	// because no one manager this connection, after all this peer is already a useless peer
	// we are unlikely to tranfer any data via this socket
}

func (this *Link) Send(msg types.Message) error {
	sink := comm.NewZeroCopySink(nil)
	types.WriteMessage(sink, msg)

	return this.SendRaw(sink.Bytes())
}

func (this *Link) SendRaw(rawPacket []byte) error {
	conn := this.conn
	if conn == nil {
		return errors.New("[p2p]tx link invalid")
	}

	nByteCnt := len(rawPacket)
	log.Tracef("[p2p]TX buf length: %d\n", nByteCnt)

	nCount := nByteCnt / common.PER_SEND_LEN
	if nCount == 0 {
		nCount = 1
	}
	conn.SetWriteDeadline(time.Now().Add(time.Duration(nCount*common.WRITE_DEADLINE) * time.Second))
	_, err := conn.Write(rawPacket)
	if err != nil {
		log.Infof("[p2p]error sending messge to %s :%s", this.GetAddr(), err.Error())
		this.disconnectNotify()
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
