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
	"bytes"
	"errors"
	"net"
	"time"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
)

//Link used to establish
type Link struct {
	id       uint64
	addr     string    // The address of the node
	conn     net.Conn  // Connect socket with the peer node
	port     uint16    // The server port of the node
	time     time.Time // The latest time the node activity
	recvChan chan *types.MsgPayload
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

func (this *Link) Rx() {
	reader := bufio.NewReaderSize(this.conn, common.MAX_BUF_LEN)

	for {
		msg, err := types.ReadMessage(reader)
		if err != nil {
			log.Error("read connection error ", err)
			break
		}

		t := time.Now()
		this.UpdateRXTime(t)
		this.recvChan <- &types.MsgPayload{
			Id:      this.id,
			Addr:    this.addr,
			Payload: msg,
		}

	}

	this.disconnectNotify()
}

//disconnectNotify push disconnect msg to channel
func (this *Link) disconnectNotify() {
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

func (this *Link) Tx(msg types.Message) error {
	conn := this.conn
	if conn == nil {
		return errors.New("tx link invalid")
	}
	buf := bytes.NewBuffer(nil)
	err := types.WriteMessage(buf, msg)
	if err != nil {
		log.Error("error serialize messge ", err.Error())
	}
	log.Debugf("TX buf length: %d\n", len(buf.Bytes()))

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		log.Error("error sending messge to peer node ", err.Error())
		this.disconnectNotify()
		return err
	}

	return nil
}
