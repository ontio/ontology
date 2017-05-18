package message

import (
	"DNA/common/log"
	. "DNA/net/protocol"
)

type ping struct {
	msgHdr
	// No payload
}

func NewPingMsg() ([]byte, error) {
	var msg ping
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("ping", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}
	return buf, err
}

func (msg ping) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg ping) Handle(node Noder) error {
	buf, err := NewPongMsg()
	if err != nil {
		log.Error("failed build a new ping message")
	} else {
		go node.Tx(buf)
	}
	return err
}
