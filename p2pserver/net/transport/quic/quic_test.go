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

package quic

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/ontio/ontology/common/log"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
	"testing"
)

var done chan struct{}
var done1 chan struct{}

func init() {
	log.Init(log.Stdout)
	done = make(chan struct{})
	done1 = make(chan struct{})
}

type messageTest struct {
	One int32
	Two int32
}

func startServer(t *testing.T) {

	tspT, _ := NewTransport()

	l, err := tspT.Listen(23798)
	if err != nil {
		t.Errorf("quicTsp.Listen happen error!, err:%s", err.Error())
	}

	conn, err := l.Accept()
	if err != nil {
		t.Errorf("error accepting, err:%s", err.Error())
	}

	for {
		reader, err := conn.GetRecvStream()

		if err != nil {
			t.Errorf("error GetReader, err:%s", err.Error())
		}

		go func(rs tsp.RecvStream) {
			for {
				msg := messageTest{}
				//msgBuf := make([]byte, 2)
				err = binary.Read(rs, binary.LittleEndian, &msg)
				//_, err = io.ReadFull(reader, msgBuf)
				if err != nil {
					t.Errorf("read message error, err:%s", err.Error())
				} else {
					log.Infof("Receive message, one=%d, two=%d", msg.One, msg.Two)
				}
			}
		}(reader)

	}

	close(done)
}

var cCon tsp.Connection
func startClient(t *testing.T) {

	tspT, _ := NewTransport()

	conn, err := tspT.Dial("127.0.0.1:23798")
	if err != nil {
		t.Errorf("Dial err:%s", err)
	}

	//conn.SetWriteDeadline(/*time.Now().Add(time.Duration(1*5) * time.Second)*/)

	cCon = conn

	mt := messageTest{100, 190}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, mt)
	if err != nil {
		t.Errorf("binary.Write failed: %s", err)
	}

	n,err := conn.Write("messageTest", buf.Bytes())
	n,err = conn.Write("messageTest", buf.Bytes())

	if err != nil || n != buf.Len() {
		t.Error("Send message error")
	}

	close(done1)

}

func keepHeat() {
	for {
		<- done1
		if cCon == nil {
			continue
		}

		mt := messageTest{100, 190}

		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, mt)
		if err != nil {
			fmt.Printf("binary.Write failed: %s", err)
		}
		cCon.Write("messageTest", buf.Bytes())
	}
}

func TestQuicTransport (t *testing.T) {
	go startServer(t)
	go startClient(t)
	go keepHeat()

	<- done
}
