package message

import (
	"GoOnchain/common"
	"GoOnchain/core/ledger"
	. "GoOnchain/net/protocol"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"unsafe"
)

type headersReq struct {
	hdr msgHdr
	p   struct {
		len       uint8
		hashStart [HASHLEN]byte
		hashEnd   [HASHLEN]byte
	}
}

type blkHeader struct {
	hdr    msgHdr
	blkHdr []byte
}

func NewHeadersReq(n Noder) ([]byte, error) {
	var h headersReq

	// Fixme correct with the exactly request length
	h.p.len = 1
	buf := n.GetLedger().Blockchain.CurrentBlockHash()
	copy(h.p.hashStart[:], reverse(buf[:]))

	p := new(bytes.Buffer)
	err := binary.Write(p, binary.LittleEndian, &(h.p))
	if err != nil {
		fmt.Println("Binary Write failed at new headersReq")
		return nil, err
	}

	s := checkSum(p.Bytes())
	h.hdr.init("getheaders", s, uint32(len(p.Bytes())))

	m, err := h.Serialization()
	str := hex.EncodeToString(m)
	fmt.Printf("The message length is %d, %s\n", len(m), str)
	return m, err
}

func (msg headersReq) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg blkHeader) Verify(buf []byte) error {
	// TODO Verify the message Content
	err := msg.hdr.Verify(buf)
	return err
}

func (msg headersReq) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *headersReq) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	buf := bytes.NewBuffer(p)
	err := binary.Read(buf, binary.LittleEndian, msg)
	return err
}

func (msg blkHeader) Serialization() ([]byte, error) {
	var buf bytes.Buffer

	fmt.Printf("The size of messge is %d in serialization\n",
		uint32(unsafe.Sizeof(msg)))
	err := binary.Write(&buf, binary.LittleEndian, msg)

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func (msg *blkHeader) Deserialization(p []byte) error {
	fmt.Printf("The size of messge is %d in deserialization\n",
		uint32(unsafe.Sizeof(*msg)))

	err := msg.hdr.Deserialization(p)
	msg.blkHdr = p[MSGHDRLEN:]
	return err
}

func (msg headersReq) Handle(node Noder) error {
	common.Trace()
	// lock
	var starthash [HASHLEN]byte //[]common.Uint256
	var stophash [HASHLEN]byte  //common.Uint256
	starthash = msg.p.hashStart
	stophash = msg.p.hashEnd
	//FIXME if HeaderHashCount > 1
	buf, _ := NewHeaders(starthash, stophash) //(starthash[0], stophash)
	go node.LocalNode().Tx(buf)
	return nil
}

func (msg blkHeader) Handle(node Noder) error {
	common.Trace()
	// TBD
	return nil
}

func NewHeaders(starthash common.Uint256, stophash common.Uint256) ([]byte, error) {
	var msg blkHeader
	var count uint32 = 0
	var empty [HASHLEN]byte
	//FIXME need add error handle for GetBlockWithHash

	bkstart, _ := ledger.DefaultLedger.Store.GetBlock(starthash)
	startheight := bkstart.Blockdata.Height
	var stopheight uint32
	if stophash != empty {
		bkstop, _ := ledger.DefaultLedger.Store.GetBlock(starthash)
		stopheight = bkstop.Blockdata.Height
		count = startheight - stopheight
		if count >= 2000 {
			count = 2000
		}
	} else {
		count = 2000
	}
	// waiting for GetBlockWithHeight commit
	/*
		var blkheaders *ledger.Blockdata
		var i uint32
		tmpBuffer := bytes.NewBuffer([]byte{})
		for i = 1; i <= count; i++ {
			//FIXME need add error handle for GetBlockWithHeight
			bk, _ := ledger.DefaultLedger.Blockchain.GetBlockWithHeight(stopheight + i)
			blkheaders = bk.Blockdata
			blkheaders.Serialize(tmpBuffer)
			i++
		}
		msg.blkHdr = tmpBuffer.Bytes()
	*/
	msg.hdr.Magic = NETMAGIC
	cmd := "headers"
	copy(msg.hdr.CMD[0:7], cmd)
	b := new(bytes.Buffer)
	err := binary.Write(b, binary.LittleEndian, &(msg.blkHdr))
	if err != nil {
		fmt.Println("Binary Write failed at new Msg")
		return nil, err
	}
	s := sha256.Sum256(b.Bytes())
	s2 := s[:]
	s = sha256.Sum256(s2)
	buf := bytes.NewBuffer(s[:4])
	binary.Read(buf, binary.LittleEndian, &(msg.hdr.Checksum))
	msg.hdr.Length = uint32(len(b.Bytes()))
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
