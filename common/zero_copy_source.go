package common

import (
	"encoding/binary"
)

type ZeroCopySource struct {
	s   []byte
	off uint64 // current reading index
}

// Len returns the number of bytes of the unread portion of the
// slice.
func (self *ZeroCopySource) Len() uint64 {
	length := uint64(len(self.s))
	if self.off >= length {
		return 0
	}
	return length - self.off
}

func (self *ZeroCopySource) Pos() uint64 {
	return self.off
}

// Size returns the original length of the underlying byte slice.
// Size is the number of bytes available for reading via ReadAt.
// The returned value is always the same and is not affected by calls
// to any other method.
func (self *ZeroCopySource) Size() uint64 { return uint64(len(self.s)) }

// Read implements the io.ZeroCopySource interface.
func (self *ZeroCopySource) NextBytes(n uint64) (data []byte, eof bool) {
	m := uint64(len(self.s))
	end, overflow := SafeAdd(self.off, n)
	if overflow || end > m{
		end = m
		eof = true
	}
	data = self.s[self.off:end]
	self.off = end

	return
}

func (self *ZeroCopySource) Skip(n uint64) (eof bool) {
	m := uint64(len(self.s))
	end, overflow := SafeAdd(self.off, n)
	if overflow || end > m {
		end = m
		eof = true
	}
	self.off = end

	return
}

// ReadByte implements the io.ByteReader interface.
func (self *ZeroCopySource) NextByte() (data byte, eof bool) {
	if self.off >= uint64(len(self.s)) {
		return 0, true
	}

	b := self.s[self.off]
	self.off++
	return b, false
}


// Backs up a number of bytes, so that the next call to NextXXX() returns data again
// that was already returned by the last call to NextXXX().
func (self *ZeroCopySource) BackUp(n uint64) {
	self.off -= n
}

func (self *ZeroCopySource) NextUint16() (data uint16, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(2)
	if eof {
		return
	}

	return binary.LittleEndian.Uint16(buf), eof
}

func (self *ZeroCopySource) NextUint32() (data uint32, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(4)
	if eof {
		return
	}

	return binary.LittleEndian.Uint32(buf), eof
}

func (self *ZeroCopySource) NextUint64() (data uint64, eof bool) {
	var buf []byte
	buf, eof = self.NextBytes(8)
	if eof {
		return
	}

	return binary.LittleEndian.Uint64(buf), eof
}

func (self *ZeroCopySource) NextInt16() (data int16, eof bool) {
	var val uint16
	val, eof = self.NextUint16()
	return int16(val), eof

}

func (self *ZeroCopySource) NextVarBytes() (data []byte, size uint64, eof bool) {
	var count uint64
	count, size, eof = self.NextVarUint()

	data, eof = self.NextBytes(count)

	return data, size + count, eof
}

func (self *ZeroCopySource) NextVarUint() (data uint64, size uint64, eof bool) {
	var fb byte
	fb, eof = self.NextByte()
	if eof {
		return
	}

	switch fb {
	case 0xFD:
		val, e := self.NextUint16()
		if e {
			return
		}
		data = uint64(val)
		size = 3
	case 0xFE:
		val, e := self.NextUint32()
		if e {
			return
		}
		data = uint64(val)
		size = 5
	case 0xFF:
		val, e := self.NextUint64()
		if e {
			return
		}
		data = uint64(val)
		size = 9
	default:
		data = uint64(fb)
		size = 1
	}

	return
}

// NewReader returns a new ZeroCopySource reading from b.
func NewZeroCopySource(b []byte) *ZeroCopySource { return &ZeroCopySource{b, 0} }