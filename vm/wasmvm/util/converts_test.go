package util

import (
	"testing"
	"math"
	"encoding/binary"
)


func TestFloat32ToByte(t *testing.T) {
	f := float32(3.1415926)
	b := Float32ToBytes(f)

	if ByteToFloat32(b) != f{
		t.Error("TestFloat32ToByte failed!")
	}

}

func TestFloat64ToByte(t *testing.T) {
	f := 3.1415926
	b := Float64ToBytes(f)

	if ByteToFloat64(b) != f{
		t.Error("TestFloat64ToByte failed!")
	}
}

func TestInt32ToBytes(t *testing.T) {
	i := math.MaxInt32
	b := Int32ToBytes(uint32(i))
	if int(binary.LittleEndian.Uint32(b)) != i {
		t.Error("TestInt32ToBytes failed!")
	}
}

func TestInt64ToBytes(t *testing.T) {
	i := math.MaxInt64
	b := Int64ToBytes(uint64(i))
	if int(binary.LittleEndian.Uint64(b)) != i {
		t.Error("TestInt32ToBytes failed!")
	}
}

func TestTrimBuffToString(t *testing.T) {
	s := "helloworld"
	b := []byte(s)
	if TrimBuffToString(b) != s{
		t.Error("TestTrimBuffToString failed")
	}
	b = append(b,byte(0))
	if TrimBuffToString(b) != s{
		t.Error("TestTrimBuffToString failed")
	}
	b = append(b,[]byte("some other string")...)
	if TrimBuffToString(b) != s{
		t.Error("TestTrimBuffToString failed")
	}
}