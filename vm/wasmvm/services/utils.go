package services
/*

import (
	"errors"
	"github.com/Ontology/vm/wasmvm/memory"
	"fmt"
)

//TODO deside to replace the PUNKNOW type

//for the c language "calloc" function
func calloc(memo *memory.VMmemory, params []uint64) (uint64, error) {

	if len(params) != 2 {
		return uint64(0), errors.New("parameter count error while call calloc")
	}
	count := int(params[0])
	length := int(params[1])
	//we don't know whats the alloc type here
	index, err := memo.MallocPointer(count*length, memory.P_UNKNOW)
	if err != nil {
		return uint64(0), err
	}
	return uint64(index), nil
}

//for the c language "malloc" function
func malloc(memo *memory.VMmemory, params []uint64) (uint64, error) {

	if len(params) != 1 {
		return uint64(0), errors.New("parameter count error while call calloc")
	}
	size := int(params[0])
	//we don't know whats the alloc type here
	index, err := memo.MallocPointer(size, memory.P_UNKNOW)
	if err != nil {
		return uint64(0), err
	}
	return uint64(index), nil
}

//TODO use arrayLen to replace 'sizeof'
func arrayLen(memo *memory.VMmemory, params []uint64) (uint64, error) {
	if len(params) != 1 {
		return uint64(0), errors.New("parameter count error while call arrayLen")
	}

	pointer := params[0]
	fmt.Printf("pointer is %v\n",pointer)


	tl,ok := memo.MemPoints[pointer]
	if ok{
		switch tl.Ptype {
		case memory.P_INT8, memory.P_STRING:
			return uint64(tl.Length / 1), nil
		case memory.P_INT16:
			return uint64(tl.Length / 2), nil
		case memory.P_INT32, memory.P_FLOAT32:
			return uint64(tl.Length / 4), nil
		case memory.P_INT64, memory.P_FLOAT64:
			return uint64(tl.Length / 8), nil
		case memory.P_UNKNOW:
			//todo ???
			return uint64(0), nil
		default:
			return uint64(0), nil
		}
	}else {
		return uint64(0), nil
	}


}

func memcpy(memo *memory.VMmemory, params []uint64) (uint64, error) {
	if len(params) != 3 {
		return uint64(0), errors.New("parameter count error while call memcpy")
	}
	dest := int(params[0])
	src := int(params[1])
	length := int(params[2])

	if dest < src && dest+length > src {
		return uint64(0), errors.New("memcpy overlapped")
	}

	copy(memo.Memory[dest:dest+length], memo.Memory[src:src+length])
	return uint64(1),nil  //this return will be dropped in wasm
}

func readMessage(memo *memory.VMmemory,params []uint64)(uint64,error){

	if len(params) != 2 {
		return uint64(0), errors.New("parameter count error while call readMessage")
	}

	addr := int(params[0])
	length := int(params[1])

	msgBytes ,err:= memo.GetMessageBytes()
	if err != nil{
		return uint64(0),err
	}
	if length != len(msgBytes) {
		return uint64(0),errors.New("readMessage length error")
	}
	copy(memo.Memory[addr:addr+length],msgBytes[:length])
	return uint64(length),nil
}*/
