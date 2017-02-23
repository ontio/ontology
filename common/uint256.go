package common

import (
	"io"
	"bytes"
	"encoding/binary"
	_ "fmt"
	."GoOnchain/errors"
	"errors"
)

type Uint256 [32]uint8

func (u *Uint256) CompareTo( o Uint256 ) int {
	x := u.ToArray()
	y := o.ToArray()

	for i:=len(x)-1; i>=0; i-- {
		if ( x[i] > y[i] ) {
			return 1
		}
		if ( x[i] < y[i] ) {
			return -1
		}
	}

	return 0
}

func (u *Uint256) ToArray() []byte {
	var x []byte = make([]byte,32)
	for i:=0; i<32; i++ {
		x[i] = byte(u[i])
	}

	return x
}

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

func (u *Uint256) ToString() string {
	return string(u.ToArray())
}

func Uint256ParseFromBytes(f []byte) (Uint256,error){
	if ( len(f) != 32 ) {
		return Uint256{},NewDetailErr(errors.New("[Common]: Uint256ParseFromBytes err, len != 32"), ErrNoCode, "");
	}

	var hash [32]uint8
	for i:=0; i<32; i++ {
		hash[i] = f[i]
	}
	return Uint256(hash),nil
}
