package vbft

import (
	"fmt"
	"testing"
)

func Test_calcParticipantPeers(t *testing.T) {
	dposTable := []uint32{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	chain := &ChainConfig{
		F:         2,
		DposTable: dposTable,
	}

	vrf := VRFValue{}
	for i := range vrf.Bytes() {
		vrf[i] = byte(i + 1)
	}
	participantCfg := &BlockParticipantConfig{
		Vrf:         vrf,
		ChainConfig: chain,
	}

	id := calcParticipant(vrf, dposTable, 0)
	if id != dposTable[1] {
		t.FailNow()
	}
	fmt.Printf("id: %d \n", id)

	ids := calcParticipantPeers(participantCfg, chain, 0, 2)
	if len(ids) != 2 || ids[0] != id || ids[1] != 1 {
		t.FailNow()
	}
	fmt.Printf("ids: %v \n", ids)

	ids2 := calcParticipantPeers(participantCfg, chain, 0, 100)
	if len(ids2) != 4 {
		// all peers should be selected
		t.FailNow()
	}
	fmt.Printf("ids: %v \n", ids2)
}
