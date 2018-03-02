package statefull

import (
	"bytes"
	"testing"
	"time"

	"github.com/Ontology/common/log"
	"github.com/Ontology/core"
	"github.com/Ontology/core/genesis"
	tx "github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/actor"
	tc "github.com/Ontology/txnpool/common"
	"github.com/Ontology/validator/db"
	"github.com/stretchr/testify/assert"
)

func init() {
	crypto.SetAlg("")
	log.Init(log.Path, log.Stdout)
}

func TestVerifier(t *testing.T) {
	store, err := db.NewStore("temp.db")
	if assert.Nil(t, err) == false {
		return
	}

	verifier := NewDBVerifier(tc.StatefulV, store)

	props := actor.FromProducer(func() actor.Actor {
		return verifier
	})
	pid := actor.Spawn(props)
	verifier.SetPID(pid)

	_, issuer, _ := crypto.GenKeyPair()
	txn, _ := core.NewBookKeeperTransaction(&issuer, true, []byte{}, &issuer)

	block, _ := genesis.GenesisBlockInit([]*crypto.PubKey{&issuer})
	block.Transactions = append(block.Transactions, txn)
	block.RebuildMerkleRoot()

	buf := bytes.NewBuffer(nil)
	block.Serialize(buf)
	block.Deserialize(buf)

	pid.Tell(block)

	time.Sleep(time.Second * 1)

	req := &tc.VerifyReq{
		WorkerId: 1,
		Txn:      &tx.Transaction{},
	}

	for i := 0; i < 100; i++ {
		pid.Tell(req)
	}

	time.Sleep(time.Second * 2)
	verifier.Stop()
}
