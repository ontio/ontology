package main

import (
	"GoOnchain/common/log"
	"GoOnchain/crypto"
	"GoOnchain/utility"
	"GoOnchain/utility/consensus"
	"GoOnchain/utility/info"
	"GoOnchain/utility/test"
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
