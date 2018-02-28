package zmqremote

import (

	"github.com/Ontology/eventbus/actor"
	zmq "github.com/pebbe/zmq4"
	"github.com/Ontology/common/log"
)

var (
	edpReader *endpointReader
	conn      *zmq.Socket
)

func Start(address string) {

	//fmt.Println("address1:" + address)
	actor.ProcessRegistry.RegisterAddressResolver(remoteHandler)
	actor.ProcessRegistry.Address = address

	spawnActivatorActor()
	startEndpointManager()

	edpReader = &endpointReader{}

	conn, _ = zmq.NewSocket(zmq.ROUTER)
	err := conn.Bind("tcp://" + address)
	if err != nil {
		log.Error("Connect bind error.......", err)
	}
	//fmt.Println("after bind " + address)
	go func() {
		edpReader.Receive(conn)
	}()
}

func Shutdonw() {
	edpReader.suspend(true)
	stopEndpointManager()
	stopActivatorActor()
	conn.Close()
}
