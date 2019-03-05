package main

import (
	"bufio"
	"fmt"
	smux "github.com/libp2p/go-stream-muxer"
	"github.com/whyrusleeping/go-smux-yamux"
	"net"
	"os"
)



func handleStream(s smux.Stream) {
	fmt.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	//go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}
func readData(rw *bufio.ReadWriter) {
	for {
		str, _ := rw.ReadString('\n')

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}

	}
}

func writeData(rw *bufio.ReadWriter) {
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		rw.WriteString(fmt.Sprintf("%s\n", sendData))
		rw.Flush()
	}

}

func main() {

	ls, err := net.Listen("tcp", "127.0.0.1:1900")
	if err == nil {
		for {
			cn, err := ls.Accept()
			if err == nil {
				smConn, _ := sm_yamux.DefaultTransport.NewConn(cn, true)
				go func(sMuxConn smux.Conn){
					for {
						recvStream, err := sMuxConn.AcceptStream()
						if err == nil {
							handleStream(recvStream)
						}
					}
				}(smConn)
			}
		}
	}
}
