package client

import (
	"fmt"
	"time"

	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/eventbus/example/serverclient/message"
)

type Client struct{}

//Call the server synchronous
func (client *Client) SyncCall(serverPID *actor.PID) (interface{}, error) {
	//props := actor.FromProducer(func() actor.Actor { return &Client{} })
	//actor.Spawn(props)
	future := serverPID.RequestFuture(&message.Request{Who: "Ontology"}, 10*time.Second)
	result, err := future.Result()
	return result, err
}

//Call the server asynchronous
func (client *Client) Receive(context actor.Context) {
	switch msg := context.Message().(type) {
	case *actor.Started:
		fmt.Println("Started, initialize client actor here")
	case *actor.Stopping:
		fmt.Println("Stopping, actor is about shut down")
	case *actor.Restarting:
		fmt.Println("Restarting, actor is about restart")
	case *message.Response:
		fmt.Println("Receive message", msg.Welcome)
	}
}

func (client *Client) AsyncCall(serverPID *actor.PID) *actor.PID {
	props := actor.FromProducer(func() actor.Actor { return &Client{} })
	clientPID := actor.Spawn(props)
	serverPID.Request(&message.Request{Who: "Ontology"}, clientPID)
	return clientPID
}

func (client *Client) Stop(pid *actor.PID) {
	pid.Stop()
}
