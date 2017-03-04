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
	err := binary.Write(p, binary.LittleEndian, msg.nodeCnt)
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}

	err = binary.Write(p, binary.LittleEndian, msg.nodeAddrs)
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
	addrstr, count = node.LocalNode().GetNeighborAddrs()
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
	err := binary.Write(&buf, binary.LittleEndian, msg.hdr)

	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.LittleEndian, msg.nodeCnt)
	if err != nil {
		return nil, err
	}
	for _, v := range msg.nodeAddrs {
		//err = binary.Write(&buf, binary.LittleEndian, v.Serialization)
		err = binary.Write(&buf, binary.LittleEndian, v)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), err
}

func (msg *addr) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, &(msg.hdr))
	//err := msg.hdr.Deserialization(p)
	err = binary.Read(buf, binary.LittleEndian, &(msg.nodeCnt))
	//err = binary.Read(p[MSGHDRLEN:p[MSGHDRLEN + 8], binary.LittleEndian, &cnt)
	fmt.Printf("The address count is %d \n", msg.nodeCnt)
	msg.nodeAddrs = make([]NodeAddr, msg.nodeCnt)
	for i := 0; i < int(msg.nodeCnt); i++ {
		err := binary.Read(buf, binary.LittleEndian, &(msg.nodeAddrs[i]))
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
		var ip net.IP
		ip = v.IpAddr[:]
		// Fixme consider the IPv6 case
		address := ip.To4().String() + ":" + strconv.Itoa(int(v.Port))
		fmt.Printf("The ip address is %s id is 0x%x\n", address, v.ID)

		if (v.ID == node.LocalNode().GetID()) {
			continue
		}
		if node.LocalNode().NodeEstablished(v.ID) {
			continue
		}

		if v.Port == 0 {
			continue
		}

		go node.LocalNode().Connect(address)
	}
	return nil
}
