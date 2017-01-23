package common

import (
	"io"
	"bytes"
	"encoding/binary"
	_ "fmt"
)

type Uint256 [32]uint8

func (u *Uint256) Serialize(w io.Writer) (int,error) {
	b_buf := bytes.NewBuffer([]byte{})
	binary.Write(b_buf, binary.LittleEndian, u)

	len, err := w.Write( b_buf.Bytes() )

	if err != nil {
		return 0, err
	}

	return len, nil
}

func (u *Uint256) Deserialize(r io.Reader) error {
	p := make([]byte, 32)
	n, err := r.Read(p)

	if n <= 0 || err != nil {
		return err
	}

	b_buf := bytes.NewBuffer(p)
	binary.Read(b_buf, binary.LittleEndian, u)

	return nil
}
