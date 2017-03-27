package main

import (
	"DNA/common/log"
	"DNA/crypto"
	"DNA/utility"
	"DNA/utility/consensus"
	"DNA/utility/info"
	"DNA/utility/test"
	"os"
)

const (
	path string = "./Log"
)

func main() {
	crypto.SetAlg(crypto.P256R1)
	log.CreatePrintLog(path)

	cmds := map[string]*utility.Command{
		"info":      info.Command,
		"consensus": consensus.Command,
		"test":      test.Command,
	}

	err := utility.Start(cmds)
	if err != nil {
		os.Exit(1)
	}
}
