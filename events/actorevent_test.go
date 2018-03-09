package events

import (
	"github.com/Ontology/eventbus/actor"
	"fmt"
	"testing"
	"time"
)

const testTopic = "test"
type testMessage struct {
	Message string
}

func testSubReceive(c actor.Context) {
	switch msg := c.Message().(type) {
	case *testMessage:
		fmt.Printf("PID:%s receive message:%s\n", c.Self().Id, msg.Message)
	}
}

func TestActorEvent(t *testing.T){
	Init()
	subPID1 := actor.Spawn( actor.FromFunc(testSubReceive))
	subPID2 := actor.Spawn(actor.FromFunc(testSubReceive))
	sub1 := NewActorSubscriber(subPID1)
	sub2 := NewActorSubscriber(subPID2)
	sub1.Subscribe(testTopic)
	sub2.Subscribe(testTopic)
	DefActorPublisher.Publish(testTopic, &testMessage{Message:"Hello"})
	time.Sleep(time.Millisecond)
	DefActorPublisher.Publish(testTopic, &testMessage{Message:"Word"})
}

