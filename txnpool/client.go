package main

import (
	//"fmt"
	"bytes"
	"encoding/hex"
	"github.com/Ontology/common"
	"github.com/Ontology/common/log"
	"github.com/Ontology/core/payload"
	"github.com/Ontology/core/types"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/remote"
	"github.com/Ontology/eventbus/zmqremote"
	"time"
)

var (
	tx *types.Transaction
)

func init() {
	log.Init(log.Path, log.Stdout)

	bookKeepingPayload := &payload.BookKeeping{
		Nonce: uint64(time.Now().UnixNano()),
	}

	tx = &types.Transaction{
		Version:    0,
		Attributes: []*types.TxAttribute{},
		TxType:     types.BookKeeping,
		Payload:    bookKeepingPayload,
	}

	tempStr := "3369930accc1ddd067245e8edadcd9bea207ba5e1753ac18a51df77a343bfe92"
	hex, _ := hex.DecodeString(tempStr)
	var hash common.Uint256
	hash.Deserialize(bytes.NewReader(hex))
	tx.SetHash(hash)
}

func main() {
	var stopCh chan bool

	stopCh = make(chan bool)

	remote.Start("192.168.153.130:50011")
	server := actor.NewPID("192.168.153.130:50010", "Txn")
	/*props := actor.FromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *tp.CheckTxnRsp:
			fmt.Println(msg)
		}
	})*/

	//client := actor.Spawn(props)
	tmpBuffer := bytes.NewBuffer([]byte{})
	tx.Serialize(tmpBuffer)
	server.Tell(&zmqremote.MsgData{MsgType: 1, Data: tmpBuffer.Bytes()})
	<-stopCh
}
