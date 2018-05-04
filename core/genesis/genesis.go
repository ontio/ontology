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
	"errors"
	"fmt"
	"time"

	"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	vconfig "github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/core/utils"
	"github.com/ontio/ontology/smartcontract/states"
	stypes "github.com/ontio/ontology/smartcontract/types"
)

const (
	BlockVersion uint32 = 0
	GenesisNonce uint64 = 2083236893
)

var (
	OntContractAddress, _   = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
	OngContractAddress, _   = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})
	OntIDContractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03})
	ParamContractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04})
	AuthContractAddress, _  = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06})
	GovernanceContractAddress, _ = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04})
	FeeSplitContractAddress, _   = common.AddressParseFromBytes([]byte{0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05})

	ONTToken   = newGoverningToken()
	ONGToken   = newUtilityToken()
	ONTTokenID = ONTToken.Hash()
	ONGTokenID = ONGToken.Hash()
)

var GenBlockTime = (config.DEFAULT_GEN_BLOCK_TIME * time.Second)

var GenesisBookkeepers []keypair.PublicKey

// GenesisBlockInit returns the genesis block with default consensus bookkeeper list
func GenesisBlockInit(defaultBookkeeper []keypair.PublicKey) (*types.Block, error) {
	//getBookkeeper
	GenesisBookkeepers = defaultBookkeeper
	nextBookkeeper, err := types.AddressFromBookkeepers(defaultBookkeeper)
	if err != nil {
		return nil, errors.New("[Block],GenesisBlockInit err with GetBookkeeperAddress")
	}
	consensusPayload, err := vconfig.GenesisConsensusPayload()
	if err != nil {
		return nil, fmt.Errorf("consensus genesus init failed: %s", err)
	}
	//blockdata
	genesisHeader := &types.Header{
		Version:          BlockVersion,
		PrevBlockHash:    common.Uint256{},
		TransactionsRoot: common.Uint256{},
		Timestamp:        uint32(uint32(time.Date(2017, time.February, 23, 0, 0, 0, 0, time.UTC).Unix())),
		Height:           uint32(0),
		ConsensusData:    GenesisNonce,
		NextBookkeeper:   nextBookkeeper,
		ConsensusPayload: consensusPayload,

		Bookkeepers: nil,
		SigData:     nil,
	}

	//block
	ont := newGoverningToken()
	ong := newUtilityToken()
	param := newParamContract()
	oid := deployOntIDContract()
	auth := deployAuthContract()
	config := newConfig()

	genesisBlock := &types.Block{
		Header: genesisHeader,
		Transactions: []*types.Transaction{
			ont,
			ong,
			param,
			oid,
			auth,
			config,
			newGoverningInit(),
			newUtilityInit(),
			newParamInit(),
		},
	}
	genesisBlock.RebuildMerkleRoot()
	return genesisBlock, nil
}

func newGoverningToken() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: OntContractAddress[:], VmType: stypes.Native}, "ONT", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONT Token", true)
	return tx
}

func newUtilityToken() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: OngContractAddress[:], VmType: stypes.Native}, "ONG", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONG Token", true)
	return tx
}

func newParamContract() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: ParamContractAddress[:], VmType: stypes.Native},
		"ParamConfig", "1.0", "Ontology Team", "contact@ont.io",
		"Chain Global Enviroment Variables Manager ", true)
func newConfig() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: GovernanceContractAddress[:], VmType: stypes.Native}, "CONFIG", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Consensus Config", true)
	return tx
}

func deployAuthContract() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: AuthContractAddress[:], VmType: stypes.Native}, "AuthContract", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network Authorization Contract", true)
	return tx
}

func deployOntIDContract() *types.Transaction {
	tx := utils.NewDeployTransaction(stypes.VmCode{Code: OntIDContractAddress[:], VmType: stypes.Native}, "OID", "1.0",
		"Ontology Team", "contact@ont.io", "Ontology Network ONT ID", true)
	return tx
}

func newGoverningInit() *types.Transaction {
	init := states.Contract{
		Address: OntContractAddress,
		Method:  "init",
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := stypes.VmCode{
		VmType: stypes.Native,
		Code:   bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}

func newUtilityInit() *types.Transaction {
	init := states.Contract{
		Address: OngContractAddress,
		Method:  "init",
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := stypes.VmCode{
		VmType: stypes.Native,
		Code:   bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}

func newParamInit() *types.Transaction {
	init := states.Contract{
		Address: ParamContractAddress,
		Method:  "init",
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := stypes.VmCode{
		VmType: stypes.Native,
		Code:   bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}

func newConfigInit() *types.Transaction {
	init := states.Contract{
		Address: GovernanceContractAddress,
		Method:  "initConfig",
	}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	vmCode := stypes.VmCode{
		VmType: stypes.Native,
		Code:   bf.Bytes(),
	}
	tx := utils.NewInvokeTransaction(vmCode)
	return tx
}
