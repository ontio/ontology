package common

import (
	"bytes"
	"io"
	"strconv"
)

//the 64 bit fixed-point number, precise 10^-8
type Fixed64 int64

func (f *Fixed64) Serialize(w io.Writer) {
	//TODO: implement Fixed64.serialize
}

func (f *Fixed64) Deserialize(r io.Reader) error {
	//TODO：Fixed64 Deserialize

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
