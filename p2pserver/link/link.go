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

// The RX buffer of this node to solve mutliple packets problem
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
	recvChan chan common.MsgPayload
}

func NewLink() *Link {
	link := &Link{}

	return link
}

//SetID set peer id to link
func (link *Link) SetID(id uint64) {
	link.id = id
}

//GetID return if from peer
func (link *Link) GetID() uint64 {
	return link.id
}

//If there is connection return true
func (link *Link) Valid() bool {
	return link.conn != nil
}

//set message channel for link layer
func (link *Link) SetChan(msgchan chan common.MsgPayload) {
	link.recvChan = msgchan
}

//get address
func (link *Link) GetAddr() string {
	return link.addr
}

//set address
func (link *Link) SetAddr(addr string) {
	link.addr = addr
}

//set port number
func (link *Link) SetPort(p uint16) {
	link.port = p
}

//get port number
func (link *Link) GetPort() uint16 {
	return link.port
}

//get connection
func (link *Link) GetConn() net.Conn {
	return link.conn
}

//set connection
func (link *Link) SetConn(conn net.Conn) {
	link.conn = conn
}

//record latest getting message time
func (link *Link) UpdateRXTime(t time.Time) {
	link.time = t
}

//UpdateRXTime return the latest message time
func (link *Link) GetRXTime() time.Time {
	return link.time
}

// Shrinking the buf to the exactly reading in byte length
//@Return @1 the start header of next message, the left length of the next message
func unpackNodeBuf(link *Link, buf []byte) {
	var msgLen int
	var msgBuf []byte

	if len(buf) == 0 {
		return
	}

	var rxBuf *RxBuf
	rxBuf = &link.rxBuf

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
		link.pushdata(msgBuf)
		rxBuf.p = nil
		rxBuf.len = 0
	} else if len(buf) < msgLen {
		rxBuf.p = append(rxBuf.p, buf[:]...)
		rxBuf.len = msgLen - len(buf)
	} else {
		msgBuf = append(rxBuf.p, buf[0:msgLen]...)
		link.pushdata(msgBuf)
		rxBuf.p = nil
		rxBuf.len = 0

		unpackNodeBuf(link, buf[msgLen:])
	}
}

//pushdata send packed data to channel
func (link *Link) pushdata(buf []byte) {
	p2pMsg := common.MsgPayload{
		Id:      link.id,
		Addr:    link.addr,
		Payload: buf,
	}
	link.recvChan <- p2pMsg
}

//Rx read conn byte then call unpackNodeBuf to parse data
func (link *Link) Rx() {
	conn := link.conn
	buf := make([]byte, common.MAX_BUF_LEN)
	for {
		len, err := conn.Read(buf[0:(common.MAX_BUF_LEN - 1)])
		buf[common.MAX_BUF_LEN-1] = 0 //Prevent overflow
		switch err {
		case nil:
			t := time.Now()
			link.UpdateRXTime(t)
			unpackNodeBuf(link, buf[0:len])
		case io.EOF:
			//log.Error("Rx io.EOF: ", err, ", node id is ", node.GetID())
			goto DISCONNECT
		default:
			log.Error("Read connection error ", err)
			goto DISCONNECT
		}
	}

DISCONNECT:
	link.disconnectNotify()
}

//disconnectNotify push disconnect msg to channel
func (link *Link) disconnectNotify() {
	link.CloseConn()

	var m types.MsgCont
	cmd := common.DISCONNECT_TYPE
	m.Hdr.Magic = common.NETMAGIC
	copy(m.Hdr.CMD[0:uint32(len(cmd))], cmd)
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, m.Hdr)
	msgbuf := buf.Bytes()

	discMsg := common.MsgPayload{
		Id:      link.id,
		Addr:    link.addr,
		Payload: msgbuf,
	}
	link.recvChan <- discMsg
}

//close connection
func (link *Link) CloseConn() {
	if link.conn != nil {
		link.conn.Close()
		link.conn = nil
	}
}

//Tx write data to link conn
func (link *Link) Tx(buf []byte) error {
	log.Debugf("TX buf length: %d\n%x", len(buf), buf)
	if link.conn == nil {
		return errors.New("tx link invalid")
	}
	_, err := link.conn.Write(buf)
	if err != nil {
		log.Error("Error sending messge to peer node ", err.Error())
		link.disconnectNotify()
		return err
	}

	return nil
}
