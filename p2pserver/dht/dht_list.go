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

package dht

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
)

// read black list from file
func (this *DHT) loadBlackList(fileName string) {
	list := make([]string, 0)
	if common.FileExisted(fileName) {
		buf, err := ioutil.ReadFile(fileName)
		if err != nil {
			log.Error("load black list from %s fail:%s",
				fileName, err.Error())
			return
		}

		err = json.Unmarshal(buf, &list)
		if err != nil {
			log.Error("parse black list file fail: ", err)
			return
		}
	}
	this.blackList = list
}

// loadWhiteList loads whitelist
func (this *DHT) loadWhiteList() {
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers) > 0 {
		this.whiteList = config.DefConfig.P2PNode.ReservedCfg.ReservedPeers
	}
}

// isInWhiteList check whether a given address is in whitelist
func (this *DHT) isInWhiteList(addr string) bool {
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers) > 0 {
		for _, ip := range this.whiteList {
			if strings.HasPrefix(addr, ip) {
				return true
			}
		}
		return false
	}
	return true
}

// isInBlackList check whether a given address is in blacklist
func (this *DHT) isInBlackList(destString string) bool {
	for _, element := range this.blackList {
		if element == destString {
			return true
		}
	}
	return false
}

// saveListToFile save list to a file
func (this *DHT) saveListToFile(list []string, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Errorf("create list file %s failed!", fileName)
		return
	}
	saveContent := ""
	for _, addr := range list {
		saveContent += addr + "\n"
	}
	saveContent = saveContent[:len(saveContent)-1] // remove last \n
	n, err := file.WriteString(saveContent)
	if n != len(saveContent) || err != nil {
		log.Errorf("save list to file %s failed!", fileName)
	}
}
