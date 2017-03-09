package test

import (
	"GoOnchain/utility"
)

var usage = `run specific test case`

var flags = []string{}

func main(args []string, p utility.Param) (err error) {

	//TODO
	return nil
}

var Command = &utility.Command{UsageText: usage, Flags: flags, Main: main}
