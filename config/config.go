package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const (
	defaultConfigFilename = "./config.json"
)

type ProtocolConfiguration struct {
	Magic         int64    `json:"Magic"`
	CoinVersion   int      `json:"CoinVersion"`
	StandbyMiners []string `json:"StandbyMiners"`
	SeedList      []string `json:"SeedList"`
}

type ProtocolFile struct {
	ProtocolConfig ProtocolConfiguration `json:"ProtocolConfiguration"`
}

func SeedNodes() ([]string, error) {
	file, e := ioutil.ReadFile("./config/protocol.json")
	if e != nil {
		log.Fatalf("File error: %v\n", e)
		os.Exit(1)
	}
	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	config := ProtocolFile{}
	e = json.Unmarshal(file, &config)
	if e != nil {
		log.Fatalf("Unmarshal json file erro %v", e)
	}
	log.Printf("Protocol configuration: %v\n", config)
	return config.ProtocolConfig.SeedList, e
}

// filesExists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
