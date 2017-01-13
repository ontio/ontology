package vm

import (
	"bytes"
	"encoding/binary"
)

type VmReader struct {
	reader     *bytes.Reader
	BaseStream []byte
}

func NewVmReader(b []byte) *VmReader {
	var vmreader VmReader
	vmreader.reader = bytes.NewReader(b)
	vmreader.BaseStream = b
	return &vmreader
}

func (r *VmReader) ReadByte() byte {
	byte, _ := r.reader.ReadByte()
	return byte
}

func (r *VmReader) ReadBytes(count int) []byte {
	var bytes []byte
	for i := 0; i < count; i++ {
		bytes = append(bytes, r.ReadByte())
	}
	return bytes
}

func (r *VmReader) ReadUint16() uint16 {
	b := r.ReadBytes(2)
	return binary.LittleEndian.Uint16(b)
}

func (r *VmReader) ReadUInt32() uint32 {
	b := r.ReadBytes(4)
	return binary.LittleEndian.Uint32(b)
}

func (r *VmReader) ReadUInt64() uint64 {
	b := r.ReadBytes(8)
	return binary.LittleEndian.Uint64(b)
}

func (r *VmReader) ReadInt16() int16 {
	b := r.ReadBytes(2)
	bytesBuffer := bytes.NewBuffer(b)
	var vi int16
	binary.Read(bytesBuffer, binary.BigEndian, &vi)
	return vi

}

func (r *VmReader) ReadInt32() int32 {
	b := r.ReadBytes(4)
	bytesBuffer := bytes.NewBuffer(b)
	var vi int32
	binary.Read(bytesBuffer, binary.BigEndian, &vi)
	return vi
}

func (r *VmReader) Position() int {
	return int(r.reader.Size()) - r.reader.Len()
}

func (r *VmReader) Length() int {
	return r.reader.Len()
}

func (r *VmReader) Seek(offset int64, whence int) (int64, error) {
	return r.reader.Seek(offset, whence)
}


func (r *VmReader) ReadVarBytes(max int) []byte {

	n := int(r.ReadVarInt(uint64(max)))
	return r.ReadBytes(n)

}

func (r *VmReader) ReadVarInt(max uint64) uint64 {

	fb := r.ReadByte()
	var value uint64

	switch fb {
	case 0xFD:
		value = uint64(r.ReadInt16())
	case 0xFE:
		value = uint64(r.ReadUInt32())
	case 0xFF:
		value = uint64(r.ReadUInt64())
	default:
		value = uint64(fb)
	}

	if value > max {
		return 0
	}

	return value
}

func (r *VmReader) ReadVarString() string{

	bs := r.ReadVarBytes(0X7fffffc7)
	return string(bs)

	//return Encoding.UTF8.GetString(reader.ReadVarBytes());
}

