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
)

const (
	DEFAULT_CONFIG_FILE_NAME = "./config.json"
	MIN_GEN_BLOCK_TIME       = 2
	DEFAULT_GEN_BLOCK_TIME   = 6
	DBFT_MIN_NODE_NUM        = 4 //min node number of dbft consensus
	SOLO_MIN_NODE_NUM        = 1 //min node number of solo consensus
	VBFTMINNODENUM           = 4 //min node number of vbft consensus
)

var Version string

type Configuration struct {
	Magic               int64            `json:"Magic"`
	Version             int              `json:"Version"`
	SeedList            []string         `json:"SeedList"`
	Bookkeepers         []string         `json:"Bookkeepers"` // The default book keepers' publickey
	HttpRestPort        int              `json:"HttpRestPort"`
	RestCertPath        string           `json:"RestCertPath"`
	RestKeyPath         string           `json:"RestKeyPath"`
	HttpInfoPort        uint16           `json:"HttpInfoPort"`
	HttpInfoStart       bool             `json:"HttpInfoStart"`
	HttpWsPort          int              `json:"HttpWsPort"`
	HttpJsonPort        int              `json:"HttpJsonPort"`
	HttpLocalPort       int              `json:"HttpLocalPort"`
	NodePort            int              `json:"NodePort"`
	NodeConsensusPort   int              `json:"NodeConsensusPort"`
	NodeType            string           `json:"NodeType"`
	WebSocketPort       int              `json:"WebSocketPort"`
	PrintLevel          int              `json:"PrintLevel"`
	IsTLS               bool             `json:"IsTLS"`
	CertPath            string           `json:"CertPath"`
	KeyPath             string           `json:"KeyPath"`
	CAPath              string           `json:"CAPath"`
	GenBlockTime        uint             `json:"GenBlockTime"`
	MultiCoreNum        uint             `json:"MultiCoreNum"`
	SignatureScheme     string           `json:"SignatureScheme"`
	MaxLogSize          int64            `json:"MaxLogSize"`
	MaxTxInBlock        int              `json:"MaxTransactionInBlock"`
	MaxHdrSyncReqs      int              `json:"MaxConcurrentSyncHeaderReqs"`
	ConsensusType       string           `json:"ConsensusType"`
	ConsensusConfigPath string           `json:"ConsensusConfigPath"`
	SystemFee           map[string]int64 `json:"SystemFee"`
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

// Polaris test net config
func newPolarisConfig() *Configuration {
	testnet := ` {
  "Configuration": {
    "Magic": 7630401,
    "Version": 23,
    "SeedList": [
      "polaris1.ont.io:20338",
      "polaris2.ont.io:20338",
      "polaris3.ont.io:20338",
      "polaris4.ont.io:20338"
    ],
    "Bookkeepers": [
	  "12020384d843c02ecef233d3dd3bc266ee0d1a67cf2a1666dc1b2fb455223efdee7452",
	  "120203fab19438e18d8a5bebb6cd3ede7650539e024d7cc45c88b95ab13f8266ce9570",
	  "120203c43f136596ee666416fedb90cde1e0aee59a79ec18ab70e82b73dd297767eddf",
	  "120202a76a434b18379e3bda651b7c04e972dadc4760d1156b5c86b3c4d27da48c91a1"
    ],
    "HttpRestPort": 20334,
    "HttpWsPort":20335,
    "HttpJsonPort": 20336,
    "HttpLocalPort": 20337,
    "NodePort": 20338,
    "NodeConsensusPort": 20339,
    "PrintLevel": 1,
    "IsTLS": false,
    "MaxTransactionInBlock": 60000,
    "ConsensusType":"dbft",
    "MultiCoreNum": 4
  }
} `

	config := configFile{}
	e := json.Unmarshal([]byte(testnet), &config)
	if e != nil {
		panic("wrong config file")
	}

	return &config.ConfigFile
}

func init() {
	file, e := ioutil.ReadFile(DEFAULT_CONFIG_FILE_NAME)
	if e != nil {
		log.Printf("[WARN] %v, use default config\n", e)
		Parameters = newPolarisConfig()
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
