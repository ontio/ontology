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

package test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/states"
	"github.com/ontio/ontology/core/store/leveldbstore"
	"github.com/ontio/ontology/core/store/overlaydb"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract"
	"github.com/ontio/ontology/smartcontract/context"
	"github.com/ontio/ontology/smartcontract/service/native"
	ccom "github.com/ontio/ontology/smartcontract/service/native/cross_chain/common"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/cross_chain_manager"
	"github.com/ontio/ontology/smartcontract/service/native/cross_chain/header_sync"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/storage"
	"github.com/stretchr/testify/assert"
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
		sr, _ := common.Uint256FromHexString("61e28237109bc99e53981bf9c4d9326bc764fda9ad1d1f95d7a5d2eb25ec01db")

		for i := uint32(0); i < 2; i++ {
			bd := &ccom.Header{
				Version:          0,
				Height:           i,
				ChainID:          4,
				Bookkeepers:      []keypair.PublicKey{acct.PublicKey},
				ConsensusPayload: payload,
				NextBookkeeper:   acct.Address,
				CrossStateRoot:   sr,
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
}

func TestCreateCrossChainTx(t *testing.T) {
	p := &cross_chain_manager.CreateCrossChainTxParam{
		Method:            "test",
		Args:              []byte("test"),
		ToContractAddress: []byte("btc"),
		ToChainID:         0,
	}
	sink := common.NewZeroCopySink(nil)
	p.Serialization(sink)

	ns := getNativeFunc(sink.Bytes())
	ok, err := cross_chain_manager.CreateCrossChainTx(ns)
	assert.NoError(t, err)
	assert.Equal(t, ok, utils.BYTE_TRUE)
}

func TestProcessCrossChainTx(t *testing.T) {
	config.DefConfig.P2PNode.NetworkId = 3
	p := &cross_chain_manager.ProcessCrossChainTxParam{
		Height:      1,
		Proof:       "b1203e52461bd03325f183d9ed47261d39d24a081b307742545bcc68bfb0aa56d528010000000000000020afe92a7d08e5e8e64a72fbb3da5f51fa07483eb01e18ca6a2c15a5c0fc45120420afe92a7d08e5e8e64a72fbb3da5f51fa07483eb01e18ca6a2c15a5c0fc45120403627463030000000000000014a8a3f92b797ae1e059b8f4681ad949b6ce93771506756e6c6f636b1d14f3b8a17f1f957f60c88f105e32ebff3f022e56a400e1f50500000000",
		FromChainID: 4,
	}
	sink := common.NewZeroCopySink(nil)
	p.Serialization(sink)

	// put contract invoking into db
	bf := common.NewZeroCopySink(nil)
	utils.EncodeAddress(bf, acct.Address)
	si := &states.StorageItem{Value: bf.Bytes()}

	ns := getNativeFunc(sink.Bytes())
	ns.CacheDB.Put(global_params.GenerateOperatorKey(utils.ParamContractAddress), si.ToArray())

	code, err := hex.DecodeString(contractCode)
	if err != nil {
		t.Fatal(err)
	}
	dc, err := payload.NewDeployCode(code, payload.NEOVM_TYPE, "", "", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	ns.CacheDB.PutContract(dc)

	// put genesis and height 1 header
	err = putHeaders(ns)
	assert.NoError(t, err)
	ns.Input = sink.Bytes()

	// invoke func
	ok, err := cross_chain_manager.ProcessCrossChainTx(ns)
	assert.NoError(t, err)
	assert.Equal(t, ok, utils.BYTE_TRUE)
}

func TestHeaderDeserialize(t *testing.T) {
	headerString := "000000000400000000000000727524826f5b3058fe69cabc86c4c305c22e0cc963125db5facdef82c0bef9ea0000000000000000000000000000000000000000000000000000000000000000641b7081aa1006abb9547611208e8d8e1af6ebb9642754031fd1b0c53026413c8d7da518e51fe6756dc8a7d709798b228656b573e9537bb1cf13a609d2e05f8d6a9d215e2a00000079299c538ab22a0afd0c017b226c6561646572223a362c227672665f76616c7565223a224241714b546b73764d50386262786c4a4870525a384d787669676c384974304f4b62586775637134323866513835766579786946516738416c75707543424e594e47517638684143654161524b38794a4e73725a794b383d222c227672665f70726f6f66223a226e71544742744a6a73477847547a65767a3153692b5278357a387737334176744762757a322b4479556a444a2f735955694b754d6b73664c476c652f7452524f343545486437585a6f413633435368666d4d49676b673d3d222c226c6173745f636f6e6669675f626c6f636b5f6e756d223a302c226e65775f636861696e5f636f6e666967223a6e756c6c7d00000000000000000000000000000000000000000623120503a6b1b0e1c6977f44f36232a5f3b61b6a859b4ce5147c49accfa9a42f483f120423120502ff7dfc705bb5ac68d2e89230c662999aeb18284131e9fce94ef9faf5b991753d23120503c6e681e5154efba64c75b0aa1ca54489bae7653077d16ddd9726f365be30622d2312050333c44837ddb94ad5fa0eb460b0f49154f90503619d4d2c8ee08330fb815841d22312050344017ccca820d90f0ebb461df46337b0923b0ae2bce583cee1a2624c9208e20823120503d101838807ec4079a46fe98d6bd9a0690abcbd8ce16e0fbc4520c7c7ef7885db0642051be164fbc41063d65b516ef61cba1193b131f4e17eed81c4d77cafa6d9e5392de64e35ca5fe83af0eaa87c05916b2bc68cfc897b4e13b33b4873c029512a8262c942051b30bad1ab0401a7d0386c038e57838d5e1d311282f5b20db64a7ab96ad127b7e442126b51265c780fa5b0c570ec9ed10a3a20180c8b0dfaaa7367ac9416738c0d42051b21a29807e3fd5d5861d7d928e518def38a10e12f6a27483bd50cecc6d17753000f4340665b07e2f97bcc75e8932a9cc6087fd99350a24bd69bd5c3c5ce5c065d42051ccd4f97587d254433df1085baf3e210dbe180f3ba485c6faff645346a213fb5b1736fcbb6b5cb05e9b2490d1ae4230c2c29a54321d4a7c09d8ea174dd62ba568842051c61172daf05e96a8092d9330fb17c5dc079f36722284f65ffa336dbc98da3366c6dbf707f1d6fb4d094ac2a38da4a20fd3755b52f157819ce94b92d5ac725b6b742051c892b966fdede7751dacdd09853dfc6d286663833be2e2c655cddc8072f44075a46ec7e467fea272c81c5a61d605c76a677cdf16992b8b5b9503ee1dabca46b1c"
	header, err := hex.DecodeString(headerString)
	assert.NoError(t, err)
	header2 := new(ccom.Header)
	err = header2.Deserialization(common.NewZeroCopySource(header))
	assert.NoError(t, err)
}
