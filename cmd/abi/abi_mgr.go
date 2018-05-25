package abi

import (
	"encoding/json"
	"fmt"
	"github.com/ontio/ontology/common/log"
	"io/ioutil"
	"strings"
)

const DefAbiPath = "./cmd/abi"

var DefAbiMgr = NewAbiMgr(DefAbiPath)

type AbiMgr struct {
	Path       string
	nativeAbis map[string]*NativeContractAbi
}

func NewAbiMgr(path string) *AbiMgr {
	return &AbiMgr{
		Path:       path,
		nativeAbis: make(map[string]*NativeContractAbi),
	}
}

func (this *AbiMgr) GetNativeAbi(address string) *NativeContractAbi {
	abi, ok := this.nativeAbis[address]
	if ok {
		return abi
	}
	return nil
}

func (this *AbiMgr) Init() {
	this.loadNativeAbi()
}

func (this *AbiMgr) loadNativeAbi() {
	dirName := this.Path + "/native"
	nativeAbiFiles, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Errorf("AbiMgr loadNativeAbi read dir:./native error:%s", err)
		return
	}
	for _, nativeAbiFile := range nativeAbiFiles {
		fileName := nativeAbiFile.Name()
		if nativeAbiFile.IsDir() {
			continue
		}
		if !strings.HasSuffix(fileName, ".json") {
			continue
		}
		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dirName, fileName))
		if err != nil {
			log.Errorf("AbiMgr loadNativeAbi name:%s error:%s", fileName, err)
			continue
		}
		nativeAbi := &NativeContractAbi{}
		err = json.Unmarshal(data, nativeAbi)
		if err != nil {
			log.Errorf("AbiMgr loadNativeAbi name:%s error:%s", fileName, err)
			continue
		}
		this.nativeAbis[nativeAbi.Address] = nativeAbi
	}
}
