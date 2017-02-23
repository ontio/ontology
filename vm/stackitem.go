package vm

import (
	"math/big"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
)

type StackItem struct {
	array []interface{}
}

func NewStackItem(data interface{}) *StackItem {
	var stackItem StackItem //= append(item.array,data)
	stackItem.array = []interface{} {data}
	return &stackItem
}

func (si *StackItem) Count() int {
	return len(si.array)

}

func (si *StackItem) Hash(op OpCode,vm *ExecutionEngine) [][]byte {
	r := si.GetBytesArray()

	nr := [][]byte{}

	var sh hash.Hash

	switch op {
	case OP_SHA1:
		sh = sha1.New()
	case OP_SHA256:
		sh = sha256.New()
	}

	for _,b := range r{

		var nb []byte
		switch op {
		case OP_SHA1:
		case OP_SHA256:
			sh.Write(b)
			nb = sh.Sum(nil)
		case OP_HASH160:
			nb = vm.crypto.Hash160(b)
		case OP_HASH256:
			nb = vm.crypto.Hash256(b)
		}

		nr = append(nr,nb)
	}
	return nr

}

func (si *StackItem) Distinct() *StackItem {
	length := len(si.array)
	if length <= 1 { return si}

	var list []interface{}

	for i := 0; i < length; i++ {
		found := false
		for j := 0; j < len(list); j++ {
			if IsEqual(si.array[i],list[j]){
				found = true
				break
			}
		}
		if !found { list = append(list,si.array[i])}
	}
	return NewStackItem(list)

}

func (si *StackItem) Concat(Item *StackItem) *StackItem {
	for _,v := range Item.array{
		si.array = append(si.array,v)
	}
	return si
}

func (si *StackItem) Intersect(item *StackItem) *StackItem {
	length := len(si.array)
	length2 := len(item.array)
	if length == 0 { return si}
	if  length2 == 0 {return item}

	var list []interface{}

	for i := 0; i < length; i++ {
		for j := 0; j < length2; j++ {
			if IsEqual(si.array[i],item.array[j]){
				list = append(list,si.array[i])
				break
			}
		}
	}
	return NewStackItem(list)
}

func (si *StackItem) Except(item *StackItem) *StackItem {

	length := len(si.array)
	length2 := len(item.array)
	if length == 0 { return si}
	if  length2 == 0 {return si}

	var list []interface{}

	for i := 0; i < length; i++ {
		found := false
		for j := 0; j < length2; j++ {
			if IsEqual(si.array[i],item.array[j]){
				found = true
				break
			}
		}
		if !found { list = append(list,si.array[i])}
	}
	return NewStackItem(list)

}

func (si *StackItem) Take(count int) *StackItem {

	// 参数小于零错误
	if ( count < 0 ) {count = 0}

	if ( count >= si.Count() ) {return si}
	//si.array = si.array[0:count-1]
	si.array = si.array[0:count]      // Unit Test modify
	return  si

}

func (si *StackItem) Skip(count int) *StackItem {
	if ( count >= 0 && count <= si.Count() ) {
		//si.array = si.array[count-1 : ]
		si.array = si.array[count: ]     // Unit Test modify
		return si
	}

	// 错误的参数，返回原数组
	return si;
}

func (si *StackItem) ElementAt(index int) *StackItem {

	if index == 0 && si.Count() == 1 {return si}

	narr := []interface{} {si.array[index]}
	return NewStackItem(narr)

}

func (si *StackItem) Reverse() StackItem {

    var nsi []interface{}
	len := si.Count()

	for i := len-1; i >= 0; i-- {
		nsi = append(nsi,si.array[i])
	}

	return *NewStackItem(nsi)
}

func (si *StackItem) GetArray() []StackItem {
	arr := []StackItem{}

	for _, b := range si.array {
		si := NewStackItem(b)
		arr = append(arr,*si)
	}
	return arr

}

func (si *StackItem) GetBytes() []byte {

	if v, ok := si.array[0].([]byte); ok {
		return v
	}

	return nil

}

func (si *StackItem) GetBytesArray() [][]byte {
	bys  := make([][]byte,0)
	for _, b := range si.array {
		switch t := b.(type){
		case []byte:
			bys = append(bys,t)
		}
	}
	return bys
}

func (si *StackItem) GetIntArray() []big.Int {
	bigInts  := make([]big.Int,0)
	for _, b := range si.array {
		bi := big.Int{}
		switch t := b.(type){
		case []byte:
			bi.SetBytes(t)
			bigInts = append(bigInts,bi)
		}



	}
	return bigInts
}

func (si *StackItem) GetBoolArray() []bool {

	bools := []bool{}
	for _, b := range si.array {
		bv := AsBool(b)
		bools = append(bools, bv)
	}
	return bools
}

func (si *StackItem) GetBool() bool {
	flag := si.array[0].(bool)
	return flag
}

func (si *StackItem) ToBool() bool {
	return AsBool(si.array[0])
}

func (si *StackItem) ToBigInt() *big.Int {
	var bi big.Int

	switch t := si.array[0].(type){
	case []byte:
		bi.SetBytes(t)
	}

	return &bi
}

