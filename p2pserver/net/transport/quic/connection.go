package quic

import (
	"io"
	"net"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/ontio/ontology/common/log"
)

type connection struct {
	sess          quic.Session
	streamWTimeOut time.Time
}

func (this * connection) GetReader() (io.Reader, error) {

	stream, err := this.sess.AcceptUniStream()
	if err != nil {
		log.Errorf("[p2p]AcceptStream lAddr=%s, rAddr=%s, ERR:%s", this.sess.LocalAddr().String(), this.sess.RemoteAddr().String(), err)
		return  nil, err
	}

	return stream, nil
}

func (this * connection) Write(b []byte) (int, error) {

	stream, err := this.sess.OpenUniStreamSync()
	if err != nil {
		log.Errorf("[p2p]OpenStreamSync lAddr=%s, rAddr=%s, ERR:%s", this.sess.LocalAddr().String(), this.sess.RemoteAddr().String(), err)
		return 0, err
	}
	defer stream.Close()

	stream.SetWriteDeadline(this.streamWTimeOut)
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
