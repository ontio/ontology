package cli

import (
	"math/rand"
	"time"

	"DNA/common/config"
	"DNA/common/log"
	"DNA/crypto"
)

func init() {
	var path string = "./Log/"
	log.CreatePrintLog(path)
	crypto.SetAlg(config.Parameters.EncryptAlg)
	//seed transaction nonce
	rand.Seed(time.Now().UnixNano())
}
