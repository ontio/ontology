package common

import (
	"io"
)

type Uint256 [32]uint8

func (u *Uint256) Serialize(w io.Writer) {
	//TODO: implement Uint256.serialize
}

func (u *Uint256) Deserialize(r io.Reader) error {
	//TODO：Uint256 Deserialize

	return nil
}
