package protocol

import (
	"fmt"
	"encoding/hex"
)

// TODO Sample function, shoule be called from ledger module
func LedgerGetHeader() ([]byte, error) {
	genesisHeader := "b3181718ef6167105b70920e4a8fbbd0a0a56aacf460d70e10ba6fa1668f1fef"

	h, err := hex.DecodeString(genesisHeader)
	if err != nil {
		fmt.Printf("Decode Header hash error")
		return nil, err
	}
	return h, nil
}

func LedgerInit() {

}
