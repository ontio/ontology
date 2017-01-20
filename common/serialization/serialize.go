package serialization

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
)

var ErrRange = errors.New("value out of range")
var ErrEof = errors.New("got EOF, can not get the next byte")

//SerializableData describe the data need be serialized.
type SerializableData interface {

	//Write data to writer
	Serialize(w io.Writer)

	//read data to reader
	Deserialize(r io.Reader)
}

func WriteDataList(w io.Writer, list []SerializableData) error {
	len := uint64(len(list))
	WriteVarUint(w, len)

	for _, data := range list {
		data.Serialize(w)
	}

	return nil
}

/*
 ******************************************************************************
 * public func for outside calling
 ******************************************************************************
 * 1. WriteVarUint func, depend on the inpute number's Actual number size,
 *    serialize to bytes.
 *      uint8  =>  (LittleEndian)num in 1 byte                 = 1bytes
 *      uint16 =>  0xfd(1 byte) + (LittleEndian)num in 2 bytes = 3bytes
 *      uint32 =>  0xfe(1 byte) + (LittleEndian)num in 4 bytes = 5bytes
 *      uint64 =>  0xff(1 byte) + (LittleEndian)num in 8 bytes = 9bytes
 * 2. ReadVarUint  func, this func will read the first byte to determined
 *    the num length to read.and retrun the uint64
 *      first byte = 0xfd, read the next 2 bytes as uint16
 *      first byte = 0xfe, read the next 4 bytes as uint32
 *      first byte = 0xff, read the next 8 bytes as uint64
 *      other else,        read this byte as uint8
 * 3. WriteVarBytes func, this func will output two item as serialization.
 *      length of bytes (uint8/uint16/uint32/uint64)  +  bytes
 * 4. WriteVarString func, this func will output two item as serialization.
 *      length of string(uint8/uint16/uint32/uint64)  +  bytes(string)
 * 5. ReadVarBytes func, this func will first read a uint to identify the
 *    length of bytes, and use it to get the next length's bytes to return.
 * 6. ReadVarString func, this func will first read a uint to identify the
 *    length of string, and use it to get the next bytes as a string.
 * 7. GetVarUintSize func, this func will return the length of a uint when it
 *    serialized by the WriteVarUint func.
 * 8. ReadBytes func, this func will read the specify lenth's bytes and retun.
 ******************************************************************************
 */
func WriteVarUint(writer io.Writer, value uint64) error {
	b_buf := new(bytes.Buffer)
	if value < 0xFD {
		valx := uint8(value)
		err := binary.Write(b_buf, binary.LittleEndian, valx)
		if err != nil {
			return err
		}
	} else if value <= 0xFFFF {
		err := b_buf.WriteByte(0xFD)
		if err != nil {
			return err
		}
		valx := uint16(value)
		err = binary.Write(b_buf, binary.LittleEndian, valx)
		if err != nil {
			return err
		}
	} else if value <= 0xFFFFFFFF {
		err := b_buf.WriteByte(0xFE)
		if err != nil {
			return err
		}
		valx := uint32(value)
		err = binary.Write(b_buf, binary.LittleEndian, valx)
		if err != nil {
			return err
		}
	} else {
		err := b_buf.WriteByte(0xFF)
		if err != nil {
			return err
		}
		valx := uint64(value)
		err = binary.Write(b_buf, binary.LittleEndian, valx)
		if err != nil {
			return err
		}
	}
	_, err := writer.Write(b_buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

func ReadVarUint(r io.Reader, maxint uint64) (uint64, error) {
	if maxint == 0x00 {
		maxint = math.MaxUint64
	}
	fb, _ := byteReader(r)
	if bytes.Equal(fb, []byte{byte(0xfd)}) {
		value, err := readerUint16(r)
		if err != nil {
			return 0, err
		}
		if value > maxint {
			return 0, ErrRange
		}
		return value, nil
	} else if bytes.Equal(fb, []byte{byte(0xfe)}) {
		value, err := readerUint32(r)
		if err != nil {
			return 0, err
		}
		if value > maxint {
			return 0, ErrRange
		}
		return value, nil
	} else if bytes.Equal(fb, []byte{byte(0xff)}) {
		value, err := readerUint64(r)
		if err != nil {
			return 0, err
		}
		if value > maxint {
			return 0, ErrRange
		}
		return value, nil
	} else {
		value, err := readerUint8(fb)
		if err != nil {
			return 0, err
		}
		if value > maxint {
			return 0, ErrRange
		}
		return value, nil
	}

	return 0, nil
}

func WriteVarBytes(writer io.Writer, value []byte) error {
	err := WriteVarUint(writer, uint64(len(value)))
	if err != nil {
		return err
	}
	_, err = writer.Write(value)
	if err != nil {
		return err
	}
	return nil
}

func WriteVarString(writer io.Writer, value string) error {
	err := WriteVarUint(writer, uint64(len(value)))
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(value))
	if err != nil {
		return err
	}
	return nil
}

func ReadVarBytes(reader io.Reader) ([]byte, error) {
	val, err := ReadVarUint(reader, 0)
	if err != nil {
		return nil, err
	}
	str, err := byteXReader(reader, val)
	if err != nil {
		return nil, err
	}
	return str, nil
}

func ReadVarString(reader io.Reader) (string, error) {
	val, err := ReadVarBytes(reader)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func GetVarUintSize(value uint64) int {
	if value < 0xfd {
		return binary.Size(uint8(0xff))
	} else if value <= 0xffff {
		return binary.Size(uint16(0xffff)) + binary.Size(uint8(0xff))
	} else if value <= 0xFFFFFFFF {
		return binary.Size(uint32(0xffffffff)) + binary.Size(uint8(0xff))
	} else {
		return binary.Size(uint64(0xffffffffffffffff)) + binary.Size(uint8(0xff))
	}
}

func ReadBytes(reader io.Reader, length uint64) ([]byte, error) {
	str, err := byteXReader(reader, length)
	if err != nil {
		return nil, err
	}
	return str, nil
}

//**************************************************************************
//**    internal func                                                    ***
//**************************************************************************
func byteReader(reader io.Reader) ([]byte, error) {
	p := make([]byte, 1)
	n, err := reader.Read(p)
	if n > 0 {
		return p[:], nil
	}
	return p, err
}

func byteXReader(reader io.Reader, x uint64) ([]byte, error) {
	p := make([]byte, x)
	n, err := reader.Read(p)
	if n > 0 {
		return p[:], nil
	}
	return p, err
}

func readerUint8(bytex []byte) (uint64, error) {
	b_buf := bytes.NewBuffer(bytex)
	var x uint8
	err := binary.Read(b_buf, binary.LittleEndian, &x)
	if err != nil {
		return 0, err
	}
	return uint64(x), nil
}

func readerUint16(reader io.Reader) (uint64, error) {
	p := make([]byte, 2)
	n, err := reader.Read(p)
	if n <= 0 || err != nil {
		return 0, ErrEof
	}
	b_buf := bytes.NewBuffer(p)
	var x uint16
	err = binary.Read(b_buf, binary.LittleEndian, &x)
	if err != nil {
		return 0, err
	}
	return uint64(x), nil
}

func readerUint32(reader io.Reader) (uint64, error) {
	p := make([]byte, 4)
	n, err := reader.Read(p)
	if n <= 0 || err != nil {
		return 0, ErrEof
	}
	b_buf := bytes.NewBuffer(p)
	var x uint32
	err = binary.Read(b_buf, binary.LittleEndian, &x)
	if err != nil {
		return 0, err
	}
	return uint64(x), nil
}

func readerUint64(reader io.Reader) (uint64, error) {
	p := make([]byte, 8)
	n, err := reader.Read(p)
	if n <= 0 || err != nil {
		return 0, ErrEof
	}
	b_buf := bytes.NewBuffer(p)
	var x uint64
	err = binary.Read(b_buf, binary.LittleEndian, &x)
	if err != nil {
		return 0, err
	}
	return x, nil
}

func ReadDataList(reader io.Reader) ([]SerializableData, error) {

	return nil,nil
}

func ToArray(data SerializableData) []byte {
	//TODO: ToArray()
	return nil
}