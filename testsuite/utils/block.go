package utils

import (
	"bytes"
	"encoding/hex"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/shardmgmt"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	TestCommon "github.com/ontio/ontology/testsuite/common"
	"testing"
)

func GenInitShardAssetBlock(t *testing.T) *types.Block {
	rootShard := common.NewShardIDUnchecked(config.DEFAULT_SHARD_ID)
	// transfer ont to shard creator and peer owner
	transferMultiParam := make([]*ont.State, 0)
	pubkeys := []keypair.PublicKey{}
	to := []common.Address{}
	for i := 0; i < 7; i++ {
		ownerName := TestCommon.GetOwnerName(rootShard, i)
		owner := TestCommon.GetAccount(ownerName)
		pubkeys = append(pubkeys, owner.PublicKey)
		to = append(to, owner.Address)
	}
	from, err := types.AddressFromMultiPubKeys(pubkeys, 5)
	if err != nil {
		t.Fatalf("generate multi sign addr failed, err: %s", err)
	}
	shardCreator := TestCommon.GetAccount(TestCommon.GetUserName(rootShard, 1))
	to = append(to, shardCreator.Address)
	for _, addr := range to {
		transferMultiParam = append(transferMultiParam, &ont.State{
			From:  from,
			To:    addr,
			Value: 1000000,
		})
	}
	transferMultiParam = append(transferMultiParam, &ont.State{From: from,
		To:    shardCreator.Address,
		Value: 1000000,
	})
	transferOntTx := TestCommon.CreateAdminTx(t, rootShard, 0, utils.OntContractAddress, ont.TRANSFER_NAME,
		[]interface{}{transferMultiParam})
	result := []*types.Transaction{transferOntTx}
	for _, addr := range to {
		transferFromParam := &ont.TransferFrom{
			Sender: from,
			From:   utils.OntContractAddress,
			To:     addr,
			Value:  1000000000000, // 1000 ong
		}
		withdrawOngTx := TestCommon.CreateAdminTx(t, rootShard, 0, utils.OngContractAddress, ont.TRANSFERFROM_NAME,
			[]interface{}{transferFromParam})
		result = append(result, withdrawOngTx)
	}
	return TestCommon.CreateBlock(t, ledger.GetShardLedger(rootShard), result)
}

func GenRunShardBlock(t *testing.T, shardID, childShard common.ShardID, creatorName string) *types.Block {
	creator := TestCommon.GetAccount(creatorName)
	createParam := &shardmgmt.CreateShardParam{ParentShardID: shardID, Creator: creator.Address}
	createShardTx := TestCommon.CreateNativeTx(t, creatorName, 0, utils.ShardMgmtContractAddress,
		shardmgmt.CREATE_SHARD_NAME, []interface{}{createParam})
	vbftCfg := TestCommon.GetConfig(t, childShard).Genesis.VBFT
	vbftCfg.Peers = []*config.VBFTPeerStakeInfo{}
	cfgBuff := new(bytes.Buffer)
	if err := vbftCfg.Serialize(cfgBuff); err != nil {
		t.Fatalf("serialize vbft config failed, err: %s", err)
	}
	configParam := &shardmgmt.ConfigShardParam{
		ShardID:           childShard,
		NetworkMin:        7,
		GasPrice:          0,
		GasLimit:          20000,
		StakeAssetAddress: utils.OntContractAddress,
		GasAssetAddress:   utils.OngContractAddress,
		VbftConfigData:    cfgBuff.Bytes(),
	}
	configTx := TestCommon.CreateNativeTx(t, creatorName, 0, utils.ShardMgmtContractAddress,
		shardmgmt.CONFIG_SHARD_NAME, []interface{}{configParam})

	runShardTxs := []*types.Transaction{createShardTx, configTx}
	pubkeys := make([]string, 0)
	for i := 0; i < 7; i++ {
		peerName := TestCommon.GetOwnerName(shardID, i)
		peer := TestCommon.GetAccount(peerName)
		pubkey := hex.EncodeToString(keypair.SerializePublicKey(peer.PublicKey))
		pubkeys = append(pubkeys, pubkey)
		applyJoinParam := &shardmgmt.ApplyJoinShardParam{
			ShardId:    childShard,
			PeerOwner:  peer.Address,
			PeerPubKey: pubkey,
		}
		applyJoinTx := TestCommon.CreateNativeTx(t, peerName, 0, utils.ShardMgmtContractAddress,
			shardmgmt.APPLY_JOIN_SHARD_NAME, []interface{}{applyJoinParam})
		runShardTxs = append(runShardTxs, applyJoinTx)
	}

	approveJoinParam := shardmgmt.ApproveJoinShardParam{ShardId: childShard, PeerPubKey: pubkeys}
	approveJoinTx := TestCommon.CreateNativeTx(t, creatorName, 0, utils.ShardMgmtContractAddress,
		shardmgmt.APPROVE_JOIN_SHARD_NAME, []interface{}{approveJoinParam})
	runShardTxs = append(runShardTxs, approveJoinTx)

	for i := 0; i < 7; i++ {
		peerName := TestCommon.GetOwnerName(shardID, i)
		peer := TestCommon.GetAccount(peerName)
		joinParam := &shardmgmt.JoinShardParam{
			ShardID:     childShard,
			IpAddress:   "http://localhost:30336",
			PeerOwner:   peer.Address,
			PeerPubKey:  pubkeys[i],
			StakeAmount: uint64(vbftCfg.MinInitStake),
		}
		joinTx := TestCommon.CreateNativeTx(t, peerName, 0, utils.ShardMgmtContractAddress,
			shardmgmt.JOIN_SHARD_NAME, []interface{}{joinParam})
		runShardTxs = append(runShardTxs, joinTx)
	}

	activateParam := &shardmgmt.ActivateShardParam{ShardID: childShard}
	activateTx := TestCommon.CreateNativeTx(t, creatorName, 0, utils.ShardMgmtContractAddress,
		shardmgmt.ACTIVATE_SHARD_NAME, []interface{}{activateParam})
	runShardTxs = append(runShardTxs, activateTx)
	return TestCommon.CreateBlock(t, ledger.GetShardLedger(shardID), runShardTxs)
}
