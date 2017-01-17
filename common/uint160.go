package common

import (
	"io"
)

type Uint160 [20]uint8

func (u *Uint160) Serialize(w io.Writer) {
	//TODO: implement Uint160.serialize
}

func (f *Uint160) Deserialize(r io.Reader) error {
	//TODO：Uint160 Deserialize
	return nil
}
