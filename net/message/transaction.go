package message

import (
	"GoOnchain/common"
	//"GoOnchain/core/transaction"
	. "GoOnchain/net/protocol"
	//"GoOnchain/core/ledger"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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
	txn []byte
	//txn  transaction.Transaction
	//hash common.Uint256
}

func (msg trn) Handle(node Noder) error {
	common.Trace()
	fmt.Printf("RX TRX message\n")
	/*
		if !node.ExistedID(msg.hash) {
			node.AppendTxnPool(&(msg.txn))
		}
	*/
	return nil
}

func reqTxnData(node Noder, hash common.Uint256) error {
	var msg dataReq
	msg.dataType = common.TRANSACTION
	// TODO handle the hash array case
	//msg.hash = hash

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

func NewTx(hash common.Uint256) ([]byte, error) {
	common.Trace()
	var msg trn
	//Wait for junjie commit GetTransactionWithHash!!!!
	/*
		trx, _ := ledger.DefaultLedger.Blockchain.GetTransactionWithHash(hash)
		txBuffer := bytes.NewBuffer([]byte{})
		trx.Serialize(txBuffer)
		msg.txn = txBuffer.Bytes()
	*/
	msg.msgHdr.Magic = NETMAGIC
	cmd := "tx"
	copy(msg.msgHdr.CMD[0:7], cmd)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, &(msg.txn))
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.msgHdr.Checksum))
	msg.msgHdr.Length = uint32(len(b.Bytes()))
	fmt.Printf("The message payload length is %d\n", msg.msgHdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		fmt.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, nil
}

func (msg trn) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}
