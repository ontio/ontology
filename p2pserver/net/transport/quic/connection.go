package quic

import (
	"github.com/ontio/ontology/p2pserver/common"
	"io"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ontio/ontology/common/log"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
)

type recvStream struct {
	io.Reader
}

type connection struct {
	sess           quic.Session
	sstreamMap     map[string]quic.SendStream
	streamWTimeOut time.Time
}

func (this * recvStream) CanContinue() bool {

	return true
}

func newConnection(sess quic.Session) tsp.Connection  {

	return &connection{
		sess:       sess,
		sstreamMap: make(map[string]quic.SendStream),
	}
}

func (this * connection) GetRecvStream() (tsp.RecvStream, error) {

	stream, err := this.sess.AcceptUniStream()
	if err != nil{
		log.Errorf("[p2p]AcceptUniStream lAddr=%s, rAddr=%s, ERR:%s", this.sess.LocalAddr().String(), this.sess.RemoteAddr().String(), err)
		return nil, err
	}

	return &recvStream{stream}, nil
}

func (this * connection) GetTransportType() byte {

	return 	common.T_QUIC
}

func (this * connection) Write(cmdType string, b []byte) (int, error) {

	var stream quic.SendStream
	if s, ok := this.sstreamMap[cmdType]; !ok {
		s, err := this.sess.OpenUniStreamSync()
		if err != nil {
			log.Errorf("[p2p]OpenUniStreamSync lAddr=%s, rAddr=%s, ERR:%s", this.sess.LocalAddr().String(), this.sess.RemoteAddr().String(), err)
			return 0, err
		}
		//s.SetWriteDeadline(this.streamWTimeOut)

		this.sstreamMap[cmdType] = s
		stream = s
	}else {
		stream = s
	}

	cntW, errW := stream.Write(b)
	if errW != nil {
		log.Errorf("[p2p] Write err by stream:%s", errW)
		return 0, errW
	}

	return cntW, errW
}

func (this * connection) Close() error {

	return this.sess.Close()
}

func (this* connection) LocalAddr() net.Addr {

	return this.sess.LocalAddr()
}

func (this* connection) RemoteAddr() net.Addr {

	return this.sess.RemoteAddr()
}

func (this * connection) SetWriteDeadline(t time.Time) error {

	this.streamWTimeOut = t

	return nil
}
