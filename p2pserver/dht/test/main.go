package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht"
	"github.com/ontio/ontology/p2pserver/dht/types"
)

// test DHT
func main() {
	log.InitLog(1, log.PATH, log.Stdout)

	seeds := make([]*types.Node, 0, len(config.DefConfig.Genesis.DHT.Seeds))
	for i := 0; i < len(config.DefConfig.Genesis.DHT.Seeds); i++ {
		node := config.DefConfig.Genesis.DHT.Seeds[i]
		pubKey, err := hex.DecodeString(node.PubKey)
		k, err := keypair.DeserializePublicKey(pubKey)
		if err != nil {
			return
		}
		seed := &types.Node{
			IP:      node.IP,
			UDPPort: node.UDPPort,
			TCPPort: node.TCPPort,
		}
		seed.ID, _ = types.PubkeyID(k)
		seeds = append(seeds, seed)
	}

	//start seed node
	//seedIndex, _ := strconv.Atoi(os.Args[1])
	////seedIndex := 1
	//testDht := startSeedNode(seedIndex, seeds)

	var acct *account.Account
	client, err := account.Open("./wallet.dat")
	if client == nil || err != nil {
		log.Fatal("Can't get local account.")
		return
	}
	acct, err = client.GetDefaultAccount([]byte("passwordtest"))
	if acct == nil || err != nil {
		log.Fatal("can not get default account, ", err)
		return
	}

	nodeID, _ := types.PubkeyID(acct.PublicKey)
	testDht := dht.NewDHT(nodeID, seeds)
	testDht.Start()
	log.Info("node ", nodeID, "start")

	stopCh := make(chan int)
	timer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-timer.C:
			fmt.Println("Now table is: ")
			testDht.DisplayRoutingTable()
		}
	}

	<-stopCh
}

func startSeedNode(seedIndex int, seeds []*types.Node) *dht.DHT {
	otherSeeds := make([]*types.Node, 3)
	seedNode := seeds[seedIndex]
	copy(otherSeeds[:seedIndex], seeds[:seedIndex])
	copy(otherSeeds[seedIndex:], seeds[seedIndex+1:])
	seedDht := dht.NewDHT(seedNode.ID, otherSeeds)
	seedDht.SetPort(seedNode.TCPPort, seedNode.UDPPort)
	seedDht.Start()
	log.Info("node ", seedNode.ID, "start")
	return seedDht
}
