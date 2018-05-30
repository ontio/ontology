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

package init

import (
	"bytes"

	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/auth"
	params "github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/ong"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/ontio/ontology/smartcontract/service/native/ontid"
	"github.com/ontio/ontology/smartcontract/service/native/utils"
	"github.com/ontio/ontology/smartcontract/states"
)

var (
	ONT_INIT_BYTES    = InitBytes(utils.OntContractAddress, ont.INIT_NAME)
	ONG_INIT_BYTES    = InitBytes(utils.OngContractAddress, ont.INIT_NAME)
	PARAM_INIT_BYTES  = InitBytes(utils.ParamContractAddress, params.INIT_NAME)
	COMMIT_DPOS_BYTES = InitBytes(utils.GovernanceContractAddress, governance.COMMIT_DPOS)
	INIT_CONFIG_BYTES = InitBytes(utils.GovernanceContractAddress, governance.INIT_CONFIG)
)

func init() {
	ong.InitOng()
	ont.InitOnt()
	params.InitGlobalParams()
	ontid.Init()
	auth.Init()
	governance.InitGovernance()
}

func InitBytes(addr common.Address, method string) []byte {
	init := states.Contract{Address: addr, Method: method}
	bf := new(bytes.Buffer)
	init.Serialize(bf)
	return bf.Bytes()
}
