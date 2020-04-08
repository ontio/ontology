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

package abi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/ontio/ontology/common/log"
)

var DefAbiMgr = NewAbiMgr()

type AbiMgr struct {
	Path       string
	nativeAbis map[string]*NativeContractAbi
}

func NewAbiMgr() *AbiMgr {
	return &AbiMgr{
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

func (this *AbiMgr) Init(path string) {
	this.Path = path
	this.loadNativeAbi()
}

func (this *AbiMgr) loadNativeAbi() {
	nativeAbiFiles, err := ioutil.ReadDir(this.Path)
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
		data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", this.Path, fileName))
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
		log.Infof("Native contract name:%s address:%s abi load success", fileName, nativeAbi.Address)
	}
}
