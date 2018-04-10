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
)

var Version string

type Configuration struct {
	Magic             int64            `json:"Magic"`
	Version           int              `json:"Version"`
	SeedList          []string         `json:"SeedList"`
	Bookkeepers       []string         `json:"Bookkeepers"` // The default book keepers' publickey
	HttpRestPort      int              `json:"HttpRestPort"`
	RestCertPath      string           `json:"RestCertPath"`
	RestKeyPath       string           `json:"RestKeyPath"`
	HttpInfoPort      uint16           `json:"HttpInfoPort"`
	HttpInfoStart     bool             `json:"HttpInfoStart"`
	HttpWsPort        int              `json:"HttpWsPort"`
	HttpJsonPort      int              `json:"HttpJsonPort"`
	HttpLocalPort     int              `json:"HttpLocalPort"`
	NodePort          uint16           `json:"NodePort"`
	NodeConsensusPort uint16           `json:"NodeConsensusPort"`
	NodeType          string           `json:"NodeType"`
	WebSocketPort     int              `json:"WebSocketPort"`
	PrintLevel        int              `json:"PrintLevel"`
	IsTLS             bool             `json:"IsTLS"`
	CertPath          string           `json:"CertPath"`
	KeyPath           string           `json:"KeyPath"`
	CAPath            string           `json:"CAPath"`
	GenBlockTime      uint             `json:"GenBlockTime"`
	MultiCoreNum      uint             `json:"MultiCoreNum"`
	SignatureScheme   string           `json:"SignatureScheme"`
	MaxLogSize        int64            `json:"MaxLogSize"`
	MaxTxInBlock      int              `json:"MaxTransactionInBlock"`
	MaxHdrSyncReqs    int              `json:"MaxConcurrentSyncHeaderReqs"`
	ConsensusType     string           `json:"ConsensusType"`
	SystemFee         map[string]int64 `json:"SystemFee"`
}

type ConfigFile struct {
	ConfigFile Configuration `json:"Configuration"`
}

var Parameters *Configuration

func init() {
	file, e := ioutil.ReadFile(DEFAULT_CONFIG_FILE_NAME)
	if e != nil {
		log.Fatalf("File error: %v\n", e)
		os.Exit(1)
	}
	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	config := ConfigFile{}
	e = json.Unmarshal(file, &config)
	if e != nil {
		log.Fatalf("Unmarshal json file erro %v", e)
		os.Exit(1)
	}
	Parameters = &(config.ConfigFile)
}
