package common

import (
	"io"
	"bytes"
	"encoding/binary"
)

type Uint160 [20]uint8

func (u *Uint160) Serialize(w io.Writer) (int,error) {
	b_buf := bytes.NewBuffer([]byte{})
	binary.Write(b_buf, binary.LittleEndian, u)

	len, err := w.Write( b_buf.Bytes() )

	if err != nil {
		return 0, err
	}

	return len, nil
}

func (f *Uint160) Deserialize(r io.Reader) error {
	p := make([]byte, 20)
	n, err := r.Read(p)

	if n <= 0 || err != nil {
		return err
	}

	b_buf := bytes.NewBuffer(p)
	binary.Read(b_buf, binary.LittleEndian, f)

	return nil
}