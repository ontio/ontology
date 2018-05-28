package main

import (
	"encoding/hex"
	"fmt"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/dht"
	"github.com/ontio/ontology/p2pserver/dht/types"
	"time"
)

func main() {
	log.Init(log.PATH, log.Stdout)
	var acct *account.Account

	log.Info("0. Open the account")
	client := account.Open("", []byte("passwordtest"))
	if client == nil {
		log.Fatal("Can't get local account.")
		return
	}
	acct = client.GetDefaultAccount()
	if acct == nil {
		log.Fatal("can not get default account")
		return
	}
	log.Debug("The Node's PublicKey ", acct.PublicKey)
	nodeID, _ := types.PubkeyID(acct.PublicKey)

	seeds := make([]*types.Node, 0, len(config.Parameters.DHTSeeds))
	for i := 0; i < len(config.Parameters.DHTSeeds); i++ {
		node := config.Parameters.DHTSeeds[i]
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
