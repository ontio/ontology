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
	"github.com/ontio/ontology/smartcontract/service/native/auth"
	params "github.com/ontio/ontology/smartcontract/service/native/global_params"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
	"github.com/ontio/ontology/smartcontract/service/native/ongx"
	"github.com/ontio/ontology/smartcontract/service/native/ontid"
	//vm "github.com/ontio/ontology/vm/neovm"
	"fmt"
)

//var (
//	COMMIT_DPOS_BYTES = InitBytes(utils.GovernanceContractAddress, governance.COMMIT_DPOS)
//)

func init() {
	fmt.Println("------init----")
	ongx.InitOngx()
	params.InitGlobalParams()
	ontid.Init()
	auth.Init()
	governance.InitGovernance()
}

//func InitBytes(addr common.Address, method string) []byte {
//	bf := new(bytes.Buffer)
//	builder := vm.NewParamsBuilder(bf)
//	builder.EmitPushByteArray([]byte{})
//	builder.EmitPushByteArray([]byte(method))
//	builder.EmitPushByteArray(addr[:])
//	builder.EmitPushInteger(big.NewInt(0))
//	builder.Emit(vm.SYSCALL)
//	builder.EmitPushByteArray([]byte(neovm.NATIVE_INVOKE_NAME))
//
//	return builder.ToArray()
//}
