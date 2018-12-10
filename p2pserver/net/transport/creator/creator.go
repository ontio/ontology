package creator

import (
	"github.com/ontio/ontology/common/log"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
	"github.com/ontio/ontology/p2pserver/net/transport/quic"
	"github.com/ontio/ontology/p2pserver/net/transport/tcp"
	"sync"
	"errors"
)

var once sync.Once
var instance *transportFactory

type transportFactory struct {
	tspMap map[byte]tsp.Transport
}

func GetTransportFactory() *transportFactory{

	once.Do(func(){
		instance = &transportFactory{}
		instance.tspMap = make(map[byte]tsp.Transport, 2)
		instance.tspMap[tsp.T_TCP], _ = tcp.NewTransport()
		instance.tspMap[tsp.T_QUICK],_ = quic.NewTransport()
	})

	return instance
}

func (this* transportFactory) GetTransport(tspType byte) (tsp.Transport, error) {

	if tsp, ok := instance.tspMap[tspType]; ok {
		return tsp, nil
	}else {
		log.Errorf("[p2p]Can't get the responding Transport, tspType=%d", tspType)
		return nil, errors.New("Can't get the responding Transport")
	}
}
