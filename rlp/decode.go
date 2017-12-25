package rlp

import (
	"io"
	"reflect"
	"errors"
	"encoding/binary"
	"fmt"
	"math/big"
	"bytes"
	"strings"
	"bufio"
)

var (
	EOL = errors.New("rlp: end of list")
	ErrCanonSize      = errors.New("rlp: non-canonical size information")
	ErrValueTooLarge  = errors.New("rlp: value size exceeds available input length")
	ErrExpectedString = errors.New("rlp: expected String or Byte")
	ErrExpectedList   = errors.New("rlp: expected List")
	ErrCanonInt       = errors.New("rlp: non-canonical integer format")
	ErrElemTooLarge   = errors.New("rlp: element is larger than containing list")

	errNotAtEOL     = errors.New("rlp: call of ListEnd not positioned at EOL")
	errUintOverflow = errors.New("rlp: uint overflow")
	errNotInList    = errors.New("rlp: call of ListEnd outside of any list")
)

type Stream struct {
	r ByteReader
	remaining uint64
	limited bool
	uintbuf []byte
	kind Kind
	size uint64
	byteval byte
	kinderr error
	stack []listpos
}

func NewStream(r io.Reader, inputLimit uint64) *Stream {
	s := new(Stream)
	s.Reset(r, inputLimit)
	return s
}

func (s *Stream) Reset(r io.Reader, inputLimit uint64) {
	if inputLimit > 0 {
		s.remaining = inputLimit
		s.limited = true
	} else {
		switch br := r.(type) {
		case *bytes.Reader:
			s.remaining = uint64(br.Len())
			s.limited = true
		case *strings.Reader:
			s.remaining = uint64(br.Len())
			s.limited = true
		default:
			s.limited = false
		}
	}
	bufr, ok := r.(ByteReader)
	if !ok {
		bufr = bufio.NewReader(r)
	}
	s.r = bufr
	s.stack = s.stack[:0]
	s.size = 0
	s.kind = -1
	s.kinderr = nil
	if s.uintbuf == nil {
		s.uintbuf = make([]byte, 8)
	}
}

func (s *Stream) Raw() ([]byte, error) {
	kind, size, err := s.Kind()
	if err != nil {
		return nil, err
	}
	if kind == Byte {
		s.kind = -1
		return []byte{s.byteval}, nil
	}
	start := headSize(size)
	buf := make([]byte, uint64(start) + size)
	if err := s.readFull(buf[start:]); err != nil {
		return nil, err
	}
	if kind == String {
		putHead(buf, 0x80, 0xB7, size)
	} else {
		putHead(buf, 0xC0, 0xF7, size)
	}
	return buf, nil
}

func (s *Stream) Kind() (Kind, uint64, error) {
	var tos *listpos
	if len(s.stack) > 0 {
		tos = &s.stack[len(s.stack)-1]
	}
	if s.kind < 0 {
		s.kinderr = nil
		if tos != nil && tos.pos == tos.size {
			return 0, 0, EOL
		}
		s.kind, s.size, s.kinderr = s.readKind()
		if s.kinderr == nil {
			if tos == nil {
				if s.limited && s.size > s.remaining {
					s.kinderr = ErrValueTooLarge
				}
			} else {
				if s.size > tos.size - tos.pos {
					s.kinderr = ErrElemTooLarge
				}
			}
		}
	}
	return s.kind, s.size, s.kinderr
}

func (s *Stream) readKind() (kind Kind, size uint64, err error) {
	b, err := s.readByte()
	if err != nil {
		if len(s.stack) == 0 {
			switch err {
			case io.ErrUnexpectedEOF:
				err = io.EOF
			case ErrValueTooLarge:
				err = io.EOF
			}
		}
		return 0, 0, err
	}
	s.byteval = 0
	switch {
	case b < 0x80:
		s.byteval = b
		return Byte, 0, nil
	case b < 0xB8:
		return String, uint64(b - 0x80), nil
	case b < 0xC0:
		size, err = s.readUint(b - 0xB7)
		if err == nil && size < 56 {
			err = ErrCanonSize
		}
		return String, size, err
	case b < 0xF8:
		return List, uint64(b - 0xC0), nil
	default:
		size, err = s.readUint(b - 0xF7)
		if err == nil && size < 56 {
			err = ErrCanonSize
		}
		return List, size, err
	}
}

func (s *Stream) readUint(size byte) (uint64, error) {
	switch size {
	case 0:
		s.kind = -1
		return 0, nil
	case 1:
		b, err := s.readByte()
		return uint64(b), err
	default:
		start := int(8 - size)
		for i := 0; i < start; i++ {
			s.uintbuf[i] = 0
		}
		if err := s.readFull(s.uintbuf[start:]); err != nil {
			return 0, ErrCanonSize
		}
		if s.uintbuf[start] == 0 {
			return 0, ErrCanonSize
		}
		return binary.BigEndian.Uint64(s.uintbuf), nil
	}
}

func (s *Stream) readFull(buf []byte) (err error) {
	if err := s.willRead(uint64(len(buf))); err != nil {
		return err
	}
	var nn, n int
	for n < len(buf) && err == nil {
		nn, err =s.r.Read(buf[n:])
		n += nn
	}
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return err
}

func (s *Stream) readByte() (byte, error) {
	if err := s.willRead(1); err != nil {
		return 0, err
	}
	b, err := s.r.ReadByte()
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return b, err
}

func (s *Stream) willRead(n uint64) error {
	s.kind = -1
	if len(s.stack) > 0 {
		tos := s.stack[len(s.stack) - 1]
		if n > tos.size - tos.pos {
			return ErrValueTooLarge
		}
		s.stack[len(s.stack) - 1].pos += n
	}
	if s.limited {
		if n > s.remaining {
			return ErrValueTooLarge
		}
		s.remaining -= n
	}
	return nil
}

func (s *Stream) Bytes() ([]byte, error) {
	kind, size, err := s.Kind()
	if err != nil {
		return nil, err
	}
	switch kind {
	case Byte:
		s.kind = -1
		return []byte{s.byteval}, nil
	case String:
		b := make([]byte, size)
		if err = s.readFull(b); err != nil {
			return nil, err
		}
		if size == 1 && b[0] < 128 {
			return nil, ErrCanonSize
		}
		return b, nil
	default:
		return nil, ErrExpectedString
	}
}

func (s *Stream) Uint() (uint64, error) {
	return s.uint(64)
}

func (s *Stream) uint(maxbits int) (uint64, error) {
	kind, size, err := s.Kind()
	if err != nil {
		return 0, err
	}
	switch kind {
	case Byte:
		if s.byteval == 0 {
			return 0, ErrCanonInt
		}
		s.kind = -1
		return uint64(s.byteval), nil
	case String:
		if size >uint64(maxbits/8) {
			return 0, errUintOverflow
		}
		v, err := s.readUint(byte(size))
		switch {
		case err == ErrCanonSize:
			return 0, ErrCanonInt
		case err != nil:
			return 0, err
		case size > 0 && v < 128:
			return 0, ErrCanonSize
		default:
			return v, nil
		}
	default:
		return 0, ErrExpectedString
	}
}

func (s *Stream) Bool() (bool, error) {
	num, err := s.uint(8)
	if err != nil {
		return false, err
	}
	switch num {
	case 0:
		return false, nil
	case 1:
		return true, nil
	default:
		return false, fmt.Errorf("rlp: invalid boolean value: %d", num)
	}
}

func (s *Stream) List() (size uint64, err error) {
	kind, size, err := s.Kind()
	if err != nil {
		return 0, err
	}
	if kind != List {
		return 0, ErrExpectedList
	}
	s.stack = append(s.stack, listpos{0, size})
	s.kind = -1
	s.size = 0
	return size, nil
}

func (s *Stream) ListEnd() error {
	if len(s.stack) == 0 {
		return errNotInList
	}
	tos := s.stack[len(s.stack)-1]
	if tos.pos != tos.size {
		return errNotAtEOL
	}
	s.stack = s.stack[:len(s.stack)-1]
	if len(s.stack) > 0 {
		s.stack[len(s.stack)-1].pos += tos.size
	}
	s.kind = -1
	s.size = 0
	return nil
}

type listpos struct {
	pos, size uint64
}

type ByteReader interface {
	io.Reader
	io.ByteReader
}

var (
	bigInt           = reflect.TypeOf(big.Int{})
)

type Decoder interface {
	DecodeRLP(*Stream) error
}

func decodeSliceElems(s *Stream, val reflect.Value, elemDec decoder) error {
	i := 0
	for ; ; i++ {
		if i >= val.Cap() {
			newcap := val.Cap() + val.Cap() / 2
			if newcap < 4 {
				newcap = 4
			}
			newv := reflect.MakeSlice(val.Type(), val.Len(), newcap)
			reflect.Copy(newv, val)
			val.Set(newv)
		}
		if i >= val.Len() {
			val.SetLen(i + 1)
		}
		if err := elemDec(s, val.Index(i)); err == EOL {
			break
		} else if err != nil {
			return err
		}
	}
	if i < val.Len() {
		val.SetLen(i)
	}
	return nil
}

func decodeListSlice(s *Stream, val reflect.Value, elemdec decoder) error {
	size, err := s.List()
	if err != nil {
		return err
	}
	if size == 0 {
		val.Set(reflect.MakeSlice(val.Type(), 0, 0))
		return s.ListEnd()
	}
	if err := decodeSliceElems(s, val, elemdec); err != nil {
		return err
	}
	return s.ListEnd()
}

func decodeInterface(s *Stream, val reflect.Value) error {
	if val.Type().NumMethod() != 0 {
		return fmt.Errorf("[decodeInterface] type %v error.", val.Type())
	}
	kind, _, err := s.Kind()
	if err != nil {
		return err
	}
	if kind == List {
		slice := reflect.New(reflect.TypeOf([]interface{}{})).Elem()
		if err := decodeListSlice(s, slice, decodeInterface); err != nil {
			return err
		}
		val.Set(slice)
	} else {
		b, err := s.Bytes()
		if err != nil {
			return err
		}
		val.Set(reflect.ValueOf(b))
	}
	return nil
}
