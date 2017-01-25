package crypto

import (
	"io"
)

type PubKey ECPoint

func NewPubKey(prikey []byte) *PubKey{
	//TODO: NewPubKey
	return nil
}

func (e *PubKey) Serialize(w io.Writer) {
	//TODO: implement PubKey.serialize
}

func (ep *PubKey) EncodePoint(commpressed bool) []byte{
	//TODO: EncodePoint
	return nil
}


func DecodePoint(encoded []byte) *PubKey{
	//TODO: DecodePoint
	return nil
}


type PubKeySlice []*PubKey

func (p PubKeySlice) Len() int           { return len(p) }
func (p PubKeySlice) Less(i, j int) bool {
	//TODO:PubKeySlice Less
	return false
}
func (p PubKeySlice) Swap(i, j int) {
	//TODO:PubKeySlice Swap
}



