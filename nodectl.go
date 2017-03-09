package main

import (
	"GoOnchain/utility"
	"GoOnchain/utility/consensus"
	"GoOnchain/utility/info"
	"GoOnchain/utility/test"
	"os"
)

func main() {
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
