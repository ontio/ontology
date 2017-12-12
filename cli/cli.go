package cli

import (
	"math/rand"
	"time"

	"github.com/Ontology/common/config"
	"github.com/Ontology/common/log"
	"github.com/Ontology/crypto"
)

func init() {
	log.Init()
	crypto.SetAlg(config.Parameters.EncryptAlg)
	//seed transaction nonce
	rand.Seed(time.Now().UnixNano())
}
