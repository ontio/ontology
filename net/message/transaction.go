package message

import (
	"GoOnchain/common"
	"GoOnchain/core/transaction"
	. "GoOnchain/net/protocol"
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

type dataReq struct {
	msgHdr
	dataType common.InventoryType
	hash     common.Uint256
}

// Transaction message
type trn struct {
	msgHdr
	// TBD
	txn  transaction.Transaction
	hash common.Uint256
}

func (msg trn) Handle(node Noder) error {
	common.Trace()
	fmt.Printf("RX TRX message\n")

	if !node.ExistedID(msg.hash) {
		node.AppendTxnPool(&(msg.txn))
	}
	return nil
}

func reqTxnData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.TRANSACTION
	// TODO handle the hash array case
	msg.hash = hash

	buf, _ := msg.Serialization()
	go node.Tx(buf)
	return nil
}

func (msg dataReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))

	//using serilization function
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *dataReq) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))
	// TODO
	return nil
}
