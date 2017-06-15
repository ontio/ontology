package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const (
	DefaultConfigFilename = "./config.json"
)

var Version string

type Configuration struct {
	Magic            int64    `json:"Magic"`
	Version          int      `json:"Version"`
	SeedList         []string `json:"SeedList"`
	BookKeepers      []string `json:"BookKeepers"` // The default book keepers' publickey
	HttpRestPort     int      `json:"HttpRestPort"`
	HttpJsonPort     int      `json:"HttpJsonPort"`
	HttpLocalPort    int      `json:"HttpLocalPort"`
	OauthServerAddr  string   `json:"OauthServerAddr"`
	NoticeServerAddr string   `json:"NoticeServerAddr"`
	NodePort         int      `json:"NodePort"`
	NodeType         string   `json:"NodeType"`
	WebSocketPort    int      `json:"WebSocketPort"`
	PrintLevel       int      `json:"PrintLevel"`
	IsTLS            bool     `json:"IsTLS"`
	CertPath         string   `json:"CertPath"`
	KeyPath          string   `json:"KeyPath"`
	CAPath           string   `json:"CAPath"`
	GenBlockTime     uint     `json:"GenBlockTime"`
	MultiCoreNum     uint     `json:"MultiCoreNum"`
	EncryptAlg       string   `json:"EncryptAlg"`
	MaxLogSize       int64    `json:"MaxLogSize"`
}

type ConfigFile struct {
	ConfigFile Configuration `json:"Configuration"`
}

var Parameters *Configuration

func init() {
	file, e := ioutil.ReadFile(DefaultConfigFilename)
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
