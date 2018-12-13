package creator

import (
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
	tsp "github.com/ontio/ontology/p2pserver/net/transport"
	"github.com/ontio/ontology/p2pserver/net/transport/quic"
	"github.com/ontio/ontology/p2pserver/net/transport/tcp"

	"errors"
	"sync"
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
		instance.tspMap[common.T_TCP], _ = tcp.NewTransport()
		instance.tspMap[common.T_QUIC],_ = quic.NewTransport()
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
