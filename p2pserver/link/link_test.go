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
	"math/rand"
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	comm "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	ct "github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/p2pserver/common"
	mt "github.com/ontio/ontology/p2pserver/message/types"
	"github.com/stretchr/testify/assert"
)

var (
	cliLink    *Link
	serverLink *Link
	cliChan    chan *mt.RecvMessage
	serverChan chan *mt.RecvMessage
	cliAddr    string
	serAddr    string
)

func init() {
	log.Init(log.Stdout)

	cliLink = NewLink()
	serverLink = NewLink()

	cliLink.id = 0x733936
	serverLink.id = 0x8274950

	cliLink.port = 50338
	serverLink.port = 50339

	cliChan = make(chan *mt.RecvMessage, 100)
	serverChan = make(chan *mt.RecvMessage, 100)
	//listen ip addr
	cliAddr = "127.0.0.1:50338"
	serAddr = "127.0.0.1:50339"

}

func TestNewLink(t *testing.T) {
    tspType := common.LegacyTSPType

	id := 0x74936295
	port := 40339

	if cliLink.GetID() != 0x733936 {
		t.Fatal("link GetID failed")
	}

	cliLink.SetID(uint64(id))
	if cliLink.GetID() != uint64(id) {
		t.Fatal("link SetID failed")
	}

	if cliLink.GetPort() != 50338 {
		t.Fatal("link GetPort failed")
	}

	cliLink.SetPort(uint16(port))
	if cliLink.GetPort() != uint16(port) {
		t.Fatal("link SetPort failed")
	}

	cliLink.SetChan(cliChan)
	serverLink.SetChan(serverChan)

	cliLink.UpdateRXTime(time.Now())

	msg := &mt.RecvMessage{
		TSPType:tspType,
		RecvMsgPayload:&mt.MsgPayload{
		Id:      cliLink.id,
		Addr:    cliLink.addr,
		Payload: &mt.NotFound{comm.UINT256_EMPTY},
		},
	}
	go func() {
		time.Sleep(5000000)
		cliChan <- msg
	}()

	timeout := time.NewTimer(time.Second)
	select {
	case <-cliLink.recvChan:
		t.Log("read data from channel")
	case <-timeout.C:
		timeout.Stop()
		t.Fatal("can`t read data from link channel")
	}

}

func TestUnpackBufNode(t *testing.T) {
	cliLink.SetChan(cliChan)

	msgType := "block"

	var msg mt.Message

	switch msgType {
	case "addr":
		var newaddrs []common.PeerAddr
		for i := 0; i < 10000000; i++ {
			newaddrs = append(newaddrs, common.PeerAddr{
				Time: time.Now().Unix(),
				ID:   uint64(i),
			})
		}
		var addr mt.Addr
		addr.NodeAddrs = newaddrs
		msg = &addr
	case "consensuspayload":
		acct := account.NewAccount("SHA256withECDSA")
		key := acct.PubKey()
		payload := mt.ConsensusPayload{
			Owner: key,
		}
		for i := 0; uint32(i) < 200000000; i++ {
			byteInt := rand.Intn(256)
			payload.Data = append(payload.Data, byte(byteInt))
		}

		msg = &mt.Consensus{payload}
	case "consensus":
		acct := account.NewAccount("SHA256withECDSA")
		key := acct.PubKey()
		payload := &mt.ConsensusPayload{
			Owner: key,
		}
		for i := 0; uint32(i) < 200000000; i++ {
			byteInt := rand.Intn(256)
			payload.Data = append(payload.Data, byte(byteInt))
		}
		consensus := mt.Consensus{
			Cons: *payload,
		}
		msg = &consensus
	case "blkheader":
		var headers []*ct.Header
		blkHeader := &mt.BlkHeader{}
		for i := 0; uint32(i) < 100000000; i++ {
			header := &ct.Header{}
			header.Height = uint32(i)
			header.Bookkeepers = make([]keypair.PublicKey, 0)
			header.SigData = make([][]byte, 0)
			headers = append(headers, header)
		}
		blkHeader.BlkHdr = headers
		msg = blkHeader
	case "tx":
		trn := &mt.Trn{}
		tx, _:= ct.TransactionFromRawBytes([]byte{0xe4,0x24,0x5d,0x83,0x60,0x7e,0x66,0x44,0xc3,0x60,0xb6,0x00,0x70,0x45,0x01,0x7b,0x5c,0x5d,0x89,0xd9,0xf0,0xf5,0xa9,0xc3,0xb3,0x78,0x01,0x01,0x8f,0x78,0x9c,0xc3})
		trn.Txn = tx
		msg = trn
	case "block":
		var blk ct.Block
		mBlk := &mt.Block{}
		header := ct.Header{}
		header.Height = uint32(1)
		header.Bookkeepers = make([]keypair.PublicKey, 0)
		header.SigData = make([][]byte, 0)
		blk.Header = &header

		blk.Transactions = make([]*ct.Transaction, 0)

		mBlk.Blk = &blk

		msg = mBlk
	}

	sink := comm.NewZeroCopySink(nil)
	err := mt.WriteMessage(sink, msg)
	assert.Nil(t, err)
}
