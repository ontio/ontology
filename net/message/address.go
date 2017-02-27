package message

import (
	"GoOnchain/common"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"unsafe"
)

type addrReq struct {
	Hdr msgHdr
	// No payload
}

type addr struct {
	hdr       msgHdr
	nodeCnt   uint64
	nodeAddrs []NodeAddr
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

func NewAddrs(nodeaddrs []NodeAddr, count uint64) ([]byte, error) {
	var msg addr
	msg.nodeAddrs = nodeaddrs
	msg.nodeCnt = count
	msg.hdr.Magic = NETMAGIC
	cmd := "addr"
	copy(msg.hdr.CMD[0:7], cmd)
	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(msg.nodeAddrs))
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(p.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(p.Bytes()))
	fmt.Printf("The message payload length is %d\n", msg.hdr.Length)

	m, err := msg.Serialization()
	if err != nil {
		fmt.Println("Error Convert net message ", err.Error())
		return nil, err
	}

	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, nil
}

func (msg addrReq) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.Hdr.Verify(buf)
	return err
}

func (msg addrReq) Handle(node Noder) error {
	common.Trace()
	// lock
	var addrstr []NodeAddr
	var count uint64
	addrstr, count = node.GetAddrs()
	buf, _ := NewAddrs(addrstr, count)
	go node.Tx(buf)
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
	var buf bytes.Buffer
	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *addr) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.hdr.Deserialization(p)

	fmt.Printf("The address buffer len is %d\n", len(p))
	// Fixme Call the serilization package
	cnt, i := common.GetCompactUint(p[MSGHDRLEN:])
	msg.nodeCnt = cnt
	fmt.Printf("The address count is %d i is %d\n", cnt, i)
	buf := p[MSGHDRLEN+i:]
	msg.nodeAddrs = make([]NodeAddr, msg.nodeCnt)
	for i := 0; i < int(msg.nodeCnt); i++ {
		nodeAddr := &msg.nodeAddrs[i]
		err = nodeAddr.Deserialization(buf[i*NODEADDRSIZE : (i+1)*NODEADDRSIZE])
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
		if v.Port != 0 {
			var ip net.IP
			ip = v.IpAddr[:]
			// Fixme consider the IPv6 case
			address := ip.To4().String() + ":" + strconv.Itoa(int(v.Port))
			fmt.Printf("The ip address is %s\n", address)
			go node.Connect(address)
		}
	}
	return nil
}
