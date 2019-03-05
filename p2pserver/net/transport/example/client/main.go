package main

import (
	"bufio"
	"net"
)

func main() {
	cn, err := net.Dial("tcp", ":1900")
	if err == nil {
		rw := bufio.NewReadWriter(bufio.NewReader(cn), bufio.NewWriter(cn))
		rw.WriteString("Hello World")
	}
}
