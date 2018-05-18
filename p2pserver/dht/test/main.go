package main

import (
	//"fmt"
	"fmt"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"time"
)

func main() {
	log.Init(log.PATH, log.Stdout)
	_, pub, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID, _ := types.PubkeyID(pub)

	seeds := make([]*types.Node, 0)
	_, pub1, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID1, _ := types.PubkeyID(pub1)

	seed1 := &types.Node{
		ID:      nodeID1,
		IP:      "127.0.0.1",
		UDPPort: 20010,
		TCPPort: 20011,
	}
	seeds = append(seeds, seed1)

	_, pub2, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID2, _ := types.PubkeyID(pub2)
	seed2 := &types.Node{
		ID:      nodeID2,
		IP:      "127.0.0.1",
		UDPPort: 30010,
		TCPPort: 30011,
	}
	seeds = append(seeds, seed2)

	_, pub3, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID3, _ := types.PubkeyID(pub3)
	seed3 := &types.Node{
		ID:      nodeID3,
		IP:      "127.0.0.1",
		UDPPort: 40010,
		TCPPort: 40011,
	}
	seeds = append(seeds, seed3)

	_, pub4, _ := keypair.GenerateKeyPair(keypair.PK_ECDSA, keypair.P256)
	nodeID4, _ := types.PubkeyID(pub4)
	seed4 := &types.Node{
		ID:      nodeID4,
		IP:      "127.0.0.1",
		UDPPort: 50010,
		TCPPort: 50011,
	}
	seeds = append(seeds, seed4)

	testDht := dht.NewDHT(nodeID, seeds)
	testDht.Start()
	timer := time.NewTicker(time.Second)
	for {
		select {
		case <-timer.C:
			fmt.Println("Now table is: ")
			testDht.DisplayRoutingTable()
		}
	}
}
