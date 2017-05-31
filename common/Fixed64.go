package common

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
)

//the 64 bit fixed-point number, precise 10^-8
type Fixed64 int64

func (f *Fixed64) Serialize(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, int64(*f))
	if err != nil {
		return err
	}

	return nil
}

func (f *Fixed64) Deserialize(r io.Reader) error {
	p := make([]byte, 8)
	n, err := r.Read(p)
	if n <= 0 || err != nil {
		return err
	}
	b_buf := bytes.NewBuffer(p)
	var x int64
	err = binary.Read(b_buf, binary.LittleEndian, &x)
	if err != nil {
		return err
	}
	*f = Fixed64(x)
	return nil
}

func (f Fixed64) GetData() int64 {
	return int64(f)
}

func (f Fixed64) String() string {
	var buffer bytes.Buffer
	value := int64(f)
	if value < 0 {
		buffer.WriteRune('-')
		value = -value
	}
	buffer.WriteString(strconv.FormatInt(value/100000000, 10))
	value %= 100000000
	if value > 0 {
		buffer.WriteRune('.')
		s := strconv.FormatInt(value, 10)
		for i := len(s); i < 8; i++ {
			buffer.WriteRune('0')
		}
		buffer.WriteString(s)
	}
	return buffer.String()
}
