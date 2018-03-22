package vbft

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Ontology/common"
	"github.com/Ontology/core/types"
	"github.com/Ontology/crypto"
)

type VbftBlockInfo struct {
	Proposer           uint32       `json:"leader"`
	LastConfigBlockNum uint64       `json:"last_config_block_num"`
	NewChainConfig     *ChainConfig `json:"new_chain_config"`
}

type Block struct {
	Block *types.Block   `json:"block"`
	Info  *VbftBlockInfo `json:"info"`
}

func (blk *Block) getProposer() uint32 {
	return blk.Info.Proposer
}

func (blk *Block) getBlockNum() uint64 {
	return uint64(blk.Block.Header.Height)
}

func (blk *Block) getPrevBlockHash() common.Uint256 {
	return blk.Block.Header.PrevBlockHash
}

func (blk *Block) getLastConfigBlockNum() uint64 {
	return blk.Info.LastConfigBlockNum
}

func (blk *Block) getNewChainConfig() *ChainConfig {
	return blk.Info.NewChainConfig
}

func (blk *Block) isEmpty() bool {
	return blk.Block.Transactions == nil || len(blk.Block.Transactions) == 0
}

func (blk *Block) Serialize() ([]byte, error) {
	infoData, err := json.Marshal(blk.Info)
	if err != nil {
		return nil, err
	}

	blk.Block.Header.ConsensusPayload = infoData

	buf := bytes.NewBuffer([]byte{})
	if err := blk.Block.Serialize(buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func initVbftBlock(block *types.Block) (*Block, error) {
	blkInfo := &VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return nil, err
	}

	return &Block{
		Block: block,
		Info:  blkInfo,
	}, nil
}

//////////////////////////////////////////////////////

const NodeIDBits = 534

// NodeID is a unique identifier for each node.
// The node identifier is a marshaled elliptic curve public key.
type NodeID [NodeIDBits / 8]byte

// Bytes returns a byte slice representation of the NodeID
func (n NodeID) Bytes() []byte {
	return n[:]
}

// NodeID prints as a long hexadecimal number.
func (n NodeID) String() string {
	return fmt.Sprintf("%x", n[:])
}

var NilID = NodeID{}

func (n NodeID) IsNil() bool {
	return bytes.Compare(n.Bytes(), NilID.Bytes()) == 0
}

func StringID(in string) (NodeID, error) {
	var id NodeID
	b, err := hex.DecodeString(strings.TrimPrefix(in, "0x"))
	if err != nil {
		return id, err
	} else if len(b) != len(id) {
		return id, fmt.Errorf("wrong length, want %d hex chars", len(id)*2)
	}
	copy(id[:], b)
	return id, nil
}

// PubkeyID returns a marshaled representation of the given public key.
func PubkeyID(pub *crypto.PubKey) (NodeID, error) {
	buf := bytes.NewBuffer([]byte(""))
	err := pub.Serialize(buf)
	if err != nil {
		return NilID, fmt.Errorf("serialize publickey: %s", err)
	}
	var id NodeID
	copy(id[:], buf.Bytes())
	return id, nil
}

func (id NodeID) Pubkey() (*crypto.PubKey, error) {
	buf := bytes.NewBuffer(id[:])
	pubKey := new(crypto.PubKey)
	err := pubKey.DeSerialize(buf)
	if err != nil {
		return nil, fmt.Errorf("deserialize failed: %s", err)
	}

	return pubKey, err
}
