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

package genesis

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/constants"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/service/neovm"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893
)

var (
	ONGXToken   = newUtilityToken()
	ONGXTokenID = ONGXToken.Hash()
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var INIT_PARAM = map[string]string{
	"gasPrice": "0",
}

var GenesisBookkeepers []keypair.PublicKey

// BuildGenesisBlock returns the genesis block with default consensus bookkeeper list
func BuildGenesisBlock(defaultBookkeeper []keypair.PublicKey, genesisConfig *config.GenesisConfig) (*types.Block, error) {
	//getBookkeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, fmt.Errorf("[Block],BuildGenesisBlock err with GetBookkeeperAddress: %s", err)
	}
	conf := bytes.NewBuffer(nil)
	if genesisConfig.VBFT != nil {
		genesisConfig.VBFT.Serialize(conf)
	}
	govConfig := newGoverConfigInit(conf.Bytes())
	consensusPayload, err := vconfig.GenesisConsensusPayload(govConfig.Hash(), 0)
	if err != nil {
		return nil, fmt.Errorf("consensus genesus init failed: %s", err)
	}
	//blockdata
	genesisHeader := &types.Header{
		Version:          BlockVersion,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        constants.GENESIS_BLOCK_TIMESTAMP,
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookkeeper:   nextBookkeeper,
		ConsensusPayload: consensusPayload,

		Bookkeepers: nil,
		SigData:     nil,
	}

	//block
	ongx := newUtilityToken()
	param := newParamContract()
	oid := deployOntIDContract()
	auth := deployAuthContract()
	config := newConfig()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			ongx,
			param,
			oid,
			auth,
			config,
			newParamInit(),
			govConfig,
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newUtilityToken() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.OngContractAddress[:], "ONG", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONG Token", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis utility token transaction error ")
	}
	return tx
}

func newParamContract() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.ParamContractAddress[:],
		"ParamConfig", "1.0", "Ontology Team", "contact@ont.io",
		"Chain Global Environment Variables Manager ", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis param transaction error ")
	}
	return tx
}

func newConfig() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.GovernanceContractAddress[:], "CONFIG", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Consensus Config", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis config transaction error ")
	}
	return tx
}

func deployAuthContract() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.AuthContractAddress[:], "AuthContract", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Authorization Contract", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis auth transaction error ")
	}
	return tx
}

func deployOntIDContract() *types.Transaction {
	mutable := utils.NewDeployTransaction(nutils.OntIDContractAddress[:], "OID", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONT ID", true)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis ontid transaction error ")
	}
	return tx
}

func newParamInit() *types.Transaction {
	params := new(global_params.Params)
	var s []string
	for k := range INIT_PARAM {
		s = append(s, k)
	}

	for k, v := range neovm.INIT_GAS_TABLE {
		INIT_PARAM[k] = strconv.FormatUint(v, 10)
		s = append(s, k)
	}

	sort.Strings(s)
	for _, v := range s {
		params.SetParam(global_params.Param{Key: v, Value: INIT_PARAM[v]})
	}
	//bf := new(bytes.Buffer)
	//params.Serialize(bf)
	//nutils.WriteAddress(bf, genAdminAddress())

	mutable := utils.BuildWasmNativeTransaction(nutils.ParamContractAddress, 0, global_params.INIT_NAME, params)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx

}

func genAdminAddress() common.Address {
	bookkeepers, _ := config.DefConfig.GetBookkeepers()
	var addr common.Address
	if len(bookkeepers) == 1 {
		addr = types.AddressFromPubKey(bookkeepers[0])
	} else {
		m := (5*len(bookkeepers) + 6) / 7
		temp, err := types.AddressFromMultiPubKeys(bookkeepers, m)
		if err != nil {
			panic(fmt.Sprint("wrong bookkeeper config, caused by", err))
		}
		addr = temp
	}
	return addr
}

func newGoverConfigInit(config []byte) *types.Transaction {
	mutable := utils.BuildWasmNativeTransaction(nutils.GovernanceContractAddress, 0, governance.INIT_CONFIG, config)
	tx, err := mutable.IntoImmutable()
	if err != nil {
		panic("constract genesis governing token transaction error ")
	}
	return tx
}
