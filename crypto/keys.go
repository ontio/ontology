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


type PubKeys []*PubKey

// Len is the number of elements in the collection.
func (ep *PubKeys) Len() int{
	return -1
}
// Less reports whether the element with
// index i should sort before the element with index j.
func (ep *PubKeys) Less(i, j int) bool {
	return false
}
// Swap swaps the elements with indexes i and j.
func (ep *PubKeys) Swap(i, j int){

}

