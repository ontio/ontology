package events

import (
	"testing"
	"fmt"
)

func TestNewEvent(t *testing.T) {
	event := NewEvent()

	var subscriber1 EventFunc = func(v interface{}){
		fmt.Println("subscriber1 event func.")
	}

	var subscriber2 EventFunc = func(v interface{}){
		fmt.Println("subscriber2 event func.")
	}

	fmt.Println("Subscribe...")
	sub1 := event.Subscribe(EventReplyTx,subscriber1)
	event.Subscribe(EventSaveBlock,subscriber2)

	fmt.Println("Notify...")
	event.Notify(EventReplyTx,nil)

	fmt.Println("Notify All...")
	event.NotifyAll()

	event.UnSubscribe(EventReplyTx,sub1)
	fmt.Println("Notify All after unsubscribe sub1...")
	event.NotifyAll()

}