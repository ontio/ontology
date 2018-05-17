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
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

// The RX buffer of this node to solve multiple packets problem
type RxBuf struct {
	p   []byte //buffer
	len int    //patload length in buffer
}

//Link used to establish
type Link struct {
	id       uint64
	addr     string    // The address of the node
	conn     net.Conn  // Connect socket with the peer node
	port     uint16    // The server port of the node
	rxBuf    RxBuf     // recv buffer
	time     time.Time // The latest time the node activity
	recvChan chan *common.MsgPayload
}

func NewLink() *Link {
	link := &Link{}

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
func (this *Link) SetChan(msgchan chan *common.MsgPayload) {
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

//record latest getting message time
func (this *Link) UpdateRXTime(t time.Time) {
	this.time = t
}

//UpdateRXTime return the latest message time
func (this *Link) GetRXTime() time.Time {
	return this.time
}

// Shrinking the buf to the exactly reading in byte length
//@Return @1 the start header of next message, the left length of the next message
func unpackNodeBuf(this *Link, buf []byte) {
	var msgLen int
	var msgBuf []byte

	if len(buf) == 0 {
		return
	}

	var rxBuf *RxBuf
	rxBuf = &this.rxBuf

	if rxBuf.len == 0 {
		length := common.MSG_HDR_LEN - len(rxBuf.p)
		if length > len(buf) {
			length = len(buf)
			rxBuf.p = append(rxBuf.p, buf[0:length]...)
			return
		}

		rxBuf.p = append(rxBuf.p, buf[0:length]...)
		if types.ValidMsgHdr(rxBuf.p) == false {
			rxBuf.p = nil
			rxBuf.len = 0
			log.Warn("Get error message header, TODO: relocate the msg header")
			// TODO Relocate the message header
			return
		}

		rxBuf.len = types.PayloadLen(rxBuf.p)
		buf = buf[length:]
	}

	msgLen = rxBuf.len
	if len(buf) == msgLen {
		msgBuf = append(rxBuf.p, buf[:]...)
		this.pushdata(msgBuf)
		rxBuf.p = nil
		rxBuf.len = 0
	} else if len(buf) < msgLen {
		rxBuf.p = append(rxBuf.p, buf[:]...)
		rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(rxBuf.p, buf[0:msgLen]...)
		this.pushdata(msgBuf)
		rxBuf.p = nil
		rxBuf.len = 0

		unpackNodeBuf(this, buf[msgLen:])
	}
}

//pushdata send packed data to channel
func (this *Link) pushdata(buf []byte) {
	p2pMsg := &common.MsgPayload{
		Id:      this.id,
		Addr:    this.addr,
		Payload: buf,
	}
	this.recvChan <- p2pMsg
}

//Rx read conn byte then call unpackNodeBuf to parse data
func (this *Link) Rx() {
	conn := this.conn
	buf := make([]byte, common.MAX_BUF_LEN)
	for {
		len, err := conn.Read(buf[0:(common.MAX_BUF_LEN - 1)])
		buf[common.MAX_BUF_LEN-1] = 0 //Prevent overflow
		switch err {
		case nil:
			t := time.Now()
			this.UpdateRXTime(t)
			unpackNodeBuf(this, buf[0:len])
		case io.EOF:
			//log.Error("Rx io.EOF: ", err, ", node id is ", node.GetID())
			goto DISCONNECT
		default:
			log.Error("Read connection error ", err)
			goto DISCONNECT
		}
	}

DISCONNECT:
	this.disconnectNotify()
}

//disconnectNotify push disconnect msg to channel
func (this *Link) disconnectNotify() {
	this.CloseConn()

	var m types.MsgCont
	cmd := common.DISCONNECT_TYPE
	m.Hdr.Magic = common.NETMAGIC
	copy(m.Hdr.CMD[0:uint32(len(cmd))], cmd)
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, m.Hdr)
	msgbuf := buf.Bytes()

	discMsg := &common.MsgPayload{
		Id:      this.id,
		Addr:    this.addr,
		Payload: msgbuf,
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

//Tx write data to link conn
func (this *Link) Tx(buf []byte) error {
	log.Debugf("TX buf length: %d\n%x", len(buf), buf)
	if this.conn == nil {
		return errors.New("tx link invalid")
	}
	_, err := this.conn.Write(buf)
	if err != nil {
		log.Error("Error sending messge to peer node ", err.Error())
		this.disconnectNotify()
		return err
	}

	return nil
}
