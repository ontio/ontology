package main

import (
	"fmt"
	"reflect"
	"sync"
	"time"
)

type buffer []byte

type pp struct {
	buf        buffer
	arg        interface{}
	value      reflect.Value
	reordered  bool
	goodArgNum bool
	panicking  bool
	erroring   bool
}

var bytePool = sync.Pool{
	New: func() interface{} {
		return new(pp)
	},
}

func main() {
	a := time.Now().Unix()
	for i := 0; i < 100000000; i++ {
		obj := new(pp)
		_ = obj
	}

	b := time.Now().Unix()
	for i := 0; i < 100000000; i++ {
		obj := bytePool.Get().(*pp)
		_ = obj
		bytePool.Put(obj)
	}

	c := time.Now().Unix()
	fmt.Println("without pool ", b-a, "s")
	fmt.Println("with    pool ", c-b, "s")

}
