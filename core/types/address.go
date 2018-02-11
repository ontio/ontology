package types

import (
	"errors"
	"fmt"
	"io"

)

const AddrLen = 20

type Address [AddrLen]byte


func (self *Address) ToHexString() string {
	return fmt.Sprintf("%x", self[:])
}

func (self *Address) Serialize(w io.Writer) error {
	_, err := w.Write(self[:])
	return err
}

func (self *Address) Deserialize(r io.Reader) error {
	n, err := r.Read(self[:])
	if n != len(self[:]) || err != nil {
		return errors.New("deserialize Address error")
	}
	return nil
}
