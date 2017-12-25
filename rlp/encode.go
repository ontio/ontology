package rlp

import (
	"sync"
	"io"
	"reflect"
	"fmt"
	"math/big"
	"errors"
)

var (
	encoderInterface = reflect.TypeOf(new(Encoder)).Elem()
	big0 = big.NewInt(0)
)

type Encoder interface {
	EncodeRLP(io.Writer) error
}

type encBuf struct {
	str     []byte
	lHeads  []*listHead
	lHSize  int
	sizeBuf []byte
}

func Encode(w io.Writer, val interface{}) error {
	if outer, ok := w.(*encBuf); ok {
		return outer.encode(val)
	}
	eb := encBufPool.Get().(*encBuf)
	defer encBufPool.Put(eb)
	eb.reset()
	if err := eb.encode(val); err != nil {
		return err
	}
	return eb.toWriter(w)
}

func (w *encBuf) encode(val interface{}) error {
	rVal := reflect.ValueOf(val)
	ti, err := cachedTypeInfo(rVal.Type(), tags{})
	if err != nil {
		return err
	}
	return ti.writer(rVal, w)
}

func (w *encBuf) Write(b []byte) (int, error) {
	w.str = append(w.str, b...)
	return len(b), nil
}

func (w *encBuf) toWriter(out io.Writer) (err error) {
	strops := 0
	for _, head := range w.lHeads {
		if head.offset - strops > 0 {
			n, err := out.Write(w.str[strops:head.offset])
			strops += n
			if err != nil {
				return err
			}
		}
		enc := head.encode(w.sizeBuf)
		if _, err = out.Write(enc); err != nil {
			return err
		}
	}
	if strops < len(w.str) {
		_, err = out.Write(w.str[strops:])
	}
	return
}

func (w *encBuf) size() int {
	return len(w.str) + w.lHSize
}

func (w *encBuf) toBytes() []byte {
	out := make([]byte, w.size())
	strPos := 0
	pos := 0
	for _, head := range w.lHeads {
		n := copy(out[pos:], w.str[strPos: head.offset])
		pos += n
		strPos += n
		enc := head.encode(out[pos:])
		pos += len(enc)
	}
	copy(out[pos:], w.str[strPos:])
	return out
}

func (w *encBuf) encodeString(b []byte) {
	if len(b) == 1 && b[0] <= 0x7F {
		w.str = append(w.str, b[0])
	} else {
		w.encodeStringHeader(len(b))
		w.str = append(w.str, b...)
	}
}

func (w *encBuf) encodeStringHeader(size int) {
	if size < 56 {
		w.str = append(w.str, 0x80 + byte(size))
	} else {
		sizeSize := putInt(w.sizeBuf[1:], uint64(size))
		w.sizeBuf[0] = 0xB7 + byte(sizeSize)
		w.str = append(w.str, w.sizeBuf[:sizeSize + 1]...)
	}
}

func (w *encBuf) listEnd(lh *listHead) {
	lh.size = w.size() - lh.offset - lh.size
	if lh.size < 56 {
		w.lHSize += 1
	} else {
		w.lHSize += 1 + intSize(uint64(lh.size))
	}
}

func (w *encBuf) list() *listHead {
	lh := &listHead{offset: len(w.str), size: w.lHSize}
	w.lHeads = append(w.lHeads, lh)
	return lh
}

func (w *encBuf) reset() {
	w.lHSize = 0
	if w.str != nil {
		w.str = w.str[:0]
	}
	if w.lHeads != nil {
		w.lHeads = w.lHeads[:0]
	}
}

func EncodeToBytes(val interface{}) ([]byte, error) {
	eb := encBufPool.Get().(*encBuf)
	defer encBufPool.Put(eb)
	eb.reset()
	if err := eb.encode(val); err != nil {
		return nil, err
	}
	return eb.toBytes(), nil
}

func makeWriter(typ reflect.Type, ts tags) (writer, error) {
	kind := typ.Kind()
	switch {
	case typ == rawValueType:
		return writeRawValue, nil
	case typ.Implements(encoderInterface):
		return writeEncoder, nil
	case kind != reflect.Ptr && reflect.PtrTo(typ).Implements(encoderInterface):
		return writeEncoderNoPtr, nil
	case kind == reflect.Interface:
		return writeInterface, nil
	case typ.AssignableTo(reflect.PtrTo(bigInt)):
		return writeBigIntPtr, nil
	case typ.AssignableTo(bigInt):
		return writeBigIntNoPtr, nil
	case isUint(kind):
		return writeUint, nil
	case kind == reflect.Bool:
		return writeBool, nil
	case kind == reflect.String:
		return writeString, nil
	case kind == reflect.Slice && isByte(typ.Elem()):
		return writeBytes, nil
	case kind == reflect.Array && isByte(typ.Elem()):
		return writeByteArray, nil
	case kind == reflect.Slice || kind == reflect.Array:
		return makeSliceWriter(typ, ts)
	case kind == reflect.Struct:
		return makeStructWriter(typ)
	case kind == reflect.Ptr:
		return makePtrWriter(typ)
	default:
		return nil, fmt.Errorf("[makeWriter] invalid type :%v", typ)
	}
}

func writeRawValue(val reflect.Value, w *encBuf) error {
	w.str = append(w.str, val.Bytes()...)
	return nil
}

func writeEncoder(val reflect.Value, w *encBuf) error {
	return val.Interface().(Encoder).EncodeRLP(w)
}

func writeEncoderNoPtr(val reflect.Value, w *encBuf) error {
	if !val.CanAddr() {
		return fmt.Errorf("[writeEncoderNoPtr] unadressable valud of type %v, EncodeRLP is pointer methid", val.Type())
	}
	return val.Addr().Interface().(Encoder).EncodeRLP(w)
}

func writeInterface(val reflect.Value, w *encBuf) error {
	if val.IsNil() {
		w.str = append(w.str, 0xC0)
		return nil
	}
	eval := val.Elem()
	ti, err := cachedTypeInfo(eval.Type(), tags{})
	if err != nil {
		return err
	}
	return ti.writer(eval, w)
}

func writeBigIntPtr(val reflect.Value, w *encBuf) error {
	ptr := val.Interface().(*big.Int)
	if ptr == nil {
		w.str = append(w.str, 0x80)
		return nil
	}
	return writeBigInt(ptr, w)
}

func writeBigIntNoPtr(val reflect.Value, w *encBuf) error {
	i := val.Interface().(big.Int)
	return writeBigInt(&i, w)
}

func writeBigInt(i *big.Int, w *encBuf) error {
	if cmp := i.Cmp(big0); cmp == -1 {
		return errors.New("[writeBigInt] cannot encode negative *big.Int")
	} else if cmp == 0 {
		w.str = append(w.str, 0x80)
	} else {
		w.encodeString(i.Bytes())
	}
	return nil
}

func writeUint(val reflect.Value, w *encBuf) error {
	i := val.Uint()
	if i == 0 {
		w.str = append(w.str, 0x80)
	} else if i < 128 {
		w.str = append(w.str, byte(i))
	} else {
		s := putInt(w.sizeBuf[1:], i)
		w.sizeBuf[0] = 0x80 + byte(s)
		w.str = append(w.str, w.sizeBuf[:s + 1]...)
	}
	return nil
}

func writeBool(val reflect.Value, w *encBuf) error {
	if val.Bool() {
		w.str = append(w.str, 0x01)
	} else {
		w.str = append(w.str, 0x80)
	}
	return nil
}

func writeString(val reflect.Value, w *encBuf) error {
	s := val.String()
	if len(s) == 1 && s[0] <= 0x7f {
		w.str = append(w.str, s[0])
	} else {
		w.encodeStringHeader(len(s))
		w.str = append(w.str, s...)
	}
	return nil
}

func isByte(typ reflect.Type) bool {
	return typ.Kind() == reflect.Uint8 && !typ.Implements(encoderInterface)
}

func writeBytes(val reflect.Value, w *encBuf) error {
	w.encodeString(val.Bytes())
	return nil
}

func writeByteArray(val reflect.Value, w *encBuf) error {
	if !val.CanAddr() {
		c := reflect.New(val.Type()).Elem()
		c.Set(val)
		val = c
	}
	size := val.Len()
	slice := val.Slice(0, size).Bytes()
	w.encodeString(slice)
	return nil
}

func makeSliceWriter(typ reflect.Type, ts tags) (writer, error) {
	etypeInfo, err := cachedTypeInfo1(typ.Elem(), tags{})
	if err != nil {
		return nil, err
	}
	writer := func(val reflect.Value, w *encBuf) error {
		if ts.tail {
			defer w.listEnd(w.list())
		}
		vLen := val.Len()
		for i := 0; i < vLen; i++ {
			if err := etypeInfo.writer(val.Index(i), w); err != nil {
				return err
			}
		}
		return nil
	}
	return writer, nil
}

func makeStructWriter(typ reflect.Type) (writer, error) {
	fields, err := structFields(typ)
	if err != nil {
		return nil, err
	}
	writer := func(val reflect.Value, w *encBuf) error {
		lh := w.list()
		for _, f := range fields {
			if err := f.info.writer(val.Field(f.index), w); err != nil {
				return err
			}
		}
		w.listEnd(lh)
		return nil
	}
	return writer, nil
}

func makePtrWriter(typ reflect.Type) (writer, error) {
	etypeInfo, err := cachedTypeInfo1(typ.Elem(), tags{})
	if err != nil {
		return nil, err
	}

	var nilFunc func(*encBuf) error
	kind := typ.Elem().Kind()
	switch {
	case kind == reflect.Array && isByte(typ.Elem().Elem()):
		nilFunc = func(w *encBuf) error {
			w.str = append(w.str, 0x80)
			return nil
		}
	case kind == reflect.Struct || kind == reflect.Array:
		nilFunc = func(w *encBuf) error {
			w.listEnd(w.list())
			return nil
		}
	default:
		zero := reflect.Zero(typ.Elem())
		nilFunc = func(w *encBuf) error {
			return etypeInfo.writer(zero, w)
		}
	}

	writer := func(val reflect.Value, w *encBuf) error {
		if val.IsNil() {
			return nilFunc(w)
		} else {
			return etypeInfo.writer(val.Elem(), w)
		}
	}

	return writer, err
}

type listHead struct {
	offset int
	size   int
}

func (head *listHead) encode(buf []byte) []byte {
	return buf[:putHead(buf, 0xC0, 0xF7, uint64(head.size))]
}

var encBufPool = sync.Pool{
	New: func() interface{} {
		return &encBuf{sizeBuf: make([]byte, 9)}
	},
}

func headSize(size uint64) int {
	if size < 56 {
		return 1
	}
	return 1 + intSize(size)
}

func putHead(buf []byte, smallTag, largeTag byte, size uint64) int {
	if size < 56 {
		buf[0] = smallTag + byte(size)
		return 1
	} else {
		sizeSize := putInt(buf[1:], size)
		buf[0] = largeTag + byte(sizeSize)
		return sizeSize + 1
	}
}

func putInt(b []byte, i uint64) (size int) {
	switch {
	case i < (1 << 8):
		b[0] = byte(i)
		size = 1
	case i < (1 << 16):
		b[0] = byte(i >> 8)
		b[1] = byte(i)
		size = 2
	case i < (1 << 24):
		b[0] = byte(i >> 16)
		b[1] = byte(i >> 8)
		b[2] = byte(i)
		size = 3
	case i < (1 << 32):
		b[0] = byte(i >> 24)
		b[1] = byte(i >> 16)
		b[2] = byte(i >> 8)
		b[3] = byte(i)
		size = 4
	case i < (1 << 40):
		b[0] = byte(i >> 32)
		b[1] = byte(i >> 24)
		b[2] = byte(i >> 16)
		b[3] = byte(i >> 8)
		b[4] = byte(i)
		size = 5
	case i < (1 << 48):
		b[0] = byte(i >> 40)
		b[1] = byte(i >> 32)
		b[2] = byte(i >> 24)
		b[3] = byte(i >> 16)
		b[4] = byte(i >> 8)
		b[5] = byte(i)
		size = 6
	case i < (1 << 56):
		b[0] = byte(i >> 48)
		b[1] = byte(i >> 40)
		b[2] = byte(i >> 32)
		b[3] = byte(i >> 24)
		b[4] = byte(i >> 16)
		b[5] = byte(i >> 8)
		b[6] = byte(i)
		size = 7
	default:
		b[0] = byte(i >> 56)
		b[1] = byte(i >> 48)
		b[2] = byte(i >> 40)
		b[3] = byte(i >> 32)
		b[4] = byte(i >> 24)
		b[5] = byte(i >> 16)
		b[6] = byte(i >> 8)
		b[7] = byte(i)
		size = 8
	}
	return
}

func intSize(i uint64) (size int) {
	for size = 1; ; size++ {
		if i >>= 8; i == 0 {
			return size
		}
	}
}

func isUint(k reflect.Kind) bool {
	return k >= reflect.Uint && k <= reflect.Uintptr
}