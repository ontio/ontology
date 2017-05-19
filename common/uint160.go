package common

import (
	. "DNA/errors"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"github.com/tv42/base58"
	"io"
	"math/big"
)

const UINT160SIZE int = 20

type Uint160 [UINT160SIZE]uint8

func (u *Uint160) CompareTo(o Uint160) int {
	x := u.ToArray()
	y := o.ToArray()

	for i := len(x) - 1; i >= 0; i-- {
		if x[i] > y[i] {
			return 1
		}
		if x[i] < y[i] {
			return -1
		}
	}

	return 0
}

func (u *Uint160) ToArray() []byte {
	var x []byte = make([]byte, UINT160SIZE)
	for i := 0; i < 20; i++ {
		x[i] = byte(u[i])
	}

	return x
}

func (u *Uint160) Serialize(w io.Writer) (int, error) {
	b_buf := bytes.NewBuffer([]byte{})
	binary.Write(b_buf, binary.LittleEndian, u)

	len, err := w.Write(b_buf.Bytes())

	if err != nil {
		return 0, err
	}

	return len, nil
}

func (f *Uint160) Deserialize(r io.Reader) error {
	p := make([]byte, UINT160SIZE)
	n, err := r.Read(p)

	if n <= 0 || err != nil {
		return err
	}

	b_buf := bytes.NewBuffer(p)
	binary.Read(b_buf, binary.LittleEndian, f)

	return nil
}

func (f *Uint160) ToAddress() string {
	data := append([]byte{23}, f.ToArray()...)
	temp := sha256.Sum256(data)
	temps := sha256.Sum256(temp[:])
	data = append(data, temps[0:4]...)

	bi := new(big.Int)
	bi.SetBytes(data)
	var dst []byte
	dst = base58.EncodeBig(dst, bi)

	return string(dst[:])
}

func Uint160ParseFromBytes(f []byte) (Uint160, error) {
	if len(f) != UINT160SIZE {
		return Uint160{}, NewDetailErr(errors.New("[Common]: Uint160ParseFromBytes err, len != 20"), ErrNoCode, "")
	}

	var hash [20]uint8
	for i := 0; i < 20; i++ {
		hash[i] = f[i]
	}
	return Uint160(hash), nil
}
