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
	tspMap map[string]tsp.Transport
}

func GetTransportFactory() *transportFactory{

	once.Do(func(){
		instance = &transportFactory{}
		instance.tspMap = make(map[string]tsp.Transport, 2)
		instance.tspMap["TCP"], _ = tcp.NewTransport()
		instance.tspMap["QUIC"],_ = quic.NewTransport()
	})

	return instance
}

func (this* transportFactory) GetTransport(tspName string) (tsp.Transport, error) {

	if tsp, ok := instance.tspMap[tspName]; ok {
		return tsp, nil
	}else {
		log.Errorf("[p2p]Can't get the responding Transport, tspName=%s", tspName)
		return nil, errors.New("Can't get the responding Transport")
	}
}
