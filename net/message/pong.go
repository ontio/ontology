package message

import (
	. "DNA/net/protocol"
)

type pong struct {
	msgHdr
	Nonce uint64
}

func NewPongMsg() ([]byte, error) {
	var msg pong
	var sum []byte
	sum = []byte{0x5d, 0xf6, 0xe0, 0xe2}
	msg.msgHdr.init("pong", sum, 0)

	buf, err := msg.Serialization()
	if err != nil {
		return nil, err
	}
	return buf, err
}

func (msg pong) Verify(buf []byte) error {
	err := msg.msgHdr.Verify(buf)
	// TODO verify the message Content
	return err
}

func (msg pong) Handle(node Noder) error {
	node.SetLastContact()
	return nil
}
