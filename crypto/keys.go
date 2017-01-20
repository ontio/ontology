package crypto

import (
	"io"
)

type PubKey ECPoint

func (e *PubKey) Serialize(w io.Writer) {
	//TODO: implement PubKey.serialize
}

func (ep *PubKey) EncodePoint(commpressed bool) []byte{
	//TODO: EncodePoint
	return nil
}