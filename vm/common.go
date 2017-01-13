package vm

import (
	"math/big"
	"encoding/binary"
	"reflect"
)

type BigIntSorter []big.Int

func (c BigIntSorter) Len() int {
	return len(c)
}
func (c BigIntSorter) Swap(i, j int) {
	if ( i >= 0 && i < len(c) && j >= 0 && j < len(c) ){  // Unit Test modify
		c[i], c[j] = c[j], c[i]
	}
}
func (c BigIntSorter) Less(i, j int) bool {
	if ( i >= 0 && i < len(c) && j >= 0 && j < len(c) ) {   // Unit Test modify
		return c[i].Cmp(&c[j]) < 0
	}

	return false
}


//common func
func SumBigInt (ints []big.Int) big.Int{
	sum := big.NewInt(0)
	for _,v := range ints {
		sum = sum.Add(sum,&v)
	}
	return *sum
}

func MinBigInt(ints []big.Int) big.Int{
	minimum :=  ints[0]

	for _, d := range ints {
		if d.Cmp(&minimum) < 0 {
			minimum = d
		}
	}

	return minimum
}

func MaxBigInt(ints []big.Int) big.Int{
	max :=  ints[0]

	for _, d := range ints {
		if d.Cmp(&max) > 0 {
			max = d
		}
	}

	return max
}

func MinInt64(datas []int64) int64 {

	var minimum int64
	for i, d := range datas {  // Unit Test modify
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
	for i, d := range datas {   // Unit Test modify
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
	for i:=0; i<len; i++ {
		array1 = append(array1, array2[i])   // Unit Test modify
	}

	return array1
}

func BigIntsOp(bigints []big.Int, op OpCode) []big.Int{
	newbigints := []big.Int{}
	for _, b := range bigints {
		var nb *big.Int

		switch op{
		case OP_1ADD:
			nb = b.Add(&b, big.NewInt(int64(1)))
		case OP_1SUB:
			nb = b.Sub(&b, big.NewInt(int64(1)))
		case OP_2MUL:
			nb = b.Lsh(&b, 2)
		case OP_2DIV:
			nb = b.Rsh(&b, 2)
		case OP_NEGATE:
			nb = b.Neg(&b)
		case OP_ABS:
			nb = b.Abs(&b)
		default:
			nb = &b

		}
		newbigints = append(newbigints, *nb)
	}

	return newbigints
}

func BigIntOp (bigints []big.Int, op OpCode) []big.Int{
	newbigints := []big.Int{}
	for _, b := range bigints {
		var nb *big.Int

		switch op{
		case OP_1ADD:
			nb = b.Add(&b, big.NewInt(int64(1)))
		case OP_1SUB:
			nb = b.Sub(&b, big.NewInt(int64(1)))
		case OP_2MUL:
			nb = b.Lsh(&b, 2)
		case OP_2DIV:
			nb = b.Rsh(&b, 2)
		case OP_NEGATE:
			nb = b.Neg(&b)
		case OP_ABS:
			nb = b.Abs(&b)
		default:
			nb = &b

		}
		newbigints = append(newbigints, *nb)
	}

	return newbigints
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
		return 0, ErrBadValue
	}

	var bs [8]byte
	copy(bs[:], b)

	res := binary.LittleEndian.Uint64(bs[:])

	return int64(res), nil
}

func ByteArrZip (s1 [][]byte, s2 [][]byte, op OpCode) [][]byte{
	len := len(s1)
	ns := [][]byte{}
	switch op {
	case OP_CONCAT:
		//for i:=1; i<len; i++ {
		for i:=0; i<len; i++ {              // Unit Test modify
			nsi := Concat(s1[i],s2[i])
			ns = append(ns,nsi)
		}

	}
	return ns
}

func BigIntZip (ints1 []big.Int, ints2 []big.Int, op OpCode) []big.Int{
	newbigints := []big.Int{}

	len := len(ints1)
	for i:=0; i<len; i++ {
		var nb *big.Int

		switch op{
		case OP_AND:
			nb = ints1[i].And(&ints1[i],&ints2[i])
		case OP_OR:
			nb = ints1[i].Or(&ints1[i],&ints2[i])
		case OP_XOR:
			nb = ints1[i].Xor(&ints1[i],&ints2[i])
		case OP_ADD:
			nb = ints1[i].Add(&ints1[i],&ints2[i])
		case OP_SUB:
			nb = ints1[i].Sub(&ints1[i],&ints2[i])
		case OP_MUL:
			nb = ints1[i].Mul(&ints1[i],&ints2[i])
		case OP_DIV:
			nb = ints1[i].Div(&ints1[i],&ints2[i])
		case OP_MOD:
			nb = ints1[i].Mod(&ints1[i],&ints2[i])
		case OP_LSHIFT:
			nb = ints1[i].Lsh(&ints1[i],uint(ints2[i].Int64()))
		case OP_RSHIFT:
			nb = ints1[i].Rsh(&ints1[i],uint(ints2[i].Int64()))
		case OP_MIN:
			c := ints1[i].Cmp(&ints2[i])
			if c <= 0 {
				nb = &ints1[i]
			}else {nb = &ints2[i]}
		case OP_MAX:
			c := ints1[i].Cmp(&ints2[i])
			if c <= 0 {
				nb = &ints2[i]
			}else {nb = &ints1[i]}

		}

		newbigints = append(newbigints, *nb)

	}

	return newbigints
}

func BigIntsComp (bigints []big.Int, op OpCode) []bool{
	compb := []bool{}
	for _, b := range bigints {
		var nb bool

		switch op{
		case OP_0NOTEQUAL:
			nb = b.Cmp(big.NewInt(int64(0))) != 0
		default:
			nb = false
		}
		compb = append(compb, nb)
	}

	return compb
}

func BigIntsMultiComp (ints1 []big.Int, ints2 []big.Int,op OpCode) []bool{

	compb := []bool{}

	len := len(ints1)
	for i:=0; i<len; i++ {
		var nb bool

		switch op{
		case OP_NUMEQUAL:
			nb = ints1[i].Cmp(&ints2[i]) == 0
		case OP_NUMNOTEQUAL:
			nb = ints1[i].Cmp(&ints2[i]) != 0
		case OP_LESSTHAN:
			nb = ints1[i].Cmp(&ints2[i]) < 0
		case OP_GREATERTHAN:
			nb = ints1[i].Cmp(&ints2[i]) > 0
		case OP_LESSTHANOREQUAL:
			nb = ints1[i].Cmp(&ints2[i]) <= 0
		case OP_GREATERTHANOREQUAL:
			nb = ints1[i].Cmp(&ints2[i]) >= 0
		}


		compb = append(compb, nb)
	}

	return compb
}

func BoolsZip (bi1 []bool, bi2 []bool, op OpCode) []bool{
	compb := []bool{}

	len := len(bi1)
	for i:=0; i<len; i++ {
		var nb bool

		switch op{
		case OP_BOOLAND:
			nb = bi1[i] && bi2[i]
		case OP_BOOLOR:
			nb = bi1[i] || bi2[i]
		}

		compb = append(compb, nb)
	}

	return compb
}

func BoolArrayOp (bools []bool, op OpCode) []bool{
	bls := []bool{}
	for _, b := range bools {
		var nb bool

		switch op{
		case OP_NOT:
			nb = !b
		default:
			nb = b
		}
		bls = append(bls, nb)
	}

	return bls
}

func IsEqualBytes(b1 []byte,b2 []byte) bool {
	len1 := len(b1)
	len2 := len(b2)
	if len1 != len2 {return false}

	for i:=0; i<len1; i++ {
		if b1[i] != b2[i] {return false}
	}

	return true
}

func IsEqual(v1 interface{},v2 interface{}) bool {


	if reflect.TypeOf(v1) != reflect.TypeOf(v2) {return false}

	switch t1 := v1.(type){

	case []byte:
		switch t2 := v2.(type) {
		case []byte:
			return 	IsEqualBytes(t1,t2)
		}
	default:
		return false
	}


	return false
}


