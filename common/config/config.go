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

package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/ontio/ontology/cmd/utils"
	"github.com/urfave/cli"
)

const (
	DEFAULT_CONFIG_FILE_NAME = "./config.json"
	MIN_GEN_BLOCK_TIME       = 2
	DEFAULT_GEN_BLOCK_TIME   = 6
	DBFT_MIN_NODE_NUM        = 4 //min node number of dbft consensus
	SOLO_MIN_NODE_NUM        = 1 //min node number of solo consensus
)

var Version string

type Configuration struct {
	Magic             int64            `json:"Magic"`
	Version           int              `json:"Version"`
	SeedList          []string         `json:"SeedList"`
	Bookkeepers       []string         `json:"Bookkeepers"` // The default book keepers' publickey
	HttpRestPort      int              `json:"HttpRestPort"`
	HttpCertPath      string           `json:"HttpCertPath"`
	HttpKeyPath       string           `json:"HttpKeyPath"`
	HttpInfoPort      uint16           `json:"HttpInfoPort"`
	HttpWsPort        int              `json:"HttpWsPort"`
	HttpJsonPort      int              `json:"HttpJsonPort"`
	HttpLocalPort     int              `json:"HttpLocalPort"`
	NodePort          int              `json:"NodePort"`
	NodeConsensusPort int              `json:"NodeConsensusPort"`
	NodeType          string           `json:"NodeType"`
	PrintLevel        int              `json:"PrintLevel"`
	IsTLS             bool             `json:"IsTLS"`
	CertPath          string           `json:"CertPath"`
	KeyPath           string           `json:"KeyPath"`
	CAPath            string           `json:"CAPath"`
	GenBlockTime      uint             `json:"GenBlockTime"`
	MultiCoreNum      uint             `json:"MultiCoreNum"`
	MaxLogSize        int64            `json:"MaxLogSize"`
	MaxTxInBlock      int              `json:"MaxTransactionInBlock"`
	MaxHdrSyncReqs    int              `json:"MaxConcurrentSyncHeaderReqs"`
	ConsensusType     string           `json:"ConsensusType"`
	SystemFee         map[string]int64 `json:"SystemFee"`
}

type configFile struct {
	ConfigFile Configuration `json:"Configuration"`
}

func newDefaultConfig() *Configuration {
	return &Configuration{
		Magic:             12345,
		Version:           0,
		HttpRestPort:      20334,
		HttpWsPort:        20335,
		HttpJsonPort:      20336,
		HttpLocalPort:     20337,
		NodePort:          20338,
		NodeConsensusPort: 20339,
		PrintLevel:        1,
		GenBlockTime:      6,
		MultiCoreNum:      4,
		MaxTxInBlock:      5000,
		ConsensusType:     "solo",
		SystemFee:         make(map[string]int64),
	}
}

var Parameters *Configuration

func Init(ctx *cli.Context) {
	var file []byte
	var e error
	configDir := ctx.GlobalString(utils.ConfigUsedFlag.Name)
	if "" == configDir {
		file, e = ioutil.ReadFile(DEFAULT_CONFIG_FILE_NAME)
	} else {
		file, e = ioutil.ReadFile(configDir)
	}
	if e != nil {
		log.Printf("[ERROR] %v, use default config\n", DEFAULT_CONFIG_FILE_NAME, e)
		Parameters = newDefaultConfig()
		return
	}

	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	config := configFile{}
	e = json.Unmarshal(file, &config)
	if e != nil {
		log.Fatalf("Unmarshal json file erro %v", e)
		os.Exit(1)
	}
	Parameters = &(config.ConfigFile)
}
