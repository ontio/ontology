package vbft

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Ontology/simulations"
)

func generTestData() []byte {
	nodeId, _ := simulations.StringID("206520e7475798520164487f7e4586bb55790097ceb786aab6d5bc889d12991a5a204c6298bef1bf43c20680a3979a213392b99c97042ebae27d2a7af6442aa7c008")
	chainPeers := make([]*PeerConfig, 0)
	peerconfig := &PeerConfig{
		Index: 12,
		ID:    nodeId,
	}
	chainPeers = append(chainPeers, peerconfig)

	tests := &ChainConfig{
		Version:       1,
		View:          12,
		N:             4,
		F:             3,
		BlockMsgDelay: 1000,
		HashMsgDelay:  1000,
		SyncDelay:     1000,
		Peers:         chainPeers,
		DposTable:     []uint32{2, 3, 1, 3, 1, 3, 2, 3, 2, 3, 2, 1, 3},
	}
	cc := new(bytes.Buffer)
	tests.Serialize(cc)
	return cc.Bytes()
}
func TestSerialize(t *testing.T) {
	res := generTestData()
	fmt.Println("serialize:", res)
}

func TestDeserialize(t *testing.T) {
	res := generTestData()
	test := &ChainConfig{}
	err := test.Deserialize(bytes.NewReader(res))

	if err != nil {
		t.Log("test failed ")
	}
	fmt.Printf("version: %d, F:%d, SyncDelay:%d \n", test.Version, test.F, test.SyncDelay)
	fmt.Println("peers:      ", test.Peers[0].ID.String())
	fmt.Println("dpostable:", test.DposTable[0])
}
