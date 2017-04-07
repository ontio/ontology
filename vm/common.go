package vm

import (
	"DNA/vm/errors"
	"DNA/vm/types"
	"encoding/binary"
	"math/big"
	"reflect"
)

type BigIntSorter []big.Int

func (c BigIntSorter) Len() int {
	return len(c)
}
func (c BigIntSorter) Swap(i, j int) {
	if i >= 0 && i < len(c) && j >= 0 && j < len(c) { // Unit Test modify
		c[i], c[j] = c[j], c[i]
	}
}
func (c BigIntSorter) Less(i, j int) bool {
	if i >= 0 && i < len(c) && j >= 0 && j < len(c) { // Unit Test modify
		return c[i].Cmp(&c[j]) < 0
	}

	return false
}

func ToBigInt(data interface{}) *big.Int {
	var bi big.Int
	switch t := data.(type) {
	case int64:
		bi.SetInt64(int64(t))
	case int32:
		bi.SetInt64(int64(t))
	case int16:
		bi.SetInt64(int64(t))
	case int8:
		bi.SetInt64(int64(t))
	case int:
		bi.SetInt64(int64(t))
	case uint64:
		bi.SetUint64(uint64(t))
	case uint32:
		bi.SetUint64(uint64(t))
	case uint16:
		bi.SetUint64(uint64(t))
	case uint8:
		bi.SetUint64(uint64(t))
	case uint:
		bi.SetUint64(uint64(t))
	case big.Int:
		bi = t
	case *big.Int:
		bi = *t
	}
	return &bi
}

//common func
func SumBigInt(ints []big.Int) big.Int {
	sum := big.NewInt(0)
	for _, v := range ints {
		sum = sum.Add(sum, &v)
	}
	return *sum
}

func MinBigInt(ints []big.Int) big.Int{
	minimum := ints[0]

	for _, d := range ints {
		if d.Cmp(&minimum) < 0 {
			minimum = d
		}
	}

	return minimum
}

func MaxBigInt(ints []big.Int) big.Int{
	max := ints[0]

	for _, d := range ints {
		if d.Cmp(&max) > 0 {
			max = d
		}
	}

	return max
}

func MinInt64(datas []int64) int64 {

	var minimum int64
	for i, d := range datas { // Unit Test modify
		if i == 0 {
			minimum = d
		}
		if d < minimum {
			minimum = d
		}
	}

	return minimum
}

func MaxInt64(datas []int64) int64 {

	var maximum int64
	//i := 0
	for i, d := range datas { // Unit Test modify
		if i == 0 {
			maximum = d
			//i++
		}
		if d > maximum {
			maximum = d
		}
	}

	return maximum
}

func Concat(array1 []byte, array2 []byte) []byte {
	len := len(array2)
	for i := 0; i < len; i++ {
		array1 = append(array1, array2[i]) // Unit Test modify
	}

	return array1
}

func BigIntOp(bi *big.Int, op OpCode) *big.Int {
	var nb *big.Int
	switch op {
	case INC:
		nb = bi.Add(bi, big.NewInt(int64(1)))
	case DEC:
		nb = bi.Sub(bi, big.NewInt(int64(1)))
	case SAL:
		nb = bi.Lsh(bi, 1)
	case SAR:
		nb = bi.Rsh(bi, 1)
	case NEGATE:
		nb = bi.Neg(bi)
	case ABS:
		nb = bi.Abs(bi)
	default:
		nb = bi
	}
	return nb
}

func AsBool(e interface{}) bool {
	if v, ok := e.([]byte); ok {
		for _, b := range v {
			if b != 0 {
				return true
			}
		}
	}
	return false
}

func AsInt64(b []byte) (int64, error) {
	if len(b) == 0 {
		return 0, nil
	}
	if len(b) > 8 {
		return 0, errors.ErrBadValue
	}

	var bs [8]byte
	copy(bs[:], b)

	res := binary.LittleEndian.Uint64(bs[:])

	return int64(res), nil
}

func ByteArrZip(s1 []byte, s2 []byte, op OpCode) []byte{
	var ns []byte
	switch op {
	case CAT:
		ns = append(s1, s2...)
	}
	return ns
}

func BigIntZip(ints1 *big.Int, ints2 *big.Int, op OpCode) *big.Int {
	var nb *big.Int
	switch op {
	case AND:
		nb = ints1.And(ints1, ints2)
	case OR:
		nb = ints1.Or(ints1, ints2)
	case XOR:
		nb = ints1.Xor(ints1, ints2)
	case ADD:
		nb = ints1.Add(ints1, ints2)
	case SUB:
		nb = ints1.Sub(ints1, ints2)
	case MUL:
		nb = ints1.Mul(ints1, ints2)
	case DIV:
		nb = ints1.Div(ints1, ints2)
	case MOD:
		nb = ints1.Mod(ints1, ints2)
	case SHL:
		nb = ints1.Lsh(ints1, uint(ints2.Int64()))
	case SHR:
		nb = ints1.Rsh(ints1, uint(ints2.Int64()))
	case MIN:
		c := ints1.Cmp(ints2)
		if c <= 0 {
			nb = ints1
		} else {
			nb = ints2
		}
	case MAX:
		c := ints1.Cmp(ints2)
		if c <= 0 {
			nb = ints2
		} else {
			nb = ints1
		}
	}
	return nb
}

func BigIntComp(bigint *big.Int, op OpCode) bool {
	var nb bool
	switch op {
	case NZ:
		nb = bigint.Cmp(big.NewInt(int64(0))) != 0
	}
	return nb
}

func BigIntMultiComp(ints1 *big.Int, ints2 *big.Int, op OpCode) bool {
	var nb bool
	switch op {
	case NUMEQUAL:
		nb = ints1.Cmp(ints2) == 0
	case NUMNOTEQUAL:
		nb = ints1.Cmp(ints2) != 0
	case LT:
		nb = ints1.Cmp(ints2) < 0
	case GT:
		nb = ints1.Cmp(ints2) > 0
	case LTE:
		nb = ints1.Cmp(ints2) <= 0
	case GTE:
		nb = ints1.Cmp(ints2) >= 0
	}
	return nb
}

func BoolZip(bi1 bool, bi2 bool, op OpCode) bool {
	var nb bool
	switch op {
	case BOOLAND:
		nb = bi1 && bi2
	case BOOLOR:
		nb = bi1 || bi2
	}
	return nb
}

func BoolArrayOp(bools []bool, op OpCode) []bool {
	bls := []bool{}
	for _, b := range bools {
		var nb bool

		switch op {
		case NOT:
			nb = !b
		default:
			nb = b
		}
		bls = append(bls, nb)
	}

	return bls
}

func IsEqualBytes(b1 []byte, b2 []byte) bool {
	len1 := len(b1)
	len2 := len(b2)
	if len1 != len2 {
		return false
	}

	for i := 0; i < len1; i++ {
		if b1[i] != b2[i] {
			return false
		}
	}

	return true
}

func IsEqual(v1 interface{}, v2 interface{}) bool {

	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {
		return false
	}
	switch t1 := v1.(type) {
	case []byte:
		switch t2 := v2.(type) {
		case []byte:
			return IsEqualBytes(t1, t2)
		}
	case int8, int16, int32, int64:
		if v1 == v2 {
			return true
		}
		return false
	default:
		return false
	}

	return false
}

func WithInOp(int1 *big.Int, int2 *big.Int, int3 *big.Int) bool {
	b1 := BigIntMultiComp(int1, int2, GTE)
	b2 := BigIntMultiComp(int1, int3, LT)
	return BoolZip(b1, b2, BOOLAND)
}

func NewStackItems() []types.StackItem {
	return make([]types.StackItem, 0)
}

func NewStackItem(data interface{}) (types.StackItem, error) {
	var stackItem types.StackItem
	var err error
	switch data.(type) {
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, *big.Int, big.Int:
		stackItem = types.NewInteger(ToBigInt(data))
	case bool:
		stackItem = types.NewBoolean(data.(bool))
	case []byte:
		stackItem = types.NewByteArray(data.([]byte))
	case []types.StackItem:
		stackItem = types.NewArray(data.([]types.StackItem))
	default:
		err = errors.ErrBadType
	}
	return stackItem, err
}

func AssertExecutionContext(context interface{}) *ExecutionContext {
	if c, ok := context.(*ExecutionContext); ok {
		return c
	}
	return nil
}

func AssertStackItem(stackItem interface{}) types.StackItem {
	if s, ok := stackItem.(types.StackItem); ok {
		return s
	}
	return nil
}
