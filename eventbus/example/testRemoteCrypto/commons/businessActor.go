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

package commons

import (
	"bytes"
	"fmt"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/actor"

	"github.com/Ontology/eventbus/eventhub"
	"strconv"
	"sync"
	"time"
)

const (
	SigTOPIC    string = "SIGTOPIC"
	VerifyTOPIC string = "VERIFYTOPIC"
	SetTOPIC    string = "SETTOPIC"
)

const loop1 = 1
const Loop2 = 2500

type BusynessActor struct {
	Datas      map[string][]byte
	privatekey []byte
	pubkey     crypto.PubKey
	start      int64
	respCount  int
	WgStop     *sync.WaitGroup
}

func (s *BusynessActor) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *RunMsg:
		fmt.Println("Recieve runMsg")
		crypto.SetAlg("")
		bb := bytes.NewBuffer([]byte(""))
		for i := 0; i < 400000; i++ {
			bb.WriteString("1234567890")
		}

		privKey, pubkey, _ := crypto.GenKeyPair()

		s.privatekey = privKey
		s.pubkey = pubkey

		setPrivMsg := &SetPrivKey{PrivKey: privKey}

		setEvent := &eventhub.Event{Topic: "SETTOPIC", Publisher: context.Self(), Message: setPrivMsg, Policy: eventhub.PUBLISH_POLICY_ALL}

		eventhub.GlobalEventHub.Publish(setEvent)

		s.start = time.Now().UnixNano()
		for i := 0; i < loop1; i++ {
			idx := strconv.Itoa(i)
			bb.WriteString(idx)
			data := bb.Bytes()
			sigMsg := &SignRequest{Seq: idx, Data: data}
			s.Datas[idx] = data
			sigEvent := &eventhub.Event{Topic: "SIGTOPIC", Publisher: context.Self(), Message: sigMsg, Policy: eventhub.PUBLISH_POLICY_ROUNDROBIN}
			eventhub.GlobalEventHub.Publish(sigEvent)
		}

	case *SignResponse:
		seq := msg.Seq
		sig := msg.Signature
		//fmt.Printf("seq:%s,sig:%v\n",seq,sig)

		buf := bytes.NewBuffer([]byte(""))
		err := s.pubkey.Serialize(buf)
		if err != nil {
			fmt.Println("ERROR Serialize publickey: ", err)
		}
		pubKeyBytes := buf.Bytes()

		vfr := &VerifyRequest{Signature: sig, Data: s.Datas[seq], PublicKey: pubKeyBytes, Seq: seq}
		//vfActor.Request(vfr,context.Self())
		vrfEvent := &eventhub.Event{Topic: "VERIFYTOPIC", Publisher: context.Self(), Message: vfr, Policy: eventhub.PUBLISH_POLICY_ROUNDROBIN}
		start := time.Now()
		for i := 0; i < Loop2; i++ {
			eventhub.GlobalEventHub.Publish(vrfEvent)
		}
		elapsed := time.Since(start)
		fmt.Printf("Elapsed %s\n", elapsed)

	case *VerifyResponse:
		s.respCount++
		if s.respCount%100 == 0 {
			fmt.Println(s.respCount)
		}
		if s.respCount == Loop2 {
			s.WgStop.Done()
		}

		//seq := msg.Seq
		//result := msg.Result
		//errmsg := msg.ErrorMsg
		//
		//if !result {
		//	fmt.Printf("seq:%s faild pass,err:%s\n", seq, errmsg)
		//} else {
		//	fmt.Printf("seq:%s passed verify\n", seq)
		//}

	default:
		fmt.Printf("unknown msg %v\n", msg)
	}
}
