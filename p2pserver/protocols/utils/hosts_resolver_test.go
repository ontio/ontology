package utils

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

type Foo struct {
	lock  sync.Mutex
	cache unsafe.Pointer
}

func (self *Foo) Get() []string {
	cache := *(*[]string)(self.cache)
	return cache
}

func (self *Foo) Set(str string) {
	cache := []string{"1", str}
	atomic.StorePointer(&self.cache, unsafe.Pointer(&cache))
}

func TestNewHostsResolver(t *testing.T) {

	host, port, err := net.SplitHostPort("127.0.0.1:")

	fmt.Printf("%s, %s, err: %v", host, port, err)

}
