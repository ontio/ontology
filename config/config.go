package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const (
	DefaultConfigFilename = "./config.json"
)

type ProtocolConfiguration struct {
	Magic           int64    `json:"Magic"`
	CoinVersion     int      `json:"CoinVersion"`
	SeedList        []string `json:"SeedList"`
	HttpJsonPort    int      `json:"HttpJsonPort"`
	HttpLocalPort   int      `json:"HttpLocalPort"`
	NodePort        int      `json:"NodePort"`
	WebSocketPort   int      `json:"WebSocketPort"`
	BookKeeperName  string   `json:"BookKeeperName"`
	PrintLevel      int      `json:"PrintLevel"`
	IsTLS           bool     `json:"IsTLS"`
	CertPath        string   `json:"CertPath"`
	KeyPath         string   `json:"KeyPath"`
	CAPath          string   `json:"CAPath"`
	GenBlockTime    uint     `json:"GenBlockTime"`
	BookKeeperCount uint32   `json:"BookKeeperCount"`
}

type ProtocolFile struct {
	ProtocolConfig ProtocolConfiguration `json:"ProtocolConfiguration"`
}

type hashStruct struct {
	PublicKeyHash string
}

var Parameters *ProtocolConfiguration

func ReadNodeID() uint64 {
	clientName := Parameters.BookKeeperName
	var n uint32
	fmt.Sscanf(clientName, "c%d", &n)
	w := fmt.Sprintf("./wallet%d.txt", n)
	file, e := ioutil.ReadFile(w)
	if e != nil {
		log.Fatalf("File error: %v\n", e)
		os.Exit(1)
	}
	// Remove the UTF-8 Byte Order Mark
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))

	var hash hashStruct
	e = json.Unmarshal(file, &hash)
	s := hash.PublicKeyHash[:16]
	id, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		fmt.Println(err)
	}
	return id
}

func init() {
	file, e := ioutil.ReadFile(DefaultConfigFilename)
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
