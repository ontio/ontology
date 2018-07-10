package dht

import (
	"encoding/json"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	"io/ioutil"
	"os"
	"strings"
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

func (this *DHT) loadWhiteList() {
	if config.DefConfig.P2PNode.ReservedPeersOnly && len(config.DefConfig.P2PNode.ReservedCfg.ReservedPeers) > 0 {
		this.whiteList = config.DefConfig.P2PNode.ReservedCfg.ReservedPeers
	}
}

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

func (this *DHT) isInBlackList(destString string) bool {
	for _, element := range this.blackList {
		if element == destString {
			return true
		}
	}
	return false
}

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
