package main

import (
	"time"

	"github.com/Ontology/common/log"
	"github.com/Ontology/core"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
	"github.com/Ontology/eventbus/actor"
	"github.com/Ontology/validator/stateless"
	vatypes "github.com/Ontology/validator/types"
	vmtypes "github.com/Ontology/vm/types"
)

type Validator struct {
	Pid       *actor.PID
	CheckType vatypes.VerifyType
}

type TxMsg struct {
	Tx types.Transaction
}

func main() {

	log.Init(log.Stdout)
	log.Log.SetDebugLevel(0)
	// pool logic
	validators := make(map[string]Validator)
	props := actor.FromFunc(func(context actor.Context) {
		switch msg := context.Message().(type) {
		case *vatypes.RegisterValidator:
			log.Infof("validator %v connected", msg.Sender)
			validators[msg.Id] = Validator{Pid: msg.Sender, CheckType: msg.Type}
		case *vatypes.UnRegisterValidator:
			log.Infof("validator %v disconnected", msg.Id)
			if validator, ok := validators[msg.Id]; ok {
				validator.Pid.Tell(&vatypes.UnRegisterAck{Id: msg.Id})
				delete(validators, msg.Id)
			}
		case *vatypes.StatelessCheckResponse:
			log.Info("got message:", msg)
		case *TxMsg:
			log.Info("pool: recevied new tx", msg.Tx)
			// select validator
			for _, v := range validators {
				v.Pid.Request(&vatypes.CheckTx{Tx: msg.Tx}, context.Self())
				break
			}
		}
	})
	pool, _ := actor.SpawnNamed(props, "txpool")

	// validator
	go func() {
		vid := "v1"

		v1, _ := stateless.NewValidator(vid)
		v1.Register(pool)
	}()

	// validator 2
	go func() {
		vid := "v2"
		pool := actor.NewLocalPID("txpool")

		v2, _ := stateless.NewValidator(vid)
		v2.Register(pool)

		v2.UnRegister(pool)

	}()

	// p2p node
	go func() {
		crypto.SetAlg("")
		priv, pub, _ := crypto.GenKeyPair()
		from := core.AddressFromPubKey(&pub)
		tx := NewONTTransferTransaction(from, from)

		sign := SignTransaction(tx, priv)
		tx.Sigs = append(tx.Sigs, &types.Sig{
			PubKeys: []*crypto.PubKey{&pub},
			M:       1,
			SigData: [][]byte{sign},
		})

		pool.Tell(&TxMsg{Tx: *tx})
		pool.Tell(&TxMsg{Tx: *tx})
		pool.Tell(&TxMsg{Tx: *tx})

	}()

	time.Sleep(time.Second * 10)

}

func NewONTTransferTransaction(from, to types.Address) *types.Transaction {
	code := []byte("ont")
	params := append([]byte("transfer"), from[:]...)
	params = append(params, to[:]...)
	vmcode := vmtypes.VmCode{
		CodeType: vmtypes.NativeVM,
		Code:     code,
	}

	tx, _ := core.NewInvokeTransaction(vmcode, params)
	return tx
}

func SignTransaction(tx *types.Transaction, privKey []byte) []byte {
	hash := tx.Hash()
	sign, _ := crypto.Sign(privKey, hash[:])

	return sign

}
