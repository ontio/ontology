package main

import (
	"fmt"
	"time"

	"github.com/Ontology/eventbus/example/serverclient/client"
	"github.com/Ontology/eventbus/example/serverclient/server"
)

func main() {
	server := &server.Server{}
	client := &client.Client{}
	serverPID := server.Start()
	result, err := client.SyncCall(serverPID)
	if err != nil {
		fmt.Println("ERROR:", err)
	}
	fmt.Println(result)
	fmt.Println("###################################")

	clientPID := client.AsyncCall(serverPID)

	time.Sleep(1 * time.Second)
	server.Stop(serverPID)
	client.Stop(clientPID)
}
