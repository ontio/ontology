/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package cross_chain_manager

import (
	"encoding/hex"
	"encoding/json"
	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/service/native"
	cccom "github.com/ontio/ontology/smartcontract/service/native/cross_chain/common"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/header_sync"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	/**
		from boa.interop.System.Runtime import Notify

		def Main(operation, args):
	    	if operation == 'unlock':
				return unlock(args[0], args[1], args[2])
	    return False

		def unlock(params, fromContractAddr, fromChainId):
	    	return True
	*/
	contractCode = "51c56b6c58c56b6a00527ac46a51527ac46a52527ac46a51c306756e6c6f636b7d9c7c756424006a52c352c36a52c351c36a52c300c3536a00c3064c00000000006e6c7566620300006c756658c56b6a00527ac46a51527ac46a52527ac46a53527ac46a54527ac4620300516c7566"

	acct *account.Account

	setAcct = func() {
		acct = account.NewAccount("")
	}

	setBKers = func() {
		genesis.GenesisBookkeepers = []keypair.PublicKey{acct.PublicKey}
	}

	getHeaders = func() [][]byte {
		hdrs := make([][]byte, 0)

		blkInfo := &vconfig.VbftBlockInfo{
			NewChainConfig: &vconfig.ChainConfig{
				Peers: []*vconfig.PeerConfig{
					{Index: 0, ID: vconfig.PubkeyID(acct.PublicKey)},
				},
			},
		}
		payload, _ := json.Marshal(blkInfo)
		sr, _ := common.Uint256FromHexString("9eb1844022bfa7e04afa4887199c6c96a13d0b472fd922a186dfbc4e152f1a26")

		for i := uint32(0); i < 2; i++ {
			bd := &cccom.Header{
				Version:          0,
				Height:           i,
				ChainID:          4,
				Bookkeepers:      []keypair.PublicKey{acct.PublicKey},
				ConsensusPayload: payload,
				NextBookkeeper:   acct.Address,
				CrossStatesRoot:  sr,
			}

			hash := bd.Hash()
			sig, _ := signature.Sign(acct, hash[:])
			bd.SigData = [][]byte{sig}
			sink := common.NewZeroCopySink(nil)
			bd.Serialization(sink)
			hdrs = append(hdrs, sink.Bytes())
		}

		return hdrs
	}

	putHeaders = func(ns *native.NativeService) error {
		hdrs := getHeaders()
		sink := common.NewZeroCopySink(nil)
		p := &header_sync.SyncGenesisHeaderParam{
			GenesisHeader: hdrs[0],
		}
		p.Serialization(sink)
		ns.Input = sink.Bytes()
		_, err := header_sync.SyncGenesisHeader(ns)
		if err != nil {
			return err
		}

		sink = common.NewZeroCopySink(nil)
		param := &header_sync.SyncBlockHeaderParam{
			Address: acct.Address,
			Headers: [][]byte{hdrs[1]},
		}
		param.Serialization(sink)

		ns.Input = sink.Bytes()
		_, err = header_sync.SyncBlockHeader(ns)
		if err != nil {
			return err
		}

		return nil
	}

	getNativeFunc = func(args []byte) *native.NativeService {
		store, _ := leveldbstore.NewMemLevelDBStore()
		db := storage.NewCacheDB(overlaydb.NewOverlayDB(store))

		return &native.NativeService{
			CacheDB: db,
			Input:   args,
			Tx:      &types.Transaction{},
			ContextRef: &smartcontract.SmartContract{
				Contexts: []*context.Context{
					{ContractAddress: utils.CrossChainContractAddress},
					{},
				},
				Config: &smartcontract.Config{
					Tx: &types.Transaction{
						SignedAddr: []common.Address{acct.Address},
					},
					Height: 1,
				},
				Gas: 1000000000000000,
			},
		}
	}
)

func init() {
	setAcct()
	setBKers()
}

func TestCreateCrossChainTx(t *testing.T) {
	p := &CreateCrossChainTxParam{
		Method:            "test",
		Args:              []byte("test"),
		Fee:               1,
		ToContractAddress: []byte("btc"),
		ToChainID:         0,
	}
	sink := common.NewZeroCopySink(nil)
	p.Serialization(sink)

	ns := getNativeFunc(sink.Bytes())
	ok, err := CreateCrossChainTx(ns)
	assert.NoError(t, err)
	assert.Equal(t, ok, utils.BYTE_TRUE)
}

func TestProcessCrossChainTx(t *testing.T) {
	config.DefConfig.P2PNode.NetworkId = 3
	p := &ProcessCrossChainTxParam{
		Height:      1,
		Proof:       "9020ecd7c8289081775f69322505c0d5a19d3992d01b965eb819744014300966849c000000000000000020400d2249c5fc7435ecbaabf0db5dfc6c0c5004db222c6aa6a89bcc55a79c688b03627463020000000000000014a8a3f92b797ae1e059b8f4681ad949b6ce93771506756e6c6f636b1d14f3b8a17f1f957f60c88f105e32ebff3f022e56a400e1f50500000000",
		FromChainID: 4,
	}
	sink := common.NewZeroCopySink(nil)
	p.Serialization(sink)

	// put contract invoking into db
	ns := getNativeFunc(nil)
	code, _ := hex.DecodeString(contractCode)
	dc, _ := payload.NewDeployCode(code, payload.NEOVM_TYPE, "", "", "", "", "")
	ns.CacheDB.PutContract(dc)

	// put genesis and height 1 header
	err := putHeaders(ns)
	assert.NoError(t, err)
	ns.Input = sink.Bytes()

	// invoke func
	ok, err := ProcessCrossChainTx(ns)
	assert.NoError(t, err)
	assert.Equal(t, ok, utils.BYTE_TRUE)
}
