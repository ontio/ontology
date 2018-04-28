package vbft

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/json"
	"fmt"

	. "github.com/Ontology/common"
	"github.com/Ontology/crypto"
)

func SignMsg(sk []byte, msg ConsensusMsg) ([]byte, error) {

	data, err := msg.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal msg when signing: %s", err)
	}

	return crypto.Sign(sk, data)
}

func HashBlock(blk *Block) (Uint256, error) {

	// FIXME: has to do marshal on each call

	data, err := json.Marshal(blk)
	if err != nil {
		return Uint256{}, fmt.Errorf("failed to marshal block: %s", err)
	}

	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return Uint256(f), nil
}

func HashMsg(msg ConsensusMsg) (Uint256, error) {

	// FIXME: has to do marshal on each call

	data, err := SerializeVbftMsg(msg)
	if err != nil {
		return Uint256{}, fmt.Errorf("failed to marshal block: %s", err)
	}

	t := sha256.Sum256(data)
	f := sha256.Sum256(t[:])
	return Uint256(f), nil
}

type vrfData struct {
	BlockNum          uint64  `json:"block_num"`
	PrevBlockHash     Uint256 `json:"prev_block_hash"`
	PrevBlockProposer uint32  `json:"prev_block_proposer"` // TODO: change to NodeID
	PrevBlockSig      []byte  `json:"prev_block_sig"`
}

func vrf(block *Block, hash Uint256) VRFValue {

	// XXX: all-zero vrf value is taken as invalid

	sig := block.Block.Header.SigData[0]
	if block.isEmpty() {
		sig = block.Block.Header.SigData[1]
	}

	data, err := json.Marshal(&vrfData{
		BlockNum:          block.getBlockNum() + 1,
		PrevBlockHash:     hash,
		PrevBlockProposer: block.getProposer(),
		PrevBlockSig:      sig,
	})
	if err != nil {
		return VRFValue{}
	}

	t := sha512.Sum512(data)
	f := sha512.Sum512(t[:])
	return VRFValue(f)
}
