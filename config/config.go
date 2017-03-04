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
	HttpJsonPort  int	`json:"HttpJsonPort"`
	NodePort      int	`json:"NodePort"`
	WebSocketPort int	`json:"WebSocketPort"`
	MinerName     string	`json:"MinerName"`
}

type ProtocolFile struct {
	ProtocolConfig ProtocolConfiguration `json:"ProtocolConfiguration"`
}

var Parameters *ProtocolConfiguration

func init() {
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
		os.Exit(1)
	}
	Parameters = &(config.ProtocolConfig)
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
