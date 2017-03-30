package main

import (
	"os"

	"github.com/DNAProject/DNA/common/log"
	"github.com/DNAProject/DNA/crypto"
	"github.com/DNAProject/DNA/utility"
	"github.com/DNAProject/DNA/utility/consensus"
	"github.com/DNAProject/DNA/utility/info"
	"github.com/DNAProject/DNA/utility/test"
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
