package message

import (
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unsafe"
)

type addrReq struct {
	Hdr msgHdr
	// No payload
}

type nodeAddr struct {
	time     uint32
	services uint64
	ipAddr   [16]byte
	port     uint16
}

type addr struct {
	hdr       msgHdr
	nodeCnt   uint64
	nodeAddrs []nodeAddr
}

const (
	NODEADDRSIZE = 30
)

func newGetAddr() ([]byte, error) {
	var msg addrReq
	// Fixme the check is the []byte{0} instead of 0
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.Hdr.init("getaddr", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}

	str := hex.EncodeToString(buf)
	fmt.Printf("The message get addr length is %d, %s\n", len(buf), str)

	return buf, err
}

func (msg addrReq) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg addrReq) Handle(node Noder) error {
	common.Trace()
	// TBD

	return nil
}

func (msg addrReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *addrReq) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg addr) Serialization() ([]byte, error) {

	// TBD
	//return buf.Bytes(), err
	return nil, nil
}

func (msg *nodeAddr) deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg *addr) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.hdr.Deserialization(p)

	fmt.Printf("The address buffer len is %d\n", len(p))
	// Fixme Call the serilization package
	cnt, i := common.GetCompactUint(p[MSGHDRLEN:])
	msg.nodeCnt = cnt
	fmt.Printf("The address count is %d\n", cnt)
	buf := p[MSGHDRLEN+i:]
	msg.nodeAddrs = make([]nodeAddr, msg.nodeCnt)
	for i := 0; i < int(msg.nodeCnt); i++ {
		nodeAddr := &msg.nodeAddrs[i]
		err = nodeAddr.deserialization(buf[i*NODEADDRSIZE : (i+1)*NODEADDRSIZE])
		if err != nil {
			goto err
		}
	}
err:
	return err
}

func (msg addr) Verify(buf []byte) error {
	err := msg.hdr.Verify(buf)
	// TODO Verify the message Content, check the ipaddr number
	return err
}

func (msg addr) Handle(node Noder) error {
	common.Trace()
	for _, v := range msg.nodeAddrs {
		if v.port != 0 {
			// TODO Convert the ipaddress to string
			node.Connect(hex.EncodeToString(v.ipAddr[:]))
		}
	}
	return nil
}
