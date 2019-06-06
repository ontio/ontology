package TestCommon

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology-crypto/vrf"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/chainmgr"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
)

func CreateBlock(t *testing.T, lgr *ledger.Ledger, txs []*types.Transaction) *types.Block {
	txHash := []common.Uint256{}
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	lastBlock, _ := lgr.GetBlockByHeight(lgr.GetCurrentBlockHeight())
	if lastBlock == nil {
		t.Fatalf("nil chain of shard: %d", lgr.ShardID)
	}
	parentHeight := lastBlock.Header.ParentHeight
	txRoot := common.ComputeMerkleRoot(txHash)
	blockRoot := lgr.GetBlockRootWithNewTxRoots(lastBlock.Header.Height, []common.Uint256{lastBlock.Header.TransactionsRoot, txRoot})
	//shardTxs := xshard.GetCrossShardTxs()
	shardTxs := make(map[uint64][]*types.CrossShardTxInfos)
	consensusPayload := buildConsensusPayload(t, lastBlock)

	timestamp := uint32(time.Now().Unix())
	if timestamp <= lastBlock.Header.Timestamp {
		timestamp = lastBlock.Header.Timestamp + 1
	}

	blkHeader := &types.Header{
		PrevBlockHash:    lastBlock.Header.Hash(),
		Version:          common.CURR_HEADER_VERSION,
		ShardID:          lgr.ShardID.ToUint64(),
		ParentHeight:     uint32(parentHeight),
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        timestamp,
		Height:           lastBlock.Header.Height + 1,
		ConsensusData:    common.GetNonce(),
		ConsensusPayload: consensusPayload,
	}
	blk := &types.Block{
		Header:       blkHeader,
		ShardTxs:     shardTxs, // Cross-Shard Txs
		Transactions: txs,
	}
	blkHash := blk.Hash()
	acc := GetAccount(chainmgr.GetShardName(lgr.ShardID) + "_peerOwner0")
	if acc == nil {
		t.Fatalf("failed to get account peerOwner0")
	}
	sig, err := signature.Sign(acc, blkHash[:])
	if err != nil {
		t.Fatalf("sign block failed, block hash:%s, error: %s", blkHash.ToHexString(), err)
	}
	blkHeader.Bookkeepers = []keypair.PublicKey{acc.PublicKey}
	blkHeader.SigData = [][]byte{sig}

	return blk
}

func buildConsensusPayload(t *testing.T, prevBlk *types.Block) []byte {
	shardID := common.NewShardIDUnchecked(prevBlk.Header.ShardID)
	acc := GetAccount(chainmgr.GetShardName(shardID) + "_peerOwner0")
	if acc == nil {
		t.Fatalf("failed to get account peerOwner0")
	}

	lastBlkInfo, err := vconfig.VbftBlock(prevBlk.Header)
	if err != nil {
		t.Fatalf("get prev block vbft info: %s", err)
	}

	vrfValue, vrfProof, err := computeVrf(acc.PrivateKey, prevBlk.Header.Height+1, lastBlkInfo.VrfValue)
	if err != nil {
		t.Fatalf("failed to get vrf and proof: %s", err)
	}

	lastConfigBlkNum := lastBlkInfo.LastConfigBlockNum
	if lastBlkInfo.NewChainConfig != nil {
		lastConfigBlkNum = prevBlk.Header.Height
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           0,
		VrfValue:           vrfValue,
		VrfProof:           vrfProof,
		LastConfigBlockNum: lastConfigBlkNum,
		NewChainConfig:     nil,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		t.Fatalf("marshal vbft block info: %s", err)
	}
	return consensusPayload
}

type vrfData struct {
	BlockNum uint32 `json:"block_num"`
	PrevVrf  []byte `json:"prev_vrf"`
}

func computeVrf(sk keypair.PrivateKey, blkNum uint32, prevVrf []byte) ([]byte, []byte, error) {
	data, err := json.Marshal(&vrfData{
		BlockNum: blkNum,
		PrevVrf:  prevVrf,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("computeVrf failed to marshal vrfData: %s", err)
	}

	return vrf.Vrf(sk, data)
}
