package message

import (
	. "GoOnchain/net/protocol"
)

type memPool struct {
	msgHdr
	//TBD
}

func ReqMemoryPool(node Noder) error {
	msg := AllocMsg("mempool", 0)
	buf, _ := msg.Serialization()
	go node.Tx(buf)

	return nil
}
