/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

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
