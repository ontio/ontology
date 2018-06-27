package msgcount

import (
	"os"

	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/common"
)

var file *os.File

//SaveMsgCountLog save message count log to file
func SaveMsgCountLog(l string) error {
	var err error
	if file == nil || checkIfNeedNewFile() {
		file, err = log.FileOpen(common.MSG_COUNT_LOG_DIR)
		if err != nil {
			return err
		}
	}
	_, err = file.WriteString(l)
	return err
}

//checkIfNeedNewFile check if need a new log file
func checkIfNeedNewFile() bool {
	f, e := file.Stat()
	if e != nil {
		return false
	}
	return f.Size() > common.MAX_MSG_COUNT_LOG_SIZE
}
